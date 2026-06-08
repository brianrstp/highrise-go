package highrise

import (
	"context"
	"encoding/json"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func respondWith(rid, body string) string {
	return `{"_type":"` + body + `","rid":"` + rid + `"}`
}

func respondWithContent(rid, typeName, contentJSON string) string {
	return `{"_type":"` + typeName + `","content":` + contentJSON + `,"rid":"` + rid + `"}`
}

func actionTestHandler(msg string) string {
	var envelope struct {
		Type string `json:"_type"`
		RID  string `json:"rid"`
	}
	json.Unmarshal([]byte(msg), &envelope)

	if envelope.Type == "KeepaliveRequest" {
		return ""
	}

	switch envelope.Type {
	case "ChatRequest", "SendWhisperRequest", "EmoteRequest",
		"ReactionRequest", "IndicatorRequest", "ChannelRequest",
		"FloorHitRequest", "AnchorHitRequest", "TeleportRequest",
		"ModerateRoomRequest", "ChangeRoomPrivilegeRequest",
		"MoveUserToRoomRequest", "InviteSpeakerRequest",
		"RemoveSpeakerRequest", "SetOutfitRequest",
		"BuyItemRequest", "BuyVoiceTimeRequest", "BuyRoomBoostRequest",
		"TipUserRequest", "LeaveConversationRequest",
		"SendMessageRequest", "SendBulkMessageRequest",
		"ChangeBackpackRequest":
		return respondWith(envelope.RID, envelope.Type)

	case "GetRoomUsersRequest":
		return respondWithContent(envelope.RID, "GetRoomUsersResponse",
			`[{"user":{"id":"u1","username":"alice"},"position":{"x":1,"y":2,"z":3,"facing":"Front"}}]`)

	case "GetWalletRequest":
		return respondWithContent(envelope.RID, "GetWalletResponse",
			`[{"type":"gold","amount":100}]`)

	case "GetRoomPrivilegeRequest":
		return `{"_type":"GetRoomPrivilegeResponse","content":{"moderator":true},"rid":"` + envelope.RID + `"}`

	case "CheckVoiceChatRequest":
		return `{"_type":"CheckVoiceChatResponse","seconds_left":300,"auto_speakers":[],"users":{},"rid":"` + envelope.RID + `"}`

	case "GetUserOutfitRequest":
		return `{"_type":"GetUserOutfitResponse","outfit":[{"type":"shirt","amount":1,"id":"item_1","account_bound":false}],"rid":"` + envelope.RID + `"}`

	case "GetInventoryRequest":
		return `{"_type":"GetInventoryResponse","items":[{"type":"shirt","amount":1,"id":"item_1","account_bound":false}],"rid":"` + envelope.RID + `"}`

	case "GetBackpackRequest":
		return `{"_type":"GetBackpackResponse","backpack":{"gold_bar_10":5},"rid":"` + envelope.RID + `"}`

	case "GetConversationsRequest":
		return `{"_type":"GetConversationsResponse","conversations":[],"not_joined":0,"rid":"` + envelope.RID + `"}`

	case "GetMessagesRequest":
		return `{"_type":"GetMessagesResponse","messages":[],"rid":"` + envelope.RID + `"}`

	case "MessageMediaRequest":
		return `{"_type":"MessageMediaResponse","media":{"type":"image","width":100,"height":100,"mediaSizeInBytes":1000,"thumbnailSizeInBytes":200},"uploadUrl":"https://upload.example.com","thumbnailUploadUrl":"https://thumb.example.com","rid":"` + envelope.RID + `"}`
	}
	return ""
}

func runActionTest(ctx context.Context, t *testing.T, fn func(ctx context.Context, hr *Highrise) error) {
	t.Helper()
	h := newTestHandler()
	ts := newTestServer(t, actionTestHandler)
	defer ts.Close()

	c := NewClient(h)
	c.url = wsURL(ts)

	go c.Run(ctx, "room1", "token1")
	time.Sleep(300 * time.Millisecond)

	if err := fn(ctx, c.Highrise()); err != nil {
		t.Fatal(err)
	}

	c.Stop()
}

func TestActions_Chat(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		return hr.Chat(ctx, "hello")
	})
}

func TestActions_SendWhisper(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		return hr.SendWhisper(ctx, "u2", "secret")
	})
}

func TestActions_WalkTo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		return hr.WalkTo(ctx, Position{X: 10, Y: 0, Z: 5, Facing: "FrontRight"})
	})
}

func TestActions_Teleport(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		return hr.Teleport(ctx, "u2", Position{X: 0, Y: 0, Z: 0, Facing: "FrontRight"})
	})
}

func TestActions_GetRoomUsers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		users, err := hr.GetRoomUsers(ctx)
		if err != nil {
			return err
		}
		if len(users) != 1 || users[0].User.ID != "u1" {
			t.Fatal("unexpected users")
		}
		return nil
	})
}

func TestActions_GetWallet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		wallet, err := hr.GetWallet(ctx)
		if err != nil {
			return err
		}
		if len(wallet) != 1 || wallet[0].Type != "gold" || wallet[0].Amount != 100 {
			t.Fatal("unexpected wallet")
		}
		return nil
	})
}

