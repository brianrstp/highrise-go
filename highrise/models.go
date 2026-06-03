package highrise

import (
	"encoding/json"
	"fmt"
)

// Position represents a 3D position in the room
type Position struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Z      float64 `json:"z"`
	Facing string  `json:"facing"`
}

// AnchorPosition represents a position anchored to an entity
type AnchorPosition struct {
	EntityID string `json:"entity_id"`
	AnchorIx int    `json:"anchor_ix"`
}

// PositionOrAnchor can be either a Position or AnchorPosition
type PositionOrAnchor struct {
	Position       *Position
	AnchorPosition *AnchorPosition
}

func (p *PositionOrAnchor) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if _, ok := raw["entity_id"]; ok {
		var ap AnchorPosition
		if err := json.Unmarshal(data, &ap); err != nil {
			return err
		}
		p.AnchorPosition = &ap
	} else {
		var pos Position
		if err := json.Unmarshal(data, &pos); err != nil {
			return err
		}
		p.Position = &pos
	}
	return nil
}

func (p PositionOrAnchor) MarshalJSON() ([]byte, error) {
	if p.AnchorPosition != nil {
		return json.Marshal(p.AnchorPosition)
	}
	return json.Marshal(p.Position)
}

// TipItem can be either a CurrencyItem or Item
type TipItem struct {
	CurrencyItem *CurrencyItem
	Item         *Item
}

func (t *TipItem) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if _, ok := raw["id"]; ok {
		var item Item
		if err := json.Unmarshal(data, &item); err != nil {
			return err
		}
		t.Item = &item
	} else {
		var ci CurrencyItem
		if err := json.Unmarshal(data, &ci); err != nil {
			return err
		}
		t.CurrencyItem = &ci
	}
	return nil
}

// UserVoiceStatus is a tuple of User and voice status
type UserVoiceStatus struct {
	User   User
	Status string
}

func (s *UserVoiceStatus) UnmarshalJSON(data []byte) error {
	var raw []any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 2 {
		return fmt.Errorf("expected 2 elements in UserVoiceStatus, got %d", len(raw))
	}
	userBytes, err := json.Marshal(raw[0])
	if err != nil {
		return err
	}
	if err := json.Unmarshal(userBytes, &s.User); err != nil {
		return err
	}
	s.Status, _ = raw[1].(string)
	return nil
}

// User represents a Highrise user
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// SessionMetadata received on connection
type SessionMetadata struct {
	Type         string           `json:"_type"`
	UserID       string           `json:"user_id"`
	RoomInfo     RoomInfo         `json:"room_info"`
	RateLimits   map[string][]any `json:"rate_limits"`
	ConnectionID string           `json:"connection_id"`
	SDKVersion   *string          `json:"sdk_version,omitempty"`
}

type SessionUserEntry struct {
	User     User             `json:"user"`
	Position PositionOrAnchor `json:"position"`
}

// RoomInfo contains room metadata
type RoomInfo struct {
	OwnerID  string `json:"owner_id"`
	RoomName string `json:"room_name"`
}

// CurrencyItem represents a currency tip
type CurrencyItem struct {
	Type   string `json:"type"`
	Amount int    `json:"amount"`
}

// Item represents a clothing or item
type Item struct {
	Type          string `json:"type"`
	Amount        int    `json:"amount"`
	ID            string `json:"id"`
	AccountBound  bool   `json:"account_bound"`
	ActivePalette *int   `json:"active_palette,omitempty"`
}

// RoomPermissions for room privilege management
type RoomPermissions struct {
	Moderator *bool `json:"moderator,omitempty"`
	Designer  *bool `json:"designer,omitempty"`
}

// IncomingMessage is the raw envelope from server
type IncomingMessage struct {
	Type string  `json:"_type"`
	RID  *string `json:"rid,omitempty"`
}

// OutgoingMessage is the base for all outgoing messages
type OutgoingMessage struct {
	Type string `json:"_type"`
	RID  string `json:"rid"`
}

// Error message from server
type Error struct {
	Type           string  `json:"_type"`
	Message        string  `json:"message"`
	DoNotReconnect bool    `json:"do_not_reconnect"`
	RID            *string `json:"rid,omitempty"`
}

// ChatEvent
type ChatEvent struct {
	Type    string `json:"_type"`
	User    User   `json:"user"`
	Message string `json:"message"`
	Whisper bool   `json:"whisper"`
}

// EmoteEvent
type EmoteEvent struct {
	Type     string `json:"_type"`
	User     User   `json:"user"`
	EmoteID  string `json:"emote_id"`
	Receiver *User  `json:"receiver,omitempty"`
}

