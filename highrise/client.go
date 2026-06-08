// Package highrise provides a Go SDK for the Highrise Bot API.
// It includes a WebSocket client for real-time bot actions (chat, walk, teleport, etc.),
// event handler interfaces, rate limiting, and a REST client for the Highrise WebAPI.
//
// Basic usage:
//
//	type MyBot struct {
//	    highrise.Bot
//	}
//
//	func (b *MyBot) OnChat(ctx context.Context, user highrise.User, message string) {
//	    b.Highrise.Chat(ctx, "Hello!")
//	}
//
//	bot := &MyBot{}
//	client := highrise.NewClient(bot)
//	bot.SetHighrise(client.Highrise())
//	client.Run(ctx, roomID, apiToken)
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
	defaultURL            = "wss://highrise.game/web/botapi"
	defaultActionTimeout  = 10 * time.Second
	sdkVersion            = "0.1.0"
	readTimeout           = 20 * time.Second
	reconnectBaseDelay    = 1 * time.Second
	reconnectMaxDelay     = 30 * time.Second
	responseQueueTimeout  = 30 * time.Second
	maxEventWorkers       = 64
	defaultWriteDeadline  = 10 * time.Second
)

type reqRIDer interface {
	getRID() string
}

// Logger is the interface for client logging
type Logger interface {
	Printf(format string, v ...any)
}

// ConnectionState represents the state of the WebSocket connection
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
)

func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateReconnecting:
		return "reconnecting"
	default:
		return "unknown"
	}
}

// Middleware wraps an event handler call. The next function is the actual
// event handler; middleware can run logic before and/or after it.
type Middleware func(next func())

// ClientOption is a functional option for configuring a Client
type ClientOption func(*Client)

// WithURL sets a custom WebSocket URL for the client
func WithURL(url string) ClientOption {
	return func(c *Client) {
		c.url = url
	}
}

// WithLogger sets a custom logger for the client
func WithLogger(l Logger) ClientOption {
	return func(c *Client) {
		if l != nil {
			c.log = l
		}
	}
}

// WithSDKVersion sets a custom SDK version string
func WithSDKVersion(version string) ClientOption {
	return func(c *Client) {
		c.sdkVersion = version
	}
}

// WithEvents sets a custom event subscription list.
// Default: "chat,emote,reaction,user_joined,user_left,user_moved,tip_reaction,voice,channel,message,moderation"
func WithEvents(events string) ClientOption {
	return func(c *Client) {
		c.eventFilter = events
	}
}

// WithActionTimeout sets a default context timeout for every action call.
// Default: 10s. Set to 0 to disable (use caller's context as-is).
func WithActionTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.actionTimeout = timeout
	}
}

// Client is a WebSocket client for the Highrise Bot API.
// It manages the connection, event routing, rate limiting, and reconnection logic.
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

	eventSem      chan struct{}
	rateLimiter   *rateLimiter
	log           Logger
	state         ConnectionState
	middlewares   []Middleware
	eventFilter   string
	actionTimeout time.Duration

	totalEvents atomic.Int64
	totalActions atomic.Int64
	totalErrors atomic.Int64
	totalReconnects atomic.Int64
}

// NewClient creates a new Client with the given BotHandler and options
func NewClient(handler BotHandler, opts ...ClientOption) *Client {
	c := &Client{
		url:           defaultURL,
		handler:       handler,
		stopCh:        make(chan struct{}),
		sdkVersion:    sdkVersion,
		eventSem:      make(chan struct{}, maxEventWorkers),
		rateLimiter:   newRateLimiter(),
		log:           log.Default(),
		actionTimeout: defaultActionTimeout,
	}
	for _, opt := range opts {
		opt(c)
	}
	c.highrise = newHighrise(c)
	return c
}

// Use adds a middleware that wraps every event handler call.
// Middleware are executed in the order they are added.
func (c *Client) Use(mw Middleware) {
	c.middlewares = append(c.middlewares, mw)
}

// IsConnected returns true if the client is currently connected
func (c *Client) IsConnected() bool {
	return c.state == StateConnected
}

// IsStopped returns true if Stop has been called
func (c *Client) IsStopped() bool {
	select {
	case <-c.stopCh:
		return true
	default:
		return false
	}
}

// Metrics returns a snapshot of internal counters
func (c *Client) Metrics() map[string]int64 {
	return map[string]int64{
		"events":    c.totalEvents.Load(),
		"actions":   c.totalActions.Load(),
		"errors":    c.totalErrors.Load(),
		"reconnects": c.totalReconnects.Load(),
	}
}

