package highrise

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type chatEvent struct {
	user    User
	message string
}

type joinEvent struct {
	user User
	pos  PositionOrAnchor
}

type testHandler struct {
	onStart     chan *SessionMetadata
	onChat      chan chatEvent
	onUserJoin  chan joinEvent
	onUserLeave chan User
}

func newTestHandler() *testHandler {
	return &testHandler{
		onStart:     make(chan *SessionMetadata, 1),
		onChat:      make(chan chatEvent, 1),
		onUserJoin:  make(chan joinEvent, 1),
		onUserLeave: make(chan User, 1),
	}
}

func (h *testHandler) OnStart(ctx context.Context, s *SessionMetadata) {
	h.onStart <- s
}
func (h *testHandler) OnChat(ctx context.Context, user User, message string) {
	h.onChat <- chatEvent{user, message}
}
func (h *testHandler) OnUserJoin(ctx context.Context, user User, pos PositionOrAnchor) {
	h.onUserJoin <- joinEvent{user, pos}
}
func (h *testHandler) OnUserLeave(ctx context.Context, user User) {
	h.onUserLeave <- user
}
func (h *testHandler) OnWhisper(ctx context.Context, user User, message string)                    {}
func (h *testHandler) OnEmote(ctx context.Context, user User, emoteID string, receiver *User)      {}
func (h *testHandler) OnReaction(ctx context.Context, user User, reaction string, receiver User)   {}
func (h *testHandler) OnUserMove(ctx context.Context, user User, position PositionOrAnchor)        {}
func (h *testHandler) OnTip(ctx context.Context, sender, receiver User, item *TipItem)             {}
func (h *testHandler) OnVoiceChange(ctx context.Context, users []UserVoiceStatus, secondsLeft int) {}
func (h *testHandler) OnChannel(ctx context.Context, senderID, message string, tags []string)      {}
func (h *testHandler) OnMessage(ctx context.Context, userID, conversationID string, isNewConversation bool) {
}
func (h *testHandler) OnModerate(ctx context.Context, moderatorID, targetUserID, moderationType string, duration *int) {
}

func newTestServer(t *testing.T, handler func(string) string) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal("upgrade:", err)
		}
		defer conn.Close()

		// Send SessionMetadata immediately on connect
		conn.WriteMessage(websocket.TextMessage, []byte(`{"_type":"SessionMetadata","user_id":"u1","room_info":{"owner_id":"o1","room_name":"TestRoom"},"rate_limits":{},"connection_id":"c1"}`))

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			resp := handler(string(msg))
			if resp != "" {
				conn.WriteMessage(websocket.TextMessage, []byte(resp))
			}
		}
	}))
}

func wsURL(ts *httptest.Server) string {
	return "ws" + strings.TrimPrefix(ts.URL, "http")
}