// ReactionEvent
type ReactionEvent struct {
	Type     string `json:"_type"`
	User     User   `json:"user"`
	Reaction string `json:"reaction"`
	Receiver User   `json:"receiver"`
}

// UserJoinedEvent
type UserJoinedEvent struct {
	Type     string           `json:"_type"`
	User     User             `json:"user"`
	Position PositionOrAnchor `json:"position"`
}

// UserLeftEvent
type UserLeftEvent struct {
	Type string `json:"_type"`
	User User   `json:"user"`
}

// UserMovedEvent
type UserMovedEvent struct {
	Type     string           `json:"_type"`
	User     User             `json:"user"`
	Position PositionOrAnchor `json:"position"`
}

// TipReactionEvent
type TipReactionEvent struct {
	Type     string  `json:"_type"`
	Sender   User    `json:"sender"`
	Receiver User    `json:"receiver"`
	Item     TipItem `json:"item"`
}

// ChannelEvent
type ChannelEvent struct {
	Type     string   `json:"_type"`
	SenderID string   `json:"sender_id"`
	Message  string   `json:"msg"`
	Tags     []string `json:"tags"`
}

// VoiceEvent
type VoiceEvent struct {
	Type        string            `json:"_type"`
	Users       []UserVoiceStatus `json:"users"`
	SecondsLeft int               `json:"seconds_left"`
}

// MessageEvent
type MessageEvent struct {
	Type              string `json:"_type"`
	UserID            string `json:"user_id"`
	ConversationID    string `json:"conversation_id"`
	IsNewConversation bool   `json:"is_new_conversation"`
}

// RoomModeratedEvent
type RoomModeratedEvent struct {
	Type           string `json:"_type"`
	ModeratorID    string `json:"moderatorId"`
	TargetUserID   string `json:"targetUserId"`
	ModerationType string `json:"moderationType"`
	Duration       *int   `json:"duration,omitempty"`
}

// --- Requests ---

type ChatRequest struct {
	Type          string  `json:"_type"`
	Message       string  `json:"message"`
	WhisperTarget *string `json:"whisper_target_id"`
	RID           string  `json:"rid"`
}

func (r ChatRequest) getRID() string { return r.RID }

type EmoteRequest struct {
	Type         string  `json:"_type"`
	EmoteID      string  `json:"emote_id"`
	TargetUserID *string `json:"target_user_id"`
	RID          string  `json:"rid"`
}

func (r EmoteRequest) getRID() string { return r.RID }

type ReactionRequest struct {
	Type         string `json:"_type"`
	Reaction     string `json:"reaction"`
	TargetUserID string `json:"target_user_id"`
	RID          string `json:"rid"`
}

func (r ReactionRequest) getRID() string { return r.RID }

type IndicatorRequest struct {
	Type string  `json:"_type"`
	Icon *string `json:"icon"`
	RID  string  `json:"rid"`
}

func (r IndicatorRequest) getRID() string { return r.RID }

type ChannelRequest struct {
	Type    string   `json:"_type"`
	Message string   `json:"message"`
	Tags    []string `json:"tags"`
	RID     string   `json:"rid"`
}

func (r ChannelRequest) getRID() string { return r.RID }

type FloorHitRequest struct {
	Type        string   `json:"_type"`
	Destination Position `json:"destination"`
	RID         string   `json:"rid"`
}

func (r FloorHitRequest) getRID() string { return r.RID }

type AnchorHitRequest struct {
	Type   string         `json:"_type"`
	Anchor AnchorPosition `json:"anchor"`
	RID    string         `json:"rid"`
}

func (r AnchorHitRequest) getRID() string { return r.RID }

type TeleportRequest struct {
	Type        string   `json:"_type"`
	UserID      string   `json:"user_id"`
	Destination Position `json:"destination"`
	RID         string   `json:"rid"`
}

func (r TeleportRequest) getRID() string { return r.RID }

type ModerateRoomRequest struct {
	Type             string `json:"_type"`
	UserID           string `json:"user_id"`
	ModerationAction string `json:"moderation_action"`
	ActionLength     *int   `json:"action_length,omitempty"`
	RID              string `json:"rid"`
}

func (r ModerateRoomRequest) getRID() string { return r.RID }

type ChangeRoomPrivilegeRequest struct {
	Type        string          `json:"_type"`
	UserID      string          `json:"user_id"`
	Permissions RoomPermissions `json:"permissions"`
	RID         string          `json:"rid"`
}

func (r ChangeRoomPrivilegeRequest) getRID() string { return r.RID }

type MoveUserToRoomRequest struct {
	Type   string `json:"_type"`
	UserID string `json:"user_id"`
	RoomID string `json:"room_id"`
	RID    string `json:"rid"`
}

