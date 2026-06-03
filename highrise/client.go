package highrise

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var keepaliveInterval = 15 * time.Second

const (
	defaultURL           = "wss://highrise.game/web/botapi"
	sdkVersion           = "0.1.0"
	readTimeout          = 20 * time.Second
	reconnectBaseDelay   = 1 * time.Second
	reconnectMaxDelay    = 30 * time.Second
	responseQueueTimeout = 30 * time.Second
	maxEventWorkers      = 64
)

type reqRIDer interface {
	getRID() string
}

type Client struct {
	conn           *websocket.Conn
	connMu         sync.Mutex
	url            string
	roomID         string
	apiToken       string
	reqID          atomic.Int64
	pendingResp    sync.Map
	handler        BotHandler
	highrise       *Highrise
	stopCh         chan struct{}
	stopOnce       sync.Once
	wg             sync.WaitGroup
	reconnectDelay time.Duration
	sdkVersion     string

	eventSem    chan struct{}
	rateLimiter *rateLimiter
}

func NewClient(handler BotHandler) *Client {
	c := &Client{
		url:         defaultURL,
		handler:     handler,
		stopCh:      make(chan struct{}),
		sdkVersion:  sdkVersion,
		eventSem:    make(chan struct{}, maxEventWorkers),
		rateLimiter: newRateLimiter(),
	}
	c.highrise = newHighrise(c)
	return c
}

func extractType(req any) string {
	t := reflect.TypeOf(req)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func (c *Client) nextReqID() int64 {
	return c.reqID.Add(1)
}

func (c *Client) subscribeEvents() string {
	return "chat,emote,reaction,user_joined,user_left,user_moved,tip_reaction,voice,channel,message,moderation"
}

func (c *Client) Run(ctx context.Context, roomID, apiToken string) error {
	c.roomID = roomID
	c.apiToken = apiToken

	if bs, ok := c.handler.(HasBeforeStart); ok {
		bs.BeforeStart(ctx)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.stopCh:
			return nil
		default:
		}

		if err := c.connect(ctx); err != nil {
			log.Printf("Connection failed: %v, reconnecting...", err)
			c.backoffSleep(ctx)
			continue
		}

		c.reconnectDelay = 0

		c.wg.Add(2)
		go c.readLoop(ctx)
		go c.keepaliveLoop(ctx)
		c.wg.Wait()

		select {
		case <-c.stopCh:
			return nil
		default:
			log.Println("Disconnected, reconnecting...")
			c.backoffSleep(ctx)
		}
	}
}

func (c *Client) backoffSleep(ctx context.Context) {
	if c.reconnectDelay == 0 {
		c.reconnectDelay = reconnectBaseDelay
	} else {
		c.reconnectDelay *= 2
		if c.reconnectDelay > reconnectMaxDelay {
			c.reconnectDelay = reconnectMaxDelay
		}
	}

	jitter := time.Duration(rand.Int63n(int64(c.reconnectDelay / 4)))
	delay := c.reconnectDelay + jitter

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func (c *Client) connect(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	events := c.subscribeEvents()
	urlStr := c.url
	if events != "" {
		urlStr = c.url + "?events=" + events
	}

	header := http.Header{}
	header.Set("room-id", c.roomID)
	header.Set("api-token", c.apiToken)
	header.Set("user-agent", "highrise-go-bot-sdk/"+c.sdkVersion)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, urlStr, header)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	conn.SetReadLimit(1 << 20)
	c.conn = conn
	return nil
}

func (c *Client) readLoop(ctx context.Context) {
	defer c.wg.Done()
	defer func() {
		c.connMu.Lock()
		if c.conn != nil {
			c.conn.Close()
		}
		c.connMu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		default:
		}

		c.connMu.Lock()
		conn := c.conn
		c.connMu.Unlock()

		if conn == nil {
			return
		}

		conn.SetReadDeadline(time.Now().Add(readTimeout))

		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			return
		}

		c.handleMessage(ctx, message)
	}
}