func TestClient_Connect(t *testing.T) {
	h := newTestHandler()
	ts := newTestServer(t, func(msg string) string {
		return `{"_type":"SessionMetadata","user_id":"u1","room_info":{"owner_id":"o1","room_name":"TestRoom"},"rate_limits":{},"connection_id":"c1"}`
	})
	defer ts.Close()

	c := NewClient(h)
	c.url = wsURL(ts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		c.Run(ctx, "room1", "token1")
	}()

	select {
	case meta := <-h.onStart:
		if meta.UserID != "u1" || meta.RoomInfo.RoomName != "TestRoom" {
			t.Fatal("bad metadata")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for OnStart")
	}

	c.Stop()
}

func TestClient_Chat(t *testing.T) {
	h := newTestHandler()
	ts := newTestServer(t, func(msg string) string {
		var envelope struct {
			Type string `json:"_type"`
			RID  string `json:"rid"`
		}
		json.Unmarshal([]byte(msg), &envelope)
		if envelope.Type == "ChatRequest" {
			return `{"_type":"ChatRequest","rid":"` + envelope.RID + `"}`
		}
		if envelope.Type == "KeepaliveRequest" {
			return ""
		}
		return ""
	})
	defer ts.Close()

	c := NewClient(h)
	c.url = wsURL(ts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		c.Run(ctx, "room1", "token1")
	}()

	// Wait for connection then send chat
	time.Sleep(500 * time.Millisecond)
	err := c.Highrise().Chat(ctx, "hello")
	if err != nil {
		t.Fatal("chat error:", err)
	}

	c.Stop()
}

func TestClient_SendRequest_TimedOut(t *testing.T) {
	h := newTestHandler()
	ts := newTestServer(t, func(msg string) string {
		// Never respond to requests
		return ""
	})
	defer ts.Close()

	c := NewClient(h)
	c.url = wsURL(ts)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go func() {
		c.Run(ctx, "room1", "token1")
	}()

	time.Sleep(200 * time.Millisecond)

	// Very short timeout
	shortCtx, shortCancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer shortCancel()

	err := c.Highrise().Chat(shortCtx, "hello")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	t.Log("got expected error:", err)

	c.Stop()
}

func TestClient_Stop_NoPanic(t *testing.T) {
	h := newTestHandler()
	c := NewClient(h)

	// Stop twice should not panic
	c.Stop()
	c.Stop()
}

func TestClient_EventRouting_Chat(t *testing.T) {
	h := newTestHandler()
	ts := newTestServer(t, func(msg string) string {
		if strings.Contains(msg, `"KeepaliveRequest"`) {
			return ""
		}
		return ""
	})
	defer ts.Close()

	c := NewClient(h)
	c.url = wsURL(ts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Manually trigger session metadata to start the bot
	go c.Run(ctx, "room1", "token1")

	time.Sleep(300 * time.Millisecond)

	// Directly inject a chat event message using handleMessage
	chatJSON := `{"_type":"ChatEvent","user":{"id":"u1","username":"alice"},"message":"hello","whisper":false}`
	c.handleMessage(ctx, []byte(chatJSON))

	select {
	case ev := <-h.onChat:
		if ev.user.ID != "u1" || ev.message != "hello" {
			t.Fatal("bad chat event")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for chat event")
	}

	c.Stop()
}

func TestClient_EventRouting_UserJoined(t *testing.T) {
	h := newTestHandler()
	c := NewClient(h)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	joinJSON := `{"_type":"UserJoinedEvent","user":{"id":"u1","username":"alice"},"position":{"x":1,"y":2,"z":3,"facing":"Front"}}`
	c.handleMessage(ctx, []byte(joinJSON))

	select {
	case ev := <-h.onUserJoin:
		if ev.user.ID != "u1" {
			t.Fatal("bad user ID")
		}
		if ev.pos.Position == nil || ev.pos.Position.X != 1 {
			t.Fatal("bad position")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for join event")
	}

}

func TestClient_EventRouting_UserLeft(t *testing.T) {
	h := newTestHandler()
	c := NewClient(h)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	leaveJSON := `{"_type":"UserLeftEvent","user":{"id":"u1","username":"alice"}}`
	c.handleMessage(ctx, []byte(leaveJSON))

	select {
	case user := <-h.onUserLeave:
		if user.ID != "u1" {
			t.Fatal("bad user ID")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for leave event")
	}
}

type anyEventHandler struct {
	Bot
	onAnyEvent chan string
}

func (h *anyEventHandler) OnAnyEvent(ctx context.Context, eventType string, data []byte) {
	h.onAnyEvent <- eventType
}

func TestClient_OnAnyEvent(t *testing.T) {
	h := &anyEventHandler{onAnyEvent: make(chan string, 3)}
	c := NewClient(h)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Inject ChatEvent
	c.handleMessage(ctx, []byte(`{"_type":"ChatEvent","user":{"id":"u1","username":"alice"},"message":"hello","whisper":false}`))

	// Inject UserJoinedEvent
	c.handleMessage(ctx, []byte(`{"_type":"UserJoinedEvent","user":{"id":"u1","username":"alice"},"position":{"x":1,"y":2,"z":3,"facing":"Front"}}`))

	// Inject unknown event
	c.handleMessage(ctx, []byte(`{"_type":"CustomEvent","data":"test"}`))

	got := make(map[string]int)
	for i := 0; i < 3; i++ {
		select {
		case ev := <-h.onAnyEvent:
			got[ev]++
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out, got %v so far", got)
		}
	}
	if got["ChatEvent"] != 1 || got["UserJoinedEvent"] != 1 || got["CustomEvent"] != 1 {
		t.Fatalf("unexpected event set: %v", got)
	}
}

func TestClient_ResponseRouting(t *testing.T) {
	c := NewClient(newTestHandler())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	respJSON := `{"_type":"GetRoomUsersResponse","content":[],"rid":"test123"}`

	// Store pending response and then send the response
	ch := make(chan []byte, 1)
	c.pendingResp.Store("test123", ch)
	defer c.pendingResp.Delete("test123")

	// Simulate response arriving
	c.handleMessage(ctx, []byte(respJSON))

	select {
	case <-ch:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("response not routed")
	}
}