func (c *Client) setState(state ConnectionState) {
	c.state = state
	if h, ok := c.handler.(HasOnConnectionChange); ok {
		c.dispatchEvent(func() { h.OnConnectionChange(context.Background(), state) })
	}
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
	if c.eventFilter != "" {
		return c.eventFilter
	}
	return "chat,emote,reaction,user_joined,user_left,user_moved,tip_reaction,voice,channel,message,moderation"
}

// Run connects to the room and starts processing events.
// It blocks until the context is cancelled or Stop is called, handling reconnections.
func (c *Client) Run(ctx context.Context, roomID, apiToken string) error {
	c.roomID = roomID
	c.apiToken = apiToken
	c.totalReconnects.Store(0)

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

		c.setState(StateConnecting)
		if err := c.connect(ctx); err != nil {
			c.log.Printf("Connection failed: %v, reconnecting...", err)
			c.backoffSleep(ctx)
			c.totalReconnects.Add(1)
			continue
		}
		c.setState(StateConnected)

		c.reconnectDelay = 0

		c.wg.Add(2)
		go c.readLoop(ctx)
		go c.keepaliveLoop(ctx)
		c.wg.Wait()

		select {
		case <-c.stopCh:
			c.setState(StateDisconnected)
			return nil
		default:
			c.log.Printf("Disconnected, reconnecting...")
			c.totalReconnects.Add(1)
			c.setState(StateReconnecting)
			c.backoffSleep(ctx)
			c.setState(StateDisconnected)
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
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: true,
	}

	conn, _, err := dialer.DialContext(ctx, urlStr, header)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	conn.SetReadLimit(1 << 20)
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})
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
			c.log.Printf("Read error: %v", err)
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
		c.log.Printf("Failed to decode envelope: %v", err)
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
		c.totalEvents.Add(1)
		c.routeEvent(ctx, envelope.Type, data)
	}
}

func (c *Client) dispatchEvent(fn func()) {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		mw := c.middlewares[i]
		prev := fn
		fn = func() { mw(prev) }
	}

	select {
	case c.eventSem <- struct{}{}:
		go func() {
			defer func() {
				<-c.eventSem
				if r := recover(); r != nil {
					c.log.Printf("PANIC in event handler: %v", r)
				}
			}()
			fn()
		}()
	default:
		c.log.Printf("Event worker pool full, dropping event")
	}
}

