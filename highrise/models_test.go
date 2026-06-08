package highrise

import (
	"encoding/json"
	"testing"
)

func TestPositionOrAnchor_Position(t *testing.T) {
	data := `{"x":1,"y":2,"z":3,"facing":"Front"}`
	var p PositionOrAnchor
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		t.Fatal(err)
	}
	if p.Position == nil || p.AnchorPosition != nil {
		t.Fatal("expected Position, got AnchorPosition")
	}
	if p.Position.X != 1 || p.Position.Y != 2 || p.Position.Z != 3 || p.Position.Facing != "Front" {
		t.Fatal("unexpected position values")
	}
	out, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var back PositionOrAnchor
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatal(err)
	}
	if back.Position == nil {
		t.Fatal("round-trip failed")
	}
}

func TestPositionOrAnchor_Anchor(t *testing.T) {
	data := `{"entity_id":"ent123","anchor_ix":1}`
	var p PositionOrAnchor
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		t.Fatal(err)
	}
	if p.AnchorPosition == nil || p.Position != nil {
		t.Fatal("expected AnchorPosition, got Position")
	}
	if p.AnchorPosition.EntityID != "ent123" || p.AnchorPosition.AnchorIx != 1 {
		t.Fatal("unexpected anchor values")
	}
	out, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var back PositionOrAnchor
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatal(err)
	}
	if back.AnchorPosition == nil {
		t.Fatal("round-trip failed")
	}
}

func TestTipItem_Currency(t *testing.T) {
	data := `{"type":"gold","amount":100}`
	var ti TipItem
	if err := json.Unmarshal([]byte(data), &ti); err != nil {
		t.Fatal(err)
	}
	if ti.CurrencyItem == nil || ti.Item != nil {
		t.Fatal("expected CurrencyItem")
	}
	if ti.CurrencyItem.Type != "gold" || ti.CurrencyItem.Amount != 100 {
		t.Fatal("unexpected currency values")
	}
}

func TestTipItem_Item(t *testing.T) {
	data := `{"type":"clothing","amount":1,"id":"item_123","account_bound":false}`
	var ti TipItem
	if err := json.Unmarshal([]byte(data), &ti); err != nil {
		t.Fatal(err)
	}
	if ti.Item == nil || ti.CurrencyItem != nil {
		t.Fatal("expected Item")
	}
	if ti.Item.Type != "clothing" || ti.Item.ID != "item_123" {
		t.Fatal("unexpected item values")
	}
}

func TestUserVoiceStatus(t *testing.T) {
	data := `[{"id":"u1","username":"alice"},"speaking"]`
	var s UserVoiceStatus
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		t.Fatal(err)
	}
	if s.User.ID != "u1" || s.User.Username != "alice" {
		t.Fatal("unexpected user values")
	}
	if s.Status != "speaking" {
		t.Fatal("expected speaking status")
	}
}

func TestUserVoiceStatus_TooFew(t *testing.T) {
	data := `[{"id":"u1","username":"alice"}]`
	var s UserVoiceStatus
	if err := json.Unmarshal([]byte(data), &s); err == nil {
		t.Fatal("expected error for wrong length")
	}
}

func TestSessionMetadata(t *testing.T) {
	data := `{"_type":"SessionMetadata","user_id":"u1","room_info":{"owner_id":"owner1","room_name":"TestRoom"},"rate_limits":{},"connection_id":"conn1"}`
	var meta SessionMetadata
	if err := json.Unmarshal([]byte(data), &meta); err != nil {
		t.Fatal(err)
	}
	if meta.UserID != "u1" || meta.RoomInfo.RoomName != "TestRoom" || meta.ConnectionID != "conn1" {
		t.Fatal("unexpected session metadata")
	}
}

func TestChatEvent(t *testing.T) {
	data := `{"_type":"ChatEvent","user":{"id":"u1","username":"alice"},"message":"hello","whisper":false}`
	var ev ChatEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.User.ID != "u1" || ev.Message != "hello" || ev.Whisper {
		t.Fatal("unexpected chat event")
	}
}

func TestEmoteEvent(t *testing.T) {
	data := `{"_type":"EmoteEvent","user":{"id":"u1","username":"alice"},"emote_id":"wave","receiver":{"id":"u2","username":"bob"}}`
	var ev EmoteEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.EmoteID != "wave" || ev.Receiver == nil || ev.Receiver.ID != "u2" {
		t.Fatal("unexpected emote event")
	}
}

func TestUserJoinedEvent(t *testing.T) {
	data := `{"_type":"UserJoinedEvent","user":{"id":"u1","username":"alice"},"position":{"x":0,"y":1,"z":2,"facing":"Front"}}`
	var ev UserJoinedEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.User.ID != "u1" || ev.Position.Position == nil {
		t.Fatal("unexpected user joined event")
	}
}