func (r MoveUserToRoomRequest) getRID() string { return r.RID }

type InviteSpeakerRequest struct {
	Type   string `json:"_type"`
	UserID string `json:"user_id"`
	RID    string `json:"rid"`
}

func (r InviteSpeakerRequest) getRID() string { return r.RID }

type RemoveSpeakerRequest struct {
	Type   string `json:"_type"`
	UserID string `json:"user_id"`
	RID    string `json:"rid"`
}

func (r RemoveSpeakerRequest) getRID() string { return r.RID }

type KeepaliveRequest struct {
	Type string  `json:"_type"`
	RID  *string `json:"rid,omitempty"`
}

type SendMessageRequest struct {
	Type           string  `json:"_type"`
	ConversationID string  `json:"conversation_id"`
	Content        string  `json:"content"`
	MessageType    string  `json:"type"`
	RoomID         *string `json:"room_id,omitempty"`
	WorldID        *string `json:"world_id,omitempty"`
	MediaID        *string `json:"media_id,omitempty"`
	RID            string  `json:"rid"`
}

func (r SendMessageRequest) getRID() string { return r.RID }

type SendBulkMessageRequest struct {
	Type        string   `json:"_type"`
	UserIDs     []string `json:"user_ids"`
	Content     string   `json:"content"`
	MessageType string   `json:"type"`
	RoomID      *string  `json:"room_id,omitempty"`
	WorldID     *string  `json:"world_id,omitempty"`
	RID         string   `json:"rid"`
}

func (r SendBulkMessageRequest) getRID() string { return r.RID }

type LeaveConversationRequest struct {
	Type           string `json:"_type"`
	ConversationID string `json:"conversation_id"`
	RID            string `json:"rid"`
}

func (r LeaveConversationRequest) getRID() string { return r.RID }

type BuyVoiceTimeRequest struct {
	Type          string `json:"_type"`
	PaymentMethod string `json:"payment_method"`
	RID           string `json:"rid"`
}

func (r BuyVoiceTimeRequest) getRID() string { return r.RID }

type BuyRoomBoostRequest struct {
	Type          string `json:"_type"`
	PaymentMethod string `json:"payment_method"`
	Amount        int    `json:"amount"`
	RID           string `json:"rid"`
}

func (r BuyRoomBoostRequest) getRID() string { return r.RID }

type TipUserRequest struct {
	Type    string `json:"_type"`
	UserID  string `json:"user_id"`
	GoldBar string `json:"gold_bar"`
	RID     string `json:"rid"`
}

func (r TipUserRequest) getRID() string { return r.RID }

type SetOutfitRequest struct {
	Type   string `json:"_type"`
	Outfit []Item `json:"outfit"`
	RID    string `json:"rid"`
}

func (r SetOutfitRequest) getRID() string { return r.RID }

type BuyItemRequest struct {
	Type   string `json:"_type"`
	ItemID string `json:"item_id"`
	RID    string `json:"rid"`
}

func (r BuyItemRequest) getRID() string { return r.RID }

type GetRoomUsersRequest struct {
	Type string `json:"_type"`
	RID  string `json:"rid"`
}

func (r GetRoomUsersRequest) getRID() string { return r.RID }

type GetRoomUsersResponse struct {
	Type    string             `json:"_type"`
	Content []SessionUserEntry `json:"content"`
	RID     string             `json:"rid"`
}

type GetWalletRequest struct {
	Type string `json:"_type"`
	RID  string `json:"rid"`
}

func (r GetWalletRequest) getRID() string { return r.RID }

type GetWalletResponse struct {
	Type    string         `json:"_type"`
	Content []CurrencyItem `json:"content"`
	RID     string         `json:"rid"`
}

type GetRoomPrivilegeRequest struct {
	Type   string `json:"_type"`
	UserID string `json:"user_id"`
	RID    string `json:"rid"`
}

func (r GetRoomPrivilegeRequest) getRID() string { return r.RID }

type GetRoomPrivilegeResponse struct {
	Type    string          `json:"_type"`
	Content RoomPermissions `json:"content"`
	RID     string          `json:"rid"`
}

type CheckVoiceChatRequest struct {
	Type string `json:"_type"`
	RID  string `json:"rid"`
}

func (r CheckVoiceChatRequest) getRID() string { return r.RID }

type CheckVoiceChatResponse struct {
	Type         string            `json:"_type"`
	SecondsLeft  int               `json:"seconds_left"`
	AutoSpeakers []string          `json:"auto_speakers"`
	Users        map[string]string `json:"users"`
	RID          string            `json:"rid"`
}

