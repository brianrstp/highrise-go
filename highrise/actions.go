package highrise

import (
	"context"
	"encoding/json"
	"fmt"
)

type Highrise struct {
	client *Client
	myID   string
}

func newHighrise(client *Client) *Highrise {
	return &Highrise{client: client}
}

func (h *Highrise) nextRID() string {
	return fmt.Sprintf("%d", h.client.nextReqID())
}

func (h *Highrise) sendRequest(ctx context.Context, msg any) ([]byte, error) {
	return h.client.sendRequest(ctx, msg)
}

func (h *Highrise) Chat(ctx context.Context, message string) error {
	req := ChatRequest{
		Type:    "ChatRequest",
		Message: message,
		RID:     h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) SendWhisper(ctx context.Context, userID, message string) error {
	req := ChatRequest{
		Type:          "ChatRequest",
		Message:       message,
		WhisperTarget: &userID,
		RID:           h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) SendEmote(ctx context.Context, emoteID string, targetUserID *string) error {
	req := EmoteRequest{
		Type:    "EmoteRequest",
		EmoteID: emoteID,
		RID:     h.nextRID(),
	}
	if targetUserID != nil {
		req.TargetUserID = targetUserID
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) React(ctx context.Context, reaction, targetUserID string) error {
	req := ReactionRequest{
		Type:         "ReactionRequest",
		Reaction:     reaction,
		TargetUserID: targetUserID,
		RID:          h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) SetIndicator(ctx context.Context, icon *string) error {
	req := IndicatorRequest{
		Type: "IndicatorRequest",
		Icon: icon,
		RID:  h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) SendChannel(ctx context.Context, message string, tags []string) error {
	req := ChannelRequest{
		Type:    "ChannelRequest",
		Message: message,
		Tags:    tags,
		RID:     h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) WalkTo(ctx context.Context, dest Position) error {
	req := FloorHitRequest{
		Type:        "FloorHitRequest",
		Destination: dest,
		RID:         h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) WalkToAnchor(ctx context.Context, anchor AnchorPosition) error {
	req := AnchorHitRequest{
		Type:   "AnchorHitRequest",
		Anchor: anchor,
		RID:    h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) Teleport(ctx context.Context, userID string, dest Position) error {
	req := TeleportRequest{
		Type:        "TeleportRequest",
		UserID:      userID,
		Destination: dest,
		RID:         h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) ModerateRoom(ctx context.Context, userID, action string, actionLength *int) error {
	req := ModerateRoomRequest{
		Type:             "ModerateRoomRequest",
		UserID:           userID,
		ModerationAction: action,
		ActionLength:     actionLength,
		RID:              h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) ChangeRoomPrivilege(ctx context.Context, userID string, perms RoomPermissions) error {
	req := ChangeRoomPrivilegeRequest{
		Type:        "ChangeRoomPrivilegeRequest",
		UserID:      userID,
		Permissions: perms,
		RID:         h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) MoveUserToRoom(ctx context.Context, userID, roomID string) error {
	req := MoveUserToRoomRequest{
		Type:   "MoveUserToRoomRequest",
		UserID: userID,
		RoomID: roomID,
		RID:    h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) InviteSpeaker(ctx context.Context, userID string) error {
	req := InviteSpeakerRequest{
		Type:   "InviteSpeakerRequest",
		UserID: userID,
		RID:    h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) RemoveSpeaker(ctx context.Context, userID string) error {
	req := RemoveSpeakerRequest{
		Type:   "RemoveSpeakerRequest",
		UserID: userID,
		RID:    h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) GetRoomUsers(ctx context.Context) ([]SessionUserEntry, error) {
	req := GetRoomUsersRequest{
		Type: "GetRoomUsersRequest",
		RID:  h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp GetRoomUsersResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Content, nil
}

func (h *Highrise) GetWallet(ctx context.Context) ([]CurrencyItem, error) {
	req := GetWalletRequest{
		Type: "GetWalletRequest",
		RID:  h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp GetWalletResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Content, nil
}

func (h *Highrise) GetRoomPrivilege(ctx context.Context, userID string) (*RoomPermissions, error) {
	req := GetRoomPrivilegeRequest{
		Type:   "GetRoomPrivilegeRequest",
		UserID: userID,
		RID:    h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp GetRoomPrivilegeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Content, nil
}

func (h *Highrise) GetVoiceStatus(ctx context.Context) (*CheckVoiceChatResponse, error) {
	req := CheckVoiceChatRequest{
		Type: "CheckVoiceChatRequest",
		RID:  h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp CheckVoiceChatResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (h *Highrise) GetUserOutfit(ctx context.Context, userID string) ([]Item, error) {
	req := GetUserOutfitRequest{
		Type:   "GetUserOutfitRequest",
		UserID: userID,
		RID:    h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp GetUserOutfitResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Outfit, nil
}

func (h *Highrise) GetBackpack(ctx context.Context, userID string) (map[string]int, error) {
	req := GetBackpackRequest{
		Type:   "GetBackpackRequest",
		UserID: userID,
		RID:    h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp GetBackpackResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Backpack, nil
}

func (h *Highrise) ChangeBackpack(ctx context.Context, userID string, changes map[string]int) error {
	req := ChangeBackpackRequest{
		Type:    "ChangeBackpackRequest",
		UserID:  userID,
		Changes: changes,
		RID:     h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) GetConversations(ctx context.Context, notJoined bool, lastID *string) ([]Conversation, int, error) {
	req := GetConversationsRequest{
		Type:      "GetConversationsRequest",
		NotJoined: notJoined,
		LastID:    lastID,
		RID:       h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	var resp GetConversationsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, 0, err
	}
	return resp.Conversations, resp.NotJoined, nil
}

func (h *Highrise) SendMessage(ctx context.Context, conversationID, content, msgType string, roomID, worldID, mediaID *string) error {
	req := SendMessageRequest{
		Type:           "SendMessageRequest",
		ConversationID: conversationID,
		Content:        content,
		MessageType:    msgType,
		RoomID:         roomID,
		WorldID:        worldID,
		MediaID:        mediaID,
		RID:            h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) SendBulkMessage(ctx context.Context, userIDs []string, content, msgType string, roomID, worldID *string) error {
	req := SendBulkMessageRequest{
		Type:        "SendBulkMessageRequest",
		UserIDs:     userIDs,
		Content:     content,
		MessageType: msgType,
		RoomID:      roomID,
		WorldID:     worldID,
		RID:         h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) GetMessages(ctx context.Context, conversationID, lastMessageID string) ([]Message, error) {
	req := GetMessagesRequest{
		Type:           "GetMessagesRequest",
		ConversationID: conversationID,
		LastMessageID:  lastMessageID,
		RID:            h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp GetMessagesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

func (h *Highrise) LeaveConversation(ctx context.Context, conversationID string) error {
	req := LeaveConversationRequest{
		Type:           "LeaveConversationRequest",
		ConversationID: conversationID,
		RID:            h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) BuyVoiceTime(ctx context.Context) (string, error) {
	req := BuyVoiceTimeRequest{
		Type:          "BuyVoiceTimeRequest",
		PaymentMethod: "bot_wallet_only",
		RID:           h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return "", err
	}
	var resp BuyVoiceTimeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.Result, nil
}

func (h *Highrise) BuyRoomBoost(ctx context.Context, amount int) (string, error) {
	req := BuyRoomBoostRequest{
		Type:          "BuyRoomBoostRequest",
		PaymentMethod: "bot_wallet_only",
		Amount:        amount,
		RID:           h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return "", err
	}
	var resp BuyRoomBoostResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.Result, nil
}

func (h *Highrise) TipUser(ctx context.Context, userID, goldBar string) (string, error) {
	req := TipUserRequest{
		Type:    "TipUserRequest",
		UserID:  userID,
		GoldBar: goldBar,
		RID:     h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return "", err
	}
	var resp TipUserResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.Result, nil
}

func (h *Highrise) GetInventory(ctx context.Context) ([]Item, error) {
	req := GetInventoryRequest{
		Type: "GetInventoryRequest",
		RID:  h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp GetInventoryResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (h *Highrise) SetOutfit(ctx context.Context, outfit []Item) error {
	req := SetOutfitRequest{
		Type:   "SetOutfitRequest",
		Outfit: outfit,
		RID:    h.nextRID(),
	}
	_, err := h.sendRequest(ctx, req)
	return err
}

func (h *Highrise) BuyItem(ctx context.Context, itemID string) (string, error) {
	req := BuyItemRequest{
		Type:   "BuyItemRequest",
		ItemID: itemID,
		RID:    h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return "", err
	}
	var resp BuyItemResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.Result, nil
}

func (h *Highrise) GetMyOutfit(ctx context.Context) ([]Item, error) {
	req := GetUserOutfitRequest{
		Type:   "GetUserOutfitRequest",
		UserID: h.myID,
		RID:    h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp GetUserOutfitResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Outfit, nil
}

func (h *Highrise) MessageMediaUpload(ctx context.Context, media MessageMedia) (*MessageMediaResponse, error) {
	req := MessageMediaRequest{
		Type:  "MessageMediaRequest",
		Media: media,
		RID:   h.nextRID(),
	}
	data, err := h.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp MessageMediaResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