func (c *Client) routeEvent(ctx context.Context, msgType string, data []byte) {
	if h, ok := c.handler.(HasOnAnyEvent); ok {
		c.dispatchEvent(func() { h.OnAnyEvent(ctx, msgType, data) })
	}

	switch msgType {
	case "SessionMetadata":
		var meta SessionMetadata
		if err := json.Unmarshal(data, &meta); err != nil {
			c.log.Printf("Failed to decode SessionMetadata: %v", err)
			return
		}
		c.highrise.setMyID(meta.UserID)
		c.rateLimiter.apply(meta.RateLimits)
		if h, ok := c.handler.(HasOnStart); ok {
			c.dispatchEvent(func() { h.OnStart(ctx, &meta) })
		}

	case "ChatEvent":
		var ev ChatEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			c.log.Printf("Failed to decode ChatEvent: %v", err)
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
			c.log.Printf("Failed to decode EmoteEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnEmote); ok {
			c.dispatchEvent(func() { h.OnEmote(ctx, ev.User, ev.EmoteID, ev.Receiver) })
		}

	case "ReactionEvent":
		var ev ReactionEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			c.log.Printf("Failed to decode ReactionEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnReaction); ok {
			c.dispatchEvent(func() { h.OnReaction(ctx, ev.User, ev.Reaction, ev.Receiver) })
		}

	case "UserJoinedEvent":
		var ev UserJoinedEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			c.log.Printf("Failed to decode UserJoinedEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnUserJoin); ok {
			c.dispatchEvent(func() { h.OnUserJoin(ctx, ev.User, ev.Position) })
		}

	case "UserLeftEvent":
		var ev UserLeftEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			c.log.Printf("Failed to decode UserLeftEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnUserLeave); ok {
			c.dispatchEvent(func() { h.OnUserLeave(ctx, ev.User) })
		}

	case "UserMovedEvent":
		var ev UserMovedEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			c.log.Printf("Failed to decode UserMovedEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnUserMove); ok {
			c.dispatchEvent(func() { h.OnUserMove(ctx, ev.User, ev.Position) })
		}

	case "TipReactionEvent":
		var ev TipReactionEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			c.log.Printf("Failed to decode TipReactionEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnTip); ok {
			evCopy := ev
			c.dispatchEvent(func() { h.OnTip(ctx, evCopy.Sender, evCopy.Receiver, &evCopy.Item) })
		}

	case "VoiceEvent":
		var ev VoiceEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			c.log.Printf("Failed to decode VoiceEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnVoiceChange); ok {
			c.dispatchEvent(func() { h.OnVoiceChange(ctx, ev.Users, ev.SecondsLeft) })
		}

	case "ChannelEvent":
		var ev ChannelEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			c.log.Printf("Failed to decode ChannelEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnChannel); ok {
			c.dispatchEvent(func() { h.OnChannel(ctx, ev.SenderID, ev.Message, ev.Tags) })
		}

	case "MessageEvent":
		var ev MessageEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			c.log.Printf("Failed to decode MessageEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnMessage); ok {
			c.dispatchEvent(func() { h.OnMessage(ctx, ev.UserID, ev.ConversationID, ev.IsNewConversation) })
		}

	case "RoomModeratedEvent":
		var ev RoomModeratedEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			c.log.Printf("Failed to decode RoomModeratedEvent: %v", err)
			return
		}
		if h, ok := c.handler.(HasOnModerate); ok {
			c.dispatchEvent(func() { h.OnModerate(ctx, ev.ModeratorID, ev.TargetUserID, ev.ModerationType, ev.Duration) })
		}

	case "Error":
		var errMsg Error
		if err := json.Unmarshal(data, &errMsg); err != nil {
			c.log.Printf("Failed to decode Error: %v", err)
			return
		}
		c.totalErrors.Add(1)
		c.log.Printf("Server error: %s", errMsg.Message)
		if h, ok := c.handler.(HasOnError); ok {
			c.dispatchEvent(func() { h.OnError(ctx, errMsg) })
		}
		if errMsg.DoNotReconnect {
			c.log.Printf("Server instructed not to reconnect, stopping...")
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
				c.log.Printf("Keepalive error: %v", err)
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
	c.conn.SetWriteDeadline(time.Now().Add(defaultWriteDeadline))
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

	reqType := extractType(req)
	if reqType != "" {
		if err := c.rateLimiter.acquire(ctx, reqType); err != nil {
			return nil, err
		}
	}

	// Apply default action timeout if not already set in context
	ctx, cancel := c.withActionTimeout(ctx)
	if cancel != nil {
		defer cancel()
	}

	ch := make(chan []byte, 1)
	c.pendingResp.Store(rid, ch)
	defer c.pendingResp.Delete(rid)

	if err := c.writeJSON(req); err != nil {
		return nil, &ConnectionError{
			ReqType: reqType,
			RID:     rid,
			Err:     err,
		}
	}

	c.totalActions.Add(1)

	resp, err := c.waitResponse(ctx, ch, rid, reqType)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) withActionTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.actionTimeout <= 0 {
		return ctx, nil
	}
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, nil
	}
	return context.WithTimeout(ctx, c.actionTimeout)
}

func (c *Client) waitResponse(ctx context.Context, ch chan []byte, rid, reqType string) ([]byte, error) {
	for attempt := 0; attempt < 2; attempt++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%s[%s]: %w", reqType, rid, ctx.Err())
		case <-c.stopCh:
			return nil, NewResponseError(fmt.Sprintf("%s[%s]: client stopped", reqType, rid))
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
			if attempt == 0 {
				c.log.Printf("%s[%s]: timeout, retrying...", reqType, rid)
				if err := c.writeJSON(struct {
					Type string `json:"_type"`
					RID  string `json:"rid"`
				}{Type: reqType, RID: rid}); err != nil {
					return nil, &ConnectionError{
						ReqType: reqType,
						RID:     rid,
						Err:     err,
					}
				}
				continue
			}
			return nil, NewResponseError(
				fmt.Sprintf("%s[%s]: request timed out after retry", reqType, rid),
			)
		}
	}
	return nil, NewResponseError(fmt.Sprintf("%s[%s]: unexpected exit", reqType, rid))
}

// Highrise returns the Highrise action interface for this client
func (c *Client) Highrise() *Highrise {
	return c.highrise
}

// Stop gracefully stops the client and closes the connection
func (c *Client) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
}