type GetUserOutfitRequest struct {
	Type   string `json:"_type"`
	UserID string `json:"user_id"`
	RID    string `json:"rid"`
}

func (r GetUserOutfitRequest) getRID() string { return r.RID }

type GetUserOutfitResponse struct {
	Type   string `json:"_type"`
	Outfit []Item `json:"outfit"`
	RID    string `json:"rid"`
}

type GetBackpackRequest struct {
	Type   string `json:"_type"`
	UserID string `json:"user_id"`
	RID    string `json:"rid"`
}

func (r GetBackpackRequest) getRID() string { return r.RID }

type GetBackpackResponse struct {
	Type     string         `json:"_type"`
	Backpack map[string]int `json:"backpack"`
	RID      string         `json:"rid"`
}

type ChangeBackpackRequest struct {
	Type    string         `json:"_type"`
	UserID  string         `json:"user_id"`
	Changes map[string]int `json:"changes"`
	RID     string         `json:"rid"`
}

func (r ChangeBackpackRequest) getRID() string { return r.RID }

type GetConversationsRequest struct {
	Type      string  `json:"_type"`
	NotJoined bool    `json:"not_joined"`
	LastID    *string `json:"last_id,omitempty"`
	RID       string  `json:"rid"`
}

func (r GetConversationsRequest) getRID() string { return r.RID }

type GetConversationsResponse struct {
	Type          string         `json:"_type"`
	Conversations []Conversation `json:"conversations"`
	NotJoined     int            `json:"not_joined"`
	RID           string         `json:"rid"`
}

type Conversation struct {
	ID          string   `json:"id"`
	DidJoin     bool     `json:"did_join"`
	UnreadCount int      `json:"unread_count"`
	LastMessage *Message `json:"last_message,omitempty"`
	Muted       bool     `json:"muted"`
	MemberIDs   []string `json:"member_ids,omitempty"`
	Name        *string  `json:"name,omitempty"`
	OwnerID     *string  `json:"owner_id,omitempty"`
}

type GetMessagesRequest struct {
	Type           string `json:"_type"`
	ConversationID string `json:"conversation_id"`
	LastMessageID  string `json:"last_message_id"`
	RID            string `json:"rid"`
}

func (r GetMessagesRequest) getRID() string { return r.RID }

type GetMessagesResponse struct {
	Type     string    `json:"_type"`
	Messages []Message `json:"messages"`
	RID      string    `json:"rid"`
}

type Message struct {
	MessageID      string  `json:"message_id"`
	ConversationID string  `json:"conversation_id"`
	CreatedAt      *string `json:"createdAt,omitempty"`
	Content        string  `json:"content"`
	SenderID       string  `json:"sender_id"`
	Category       string  `json:"category"`
}

type GetInventoryRequest struct {
	Type string `json:"_type"`
	RID  string `json:"rid"`
}

func (r GetInventoryRequest) getRID() string { return r.RID }

type GetInventoryResponse struct {
	Type  string `json:"_type"`
	Items []Item `json:"items"`
	RID   string `json:"rid"`
}

type MessageMedia struct {
	Type                 string `json:"type"`
	Width                int    `json:"width"`
	Height               int    `json:"height"`
	MediaSizeInBytes     int    `json:"mediaSizeInBytes"`
	ThumbnailSizeInBytes int    `json:"thumbnailSizeInBytes"`
	ID                   string `json:"id,omitempty"`
	URL                  string `json:"url,omitempty"`
	ThumbnailURL         string `json:"thumbnailUrl,omitempty"`
}

type MessageMediaRequest struct {
	Type  string       `json:"_type"`
	Media MessageMedia `json:"media"`
	RID   string       `json:"rid"`
}

func (r MessageMediaRequest) getRID() string { return r.RID }

type MessageMediaResponse struct {
	Type               string       `json:"_type"`
	Media              MessageMedia `json:"media"`
	UploadURL          string       `json:"uploadUrl,omitempty"`
	ThumbnailUploadURL string       `json:"thumbnailUploadUrl,omitempty"`
	RID                string       `json:"rid"`
}

type BuyVoiceTimeResponse struct {
	Type   string `json:"_type"`
	Result string `json:"result"`
	RID    string `json:"rid"`
}

type BuyRoomBoostResponse struct {
	Type   string `json:"_type"`
	Result string `json:"result"`
	RID    string `json:"rid"`
}

type TipUserResponse struct {
	Type   string `json:"_type"`
	Result string `json:"result"`
	RID    string `json:"rid"`
}

type BuyItemResponse struct {
	Type   string `json:"_type"`
	Result string `json:"result"`
	RID    string `json:"rid"`
}