func TestTipReactionEvent(t *testing.T) {
	data := `{"_type":"TipReactionEvent","sender":{"id":"u1","username":"alice"},"receiver":{"id":"u2","username":"bob"},"item":{"type":"gold","amount":50}}`
	var ev TipReactionEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.Sender.ID != "u1" || ev.Receiver.ID != "u2" {
		t.Fatal("unexpected tip event")
	}
	if ev.Item.CurrencyItem == nil || ev.Item.CurrencyItem.Amount != 50 {
		t.Fatal("unexpected tip item")
	}
}

func TestVoiceEvent(t *testing.T) {
	data := `{"_type":"VoiceEvent","users":[[{"id":"u1","username":"alice"},"speaking"]],"seconds_left":300}`
	var ev VoiceEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		t.Fatal(err)
	}
	if len(ev.Users) != 1 || ev.Users[0].User.ID != "u1" || ev.Users[0].Status != "speaking" {
		t.Fatal("unexpected voice event")
	}
	if ev.SecondsLeft != 300 {
		t.Fatal("unexpected seconds left")
	}
}

func TestChannelEvent(t *testing.T) {
	data := `{"_type":"ChannelEvent","sender_id":"u1","msg":"hello","tags":["tag1"]}`
	var ev ChannelEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.SenderID != "u1" || ev.Message != "hello" || len(ev.Tags) != 1 {
		t.Fatal("unexpected channel event")
	}
}

func TestMessageEvent(t *testing.T) {
	data := `{"_type":"MessageEvent","user_id":"u1","conversation_id":"conv1","is_new_conversation":true}`
	var ev MessageEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.UserID != "u1" || ev.ConversationID != "conv1" || !ev.IsNewConversation {
		t.Fatal("unexpected message event")
	}
}

func TestRoomModeratedEvent(t *testing.T) {
	data := `{"_type":"RoomModeratedEvent","moderatorId":"mod1","targetUserId":"u1","moderationType":"mute","duration":60}`
	var ev RoomModeratedEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.ModeratorID != "mod1" || ev.ModerationType != "mute" || *ev.Duration != 60 {
		t.Fatal("unexpected moderation event")
	}
}

func TestChatRequest(t *testing.T) {
	req := ChatRequest{Type: "ChatRequest", Message: "hello", RID: "1"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var back ChatRequest
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Message != "hello" || back.RID != "1" {
		t.Fatal("round-trip failed")
	}
}

func TestChatRequest_Whisper(t *testing.T) {
	target := "u2"
	req := ChatRequest{Type: "ChatRequest", Message: "secret", WhisperTarget: &target, RID: "2"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var back ChatRequest
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.WhisperTarget == nil || *back.WhisperTarget != "u2" {
		t.Fatal("whisper target lost")
	}
}

func TestReactionRequest(t *testing.T) {
	req := ReactionRequest{Type: "ReactionRequest", Reaction: "like", TargetUserID: "u2", RID: "3"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var back ReactionRequest
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Reaction != "like" || back.TargetUserID != "u2" {
		t.Fatal("round-trip failed")
	}
}

func TestRoomPermissions(t *testing.T) {
	mod := true
	perms := RoomPermissions{Moderator: &mod}
	data, err := json.Marshal(perms)
	if err != nil {
		t.Fatal(err)
	}
	var back RoomPermissions
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Moderator == nil || !*back.Moderator {
		t.Fatal("moderator permission lost")
	}
}

func TestError(t *testing.T) {
	data := `{"_type":"Error","message":"test error","do_not_reconnect":true}`
	var errMsg Error
	if err := json.Unmarshal([]byte(data), &errMsg); err != nil {
		t.Fatal(err)
	}
	if errMsg.Message != "test error" || !errMsg.DoNotReconnect {
		t.Fatal("unexpected error")
	}
}

func TestGetWalletResponse(t *testing.T) {
	data := `{"_type":"GetWalletResponse","content":[{"type":"gold","amount":100}],"rid":"1"}`
	var resp GetWalletResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Content) != 1 || resp.Content[0].Type != "gold" {
		t.Fatal("unexpected wallet")
	}
}

func TestGetRoomUsersResponse(t *testing.T) {
	data := `{"_type":"GetRoomUsersResponse","content":[{"user":{"id":"u1","username":"alice"},"position":{"x":0,"y":1,"z":2,"facing":"Front"}}],"rid":"1"}`
	var resp GetRoomUsersResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Content) != 1 || resp.Content[0].User.ID != "u1" {
		t.Fatal("unexpected room users")
	}
}