func (c *Client) handleMessage(ctx context.Context, data []byte) {
	var envelope struct {
		Type string  `json:"_type"`
		RID  *string `json:"rid,omitempty"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		log.Printf("Failed to decode envelope: %v", err)
		return
	}

	if envelope.RID != nil {
		if ch, ok := c.pendingResp.Load(*envelope.RID); ok {
			respCh, ok := ch.(chan []byte)
			if ok {
				select {
				case respCh <- data:
				default:
				}
			}
			return
		}
	}

	if envelope.RID == nil {
		c.routeEvent(ctx, envelope.Type, data)
	}
}

func (c *Client) dispatchEvent(fn func()) {
	select {
	case c.eventSem <- struct{}{}:
		go func() {
			defer func() { <-c.eventSem }()
			fn()
		}()
	default:
		log.Println("Event worker pool full, dropping event")
	}
}

func (c *Client) routeEvent(ctx context.Context, msgType string, data []byte) {
	switch msgType {
	case "SessionMetadata":
		var meta SessionMetadata
		if err := json.Unmarshal(data, &meta); err != nil {
			log.Printf("Failed to decode SessionMetadata: %v", err)
			return
		}
		c.highrise.myID = meta.UserID
		c.rateLimiter.apply(meta.RateLimits)
		if h, ok := c.handler.(HasOnStart); ok {
			c.dispatchEvent(func() { h.OnStart(ctx, &meta) })
		}

	case "ChatEvent":
		var ev ChatEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			log.Printf("Failed to decode ChatEvent: %v", err)
			return
		}
		if ev.Whisper {
			if hw, ok := c.handler.(HasOnWhisper); ok {
				c.dispatchEvent(func() { hw.OnWhisper(ctx, ev.User, ev.Message) })
			}
		} else if h, ok := c.handler.(HasOnChat); ok {
			c.dispatchEvent(func() { h.OnChat(ctx, ev.User, ev.Message) })
		}

	case "EmoteEvent":
		var ev EmoteEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			log.Printf("Failed to decode EmoteEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnEmote); ok {
			c.dispatchEvent(func() { h.OnEmote(ctx, ev.User, ev.EmoteID, ev.Receiver) })
		}

	case "ReactionEvent":
		var ev ReactionEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			log.Printf("Failed to decode ReactionEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnReaction); ok {
			c.dispatchEvent(func() { h.OnReaction(ctx, ev.User, ev.Reaction, ev.Receiver) })
		}

	case "UserJoinedEvent":
		var ev UserJoinedEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			log.Printf("Failed to decode UserJoinedEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnUserJoin); ok {
			c.dispatchEvent(func() { h.OnUserJoin(ctx, ev.User, ev.Position) })
		}

	case "UserLeftEvent":
		var ev UserLeftEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			log.Printf("Failed to decode UserLeftEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnUserLeave); ok {
			c.dispatchEvent(func() { h.OnUserLeave(ctx, ev.User) })
		}

	case "UserMovedEvent":
		var ev UserMovedEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			log.Printf("Failed to decode UserMovedEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnUserMove); ok {
			c.dispatchEvent(func() { h.OnUserMove(ctx, ev.User, ev.Position) })
		}

	case "TipReactionEvent":
		var ev TipReactionEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			log.Printf("Failed to decode TipReactionEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnTip); ok {
			evCopy := ev
			c.dispatchEvent(func() { h.OnTip(ctx, evCopy.Sender, evCopy.Receiver, &evCopy.Item) })
		}

	case "VoiceEvent":
		var ev VoiceEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			log.Printf("Failed to decode VoiceEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnVoiceChange); ok {
			c.dispatchEvent(func() { h.OnVoiceChange(ctx, ev.Users, ev.SecondsLeft) })
		}

	case "ChannelEvent":
		var ev ChannelEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			log.Printf("Failed to decode ChannelEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnChannel); ok {
			c.dispatchEvent(func() { h.OnChannel(ctx, ev.SenderID, ev.Message, ev.Tags) })
		}

	case "MessageEvent":
		var ev MessageEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			log.Printf("Failed to decode MessageEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnMessage); ok {
			c.dispatchEvent(func() { h.OnMessage(ctx, ev.UserID, ev.ConversationID, ev.IsNewConversation) })
		}

	case "RoomModeratedEvent":
		var ev RoomModeratedEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			log.Printf("Failed to decode RoomModeratedEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnModerate); ok {
			c.dispatchEvent(func() { h.OnModerate(ctx, ev.ModeratorID, ev.TargetUserID, ev.ModerationType, ev.Duration) })
		}

	case "Error":
		var errMsg Error
		if err := json.Unmarshal(data, &errMsg); err != nil {
			log.Printf("Failed to decode Error: %v", err)
			return
		}
		log.Printf("Server error: %s", errMsg.Message)
		if h, ok := c.handler.(HasOnError); ok {
			c.dispatchEvent(func() { h.OnError(ctx, errMsg) })
		}
		if errMsg.DoNotReconnect {
			log.Println("Server instructed not to reconnect, stopping...")
			c.Stop()
		}
	}
}

func (c *Client) keepaliveLoop(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(keepaliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			msg := KeepaliveRequest{
				Type: "KeepaliveRequest",
			}
			if err := c.writeJSON(msg); err != nil {
				log.Printf("Keepalive error: %v", err)
				return
			}
		}
	}
}

func (c *Client) writeJSON(v any) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *Client) sendRequest(ctx context.Context, req any) ([]byte, error) {
	ridder, ok := req.(reqRIDer)
	if !ok {
		return nil, fmt.Errorf("request type %T does not implement reqRIDer", req)
	}
	rid := ridder.getRID()

	if t := extractType(req); t != "" {
		if err := c.rateLimiter.acquire(ctx, t); err != nil {
			return nil, err
		}
	}

	ch := make(chan []byte, 1)
	c.pendingResp.Store(rid, ch)
	defer c.pendingResp.Delete(rid)

	if err := c.writeJSON(req); err != nil {
		return nil, &ConnectionError{Err: err}
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-ch:
		var errMsg struct {
			Type    string `json:"_type"`
			Message string `json:"message"`
		}
		if json.Unmarshal(resp, &errMsg) == nil && errMsg.Type == "Error" {
			return nil, NewResponseError(errMsg.Message)
		}
		return resp, nil
	case <-time.After(responseQueueTimeout):
		return nil, NewResponseError("request timed out")
	}
}

func (c *Client) Highrise() *Highrise {
	return c.highrise
}

func (c *Client) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
}