func TestActions_WalkToAnchor(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		return hr.WalkToAnchor(ctx, AnchorPosition{EntityID: "ent_1", AnchorIx: 0})
	})
}

func TestActions_SendEmote(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		return hr.SendEmote(ctx, "emoji_laugh", nil)
	})
}

func TestActions_React(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		return hr.React(ctx, "heart", "u2")
	})
}

func TestActions_ModerateRoom(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		return hr.ModerateRoom(ctx, "u2", "kick", nil)
	})
}

func TestActions_GetVoiceStatus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		status, err := hr.GetVoiceStatus(ctx)
		if err != nil {
			return err
		}
		if status.SecondsLeft != 300 {
			t.Fatal("unexpected voice status")
		}
		return nil
	})
}

func TestActions_GetUserOutfit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		outfit, err := hr.GetUserOutfit(ctx, "u1")
		if err != nil {
			return err
		}
		if len(outfit) != 1 || outfit[0].ID != "item_1" {
			t.Fatal("unexpected outfit")
		}
		return nil
	})
}

func TestActions_GetInventory(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		items, err := hr.GetInventory(ctx)
		if err != nil {
			return err
		}
		if len(items) != 1 || items[0].ID != "item_1" {
			t.Fatal("unexpected inventory")
		}
		return nil
	})
}

func TestActions_NewClient_Options(t *testing.T) {
	h := newTestHandler()
	customLogger := &testLogger{}
	c := NewClient(h, WithURL("ws://custom"), WithLogger(customLogger), WithSDKVersion("test"))
	if c.url != "ws://custom" {
		t.Fatal("url not set")
	}
	if c.sdkVersion != "test" {
		t.Fatal("sdk version not set")
	}
}

type testLogger struct {
	lastMsg string
}

func (l *testLogger) Printf(format string, v ...any) {
	l.lastMsg = format
}

func TestActions_Reply(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		return hr.Reply(ctx, User{ID: "u1", Username: "alice"}, "hello!")
	})
}

func TestActions_WhisperReply(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runActionTest(ctx, t, func(ctx context.Context, hr *Highrise) error {
		return hr.WhisperReply(ctx, User{ID: "u1", Username: "alice"}, "secret")
	})
}

func TestActions_ConnectionState(t *testing.T) {
	stateCh := make(chan ConnectionState, 10)
	h := &connectionStateHandler{testHandler: newTestHandler(), stateCh: stateCh}
	ts := newTestServer(t, func(msg string) string {
		var envelope struct {
			Type string `json:"_type"`
			RID  string `json:"rid"`
		}
		json.Unmarshal([]byte(msg), &envelope)
		if envelope.Type == "KeepaliveRequest" {
			return ""
		}
		return `{"_type":"` + envelope.Type + `","rid":"` + envelope.RID + `"}`
	})
	defer ts.Close()

	c := NewClient(h)
	c.url = wsURL(ts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go c.Run(ctx, "room1", "token1")

	var gotConnecting, gotConnected bool
	for i := 0; i < 2; i++ {
		select {
		case state := <-stateCh:
			switch state {
			case StateConnecting:
				gotConnecting = true
			case StateConnected:
				gotConnected = true
			}
		case <-time.After(3 * time.Second):
			t.Fatal("timed out waiting for connection state")
		}
	}
	if !gotConnecting || !gotConnected {
		t.Fatal("expected connecting and connected states")
	}

	c.Stop()
}

type connectionStateHandler struct {
	*testHandler
	stateCh chan ConnectionState
}

func (h *connectionStateHandler) OnConnectionChange(ctx context.Context, state ConnectionState) {
	h.stateCh <- state
}

type middlewareRecorder struct {
	called atomic.Bool
}

func TestActions_Middleware(t *testing.T) {
	h := newTestHandler()
	rec := &middlewareRecorder{}

	ts := newTestServer(t, func(msg string) string {
		var envelope struct {
			Type string `json:"_type"`
			RID  string `json:"rid"`
		}
		json.Unmarshal([]byte(msg), &envelope)
		if envelope.Type == "KeepaliveRequest" {
			return ""
		}
		return `{"_type":"` + envelope.Type + `","rid":"` + envelope.RID + `"}`
	})
	defer ts.Close()

	c := NewClient(h)
	c.url = wsURL(ts)
	c.Use(func(next func()) {
		rec.called.Store(true)
		next()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go c.Run(ctx, "room1", "token1")
	time.Sleep(300 * time.Millisecond)

	if !rec.called.Load() {
		t.Fatal("middleware was not called on OnStart event")
	}

	c.Stop()
}

func TestActions_ErrorResponse(t *testing.T) {
	h := newTestHandler()
	ts := newTestServer(t, func(msg string) string {
		var envelope struct {
			Type string `json:"_type"`
			RID  string `json:"rid"`
		}
		json.Unmarshal([]byte(msg), &envelope)
		if envelope.Type == "KeepaliveRequest" {
			return ""
		}
		return `{"_type":"Error","message":"rate limited","rid":"` + envelope.RID + `"}`
	})
	defer ts.Close()

	c := NewClient(h)
	c.url = wsURL(ts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go c.Run(ctx, "room1", "token1")
	time.Sleep(300 * time.Millisecond)

	err := c.Highrise().Chat(ctx, "hello")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("unexpected error: %v", err)
	}

	c.Stop()
}
