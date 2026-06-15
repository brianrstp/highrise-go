package highrise

import (
	"bytes"
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
	if bytes.Contains(data, []byte(`"entity_id"`)) {
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
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		if _, hasID := m["id"]; hasID {
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

	if raw[0] == nil {
		return fmt.Errorf("UserVoiceStatus: user field cannot be null")
	}
	userBytes, err := json.Marshal(raw[0])
	if err != nil {
		return err
	}
	if err := json.Unmarshal(userBytes, &s.User); err != nil {
		return err
	}

	if raw[1] == nil {
		return fmt.Errorf("UserVoiceStatus: status field cannot be null")
	}
	if status, ok := raw[1].(string); ok {
		s.Status = status
	} else {
		return fmt.Errorf("UserVoiceStatus: status must be a string")
	}
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

// SessionUserEntry represents a user currently in the session/room
type SessionUserEntry struct {
	User     User             `json:"user"`
	Position PositionOrAnchor `json:"position"`
}

func (e *SessionUserEntry) UnmarshalJSON(data []byte) error {
	var obj struct {
		User     User             `json:"user"`
		Position PositionOrAnchor `json:"position"`
	}
	if err := json.Unmarshal(data, &obj); err == nil && obj.User.ID != "" {
		e.User = obj.User
		e.Position = obj.Position
		return nil
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil || len(arr) < 2 {
		return fmt.Errorf("SessionUserEntry: expected object or array of [user, position], got %s", string(data))
	}

	if err := json.Unmarshal(arr[0], &e.User); err != nil {
		return err
	}
	if err := json.Unmarshal(arr[1], &e.Position); err != nil {
		return err
	}
	return nil
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

// ChatEvent is sent when a user sends a chat message
type ChatEvent struct {
	Type    string `json:"_type"`
	User    User   `json:"user"`
	Message string `json:"message"`
	Whisper bool   `json:"whisper"`
}

// EmoteEvent is sent when a user performs an emote
type EmoteEvent struct {
	Type     string `json:"_type"`
	User     User   `json:"user"`
	EmoteID  string `json:"emote_id"`
	Receiver *User  `json:"receiver,omitempty"`
}

// ReactionEvent is sent when a user reacts to another user
type ReactionEvent struct {
	Type     string `json:"_type"`
	User     User   `json:"user"`
	Reaction string `json:"reaction"`
	Receiver User   `json:"receiver"`
}

// UserJoinedEvent is sent when a user joins the room
type UserJoinedEvent struct {
	Type     string           `json:"_type"`
	User     User             `json:"user"`
	Position PositionOrAnchor `json:"position"`
}

// UserLeftEvent is sent when a user leaves the room
type UserLeftEvent struct {
	Type string `json:"_type"`
	User User   `json:"user"`
}

// UserMovedEvent is sent when a user moves to a new position
type UserMovedEvent struct {
	Type     string           `json:"_type"`
	User     User             `json:"user"`
	Position PositionOrAnchor `json:"position"`
}

// TipReactionEvent is sent when a user tips another user
type TipReactionEvent struct {
	Type     string  `json:"_type"`
	Sender   User    `json:"sender"`
	Receiver User    `json:"receiver"`
	Item     TipItem `json:"item"`
}

// ChannelEvent is sent when a channel message is received
type ChannelEvent struct {
	Type     string   `json:"_type"`
	SenderID string   `json:"sender_id"`
	Message  string   `json:"msg"`
	Tags     []string `json:"tags"`
}

// VoiceEvent is sent when voice chat status changes
type VoiceEvent struct {
	Type        string            `json:"_type"`
	Users       []UserVoiceStatus `json:"users"`
	SecondsLeft int               `json:"seconds_left"`
}

// MessageEvent is sent when a conversation message is received
type MessageEvent struct {
	Type              string `json:"_type"`
	UserID            string `json:"user_id"`
	ConversationID    string `json:"conversation_id"`
	IsNewConversation bool   `json:"is_new_conversation"`
}

// RoomModeratedEvent is sent when a moderation action occurs
type RoomModeratedEvent struct {
	Type           string `json:"_type"`
	ModeratorID    string `json:"moderatorId"`
	TargetUserID   string `json:"targetUserId"`
	ModerationType string `json:"moderationType"`
	Duration       *int   `json:"duration,omitempty"`
}

// --- Requests ---

// ChatRequest is sent to send a chat message
type ChatRequest struct {
	Type          string  `json:"_type"`
	Message       string  `json:"message"`
	WhisperTarget *string `json:"whisper_target_id"`
	RID           string  `json:"rid"`
}

func (r ChatRequest) getRID() string { return r.RID }

// EmoteRequest is sent to perform an emote
type EmoteRequest struct {
	Type         string  `json:"_type"`
	EmoteID      string  `json:"emote_id"`
	TargetUserID *string `json:"target_user_id"`
	RID          string  `json:"rid"`
}

func (r EmoteRequest) getRID() string { return r.RID }

// ReactionRequest is sent to react to another user
type ReactionRequest struct {
	Type         string `json:"_type"`
	Reaction     string `json:"reaction"`
	TargetUserID string `json:"target_user_id"`
	RID          string `json:"rid"`
}

func (r ReactionRequest) getRID() string { return r.RID }

// IndicatorRequest is sent to set the user's indicator icon
type IndicatorRequest struct {
	Type string  `json:"_type"`
	Icon *string `json:"icon"`
	RID  string  `json:"rid"`
}

func (r IndicatorRequest) getRID() string { return r.RID }

// ChannelRequest is sent to broadcast a channel message
type ChannelRequest struct {
	Type    string   `json:"_type"`
	Message string   `json:"message"`
	Tags    []string `json:"tags"`
	RID     string   `json:"rid"`
}

func (r ChannelRequest) getRID() string { return r.RID }

// FloorHitRequest is sent to walk to a position
type FloorHitRequest struct {
	Type        string   `json:"_type"`
	Destination Position `json:"destination"`
	RID         string   `json:"rid"`
}

func (r FloorHitRequest) getRID() string { return r.RID }

// AnchorHitRequest is sent to walk to an anchor point
type AnchorHitRequest struct {
	Type   string         `json:"_type"`
	Anchor AnchorPosition `json:"anchor"`
	RID    string         `json:"rid"`
}

func (r AnchorHitRequest) getRID() string { return r.RID }

// TeleportRequest is sent to teleport a user to a position
type TeleportRequest struct {
	Type        string   `json:"_type"`
	UserID      string   `json:"user_id"`
	Destination Position `json:"destination"`
	RID         string   `json:"rid"`
}

func (r TeleportRequest) getRID() string { return r.RID }

// ModerateRoomRequest is sent to moderate a user in the room
type ModerateRoomRequest struct {
	Type             string `json:"_type"`
	UserID           string `json:"user_id"`
	ModerationAction string `json:"moderation_action"`
	ActionLength     *int   `json:"action_length,omitempty"`
	RID              string `json:"rid"`
}

func (r ModerateRoomRequest) getRID() string { return r.RID }

// ChangeRoomPrivilegeRequest is sent to change a user's room permissions
type ChangeRoomPrivilegeRequest struct {
	Type        string          `json:"_type"`
	UserID      string          `json:"user_id"`
	Permissions RoomPermissions `json:"permissions"`
	RID         string          `json:"rid"`
}

func (r ChangeRoomPrivilegeRequest) getRID() string { return r.RID }

// MoveUserToRoomRequest is sent to move a user to another room
type MoveUserToRoomRequest struct {
	Type   string `json:"_type"`
	UserID string `json:"user_id"`
	RoomID string `json:"room_id"`
	RID    string `json:"rid"`
}

func (r MoveUserToRoomRequest) getRID() string { return r.RID }

// InviteSpeakerRequest is sent to invite a user as a speaker
type InviteSpeakerRequest struct {
	Type   string `json:"_type"`
	UserID string `json:"user_id"`
	RID    string `json:"rid"`
}

func (r InviteSpeakerRequest) getRID() string { return r.RID }

// RemoveSpeakerRequest is sent to remove a user as a speaker
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

// SendMessageRequest is sent to send a direct message
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

// SendBulkMessageRequest is sent to send a bulk message to multiple users
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

// LeaveConversationRequest is sent to leave a conversation
type LeaveConversationRequest struct {
	Type           string `json:"_type"`
	ConversationID string `json:"conversation_id"`
	RID            string `json:"rid"`
}

func (r LeaveConversationRequest) getRID() string { return r.RID }

// BuyVoiceTimeRequest is sent to purchase voice time
type BuyVoiceTimeRequest struct {
	Type          string `json:"_type"`
	PaymentMethod string `json:"payment_method"`
	RID           string `json:"rid"`
}

func (r BuyVoiceTimeRequest) getRID() string { return r.RID }

// BuyRoomBoostRequest is sent to purchase a room boost
type BuyRoomBoostRequest struct {
	Type          string `json:"_type"`
	PaymentMethod string `json:"payment_method"`
	Amount        int    `json:"amount"`
	RID           string `json:"rid"`
}

func (r BuyRoomBoostRequest) getRID() string { return r.RID }

// TipUserRequest is sent to tip a user with gold bars
type TipUserRequest struct {
	Type    string `json:"_type"`
	UserID  string `json:"user_id"`
	GoldBar string `json:"gold_bar"`
	RID     string `json:"rid"`
}

func (r TipUserRequest) getRID() string { return r.RID }

// SetOutfitRequest is sent to set the bot's outfit
type SetOutfitRequest struct {
	Type   string `json:"_type"`
	Outfit []Item `json:"outfit"`
	RID    string `json:"rid"`
}

func (r SetOutfitRequest) getRID() string { return r.RID }

// BuyItemRequest is sent to purchase an item
type BuyItemRequest struct {
	Type   string `json:"_type"`
	ItemID string `json:"item_id"`
	RID    string `json:"rid"`
}

func (r BuyItemRequest) getRID() string { return r.RID }

// GetRoomUsersRequest is sent to request the room user list
type GetRoomUsersRequest struct {
	Type string `json:"_type"`
	RID  string `json:"rid"`
}

func (r GetRoomUsersRequest) getRID() string { return r.RID }

// GetRoomUsersResponse contains the room user list
type GetRoomUsersResponse struct {
	Type    string             `json:"_type"`
	Content []SessionUserEntry `json:"content"`
	RID     string             `json:"rid"`
}

// GetWalletRequest is sent to request the wallet balance
type GetWalletRequest struct {
	Type string `json:"_type"`
	RID  string `json:"rid"`
}

func (r GetWalletRequest) getRID() string { return r.RID }

// GetWalletResponse contains the wallet balance
type GetWalletResponse struct {
	Type    string         `json:"_type"`
	Content []CurrencyItem `json:"content"`
	RID     string         `json:"rid"`
}

// GetRoomPrivilegeRequest is sent to request a user's room privileges
type GetRoomPrivilegeRequest struct {
	Type   string `json:"_type"`
	UserID string `json:"user_id"`
	RID    string `json:"rid"`
}

func (r GetRoomPrivilegeRequest) getRID() string { return r.RID }

// GetRoomPrivilegeResponse contains a user's room permissions
type GetRoomPrivilegeResponse struct {
	Type    string          `json:"_type"`
	Content RoomPermissions `json:"content"`
	RID     string          `json:"rid"`
}

// CheckVoiceChatRequest is sent to check voice chat status
type CheckVoiceChatRequest struct {
	Type string `json:"_type"`
	RID  string `json:"rid"`
}

func (r CheckVoiceChatRequest) getRID() string { return r.RID }

// CheckVoiceChatResponse contains voice chat status information
type CheckVoiceChatResponse struct {
	Type         string            `json:"_type"`
	SecondsLeft  int               `json:"seconds_left"`
	AutoSpeakers []string          `json:"auto_speakers"`
	Users        map[string]string `json:"users"`
	RID          string            `json:"rid"`
}

// GetUserOutfitRequest is sent to request a user's outfit
type GetUserOutfitRequest struct {
	Type   string `json:"_type"`
	UserID string `json:"user_id"`
	RID    string `json:"rid"`
}

func (r GetUserOutfitRequest) getRID() string { return r.RID }

// GetUserOutfitResponse contains a user's outfit
type GetUserOutfitResponse struct {
	Type   string `json:"_type"`
	Outfit []Item `json:"outfit"`
	RID    string `json:"rid"`
}

// GetBackpackRequest is sent to request a user's backpack
type GetBackpackRequest struct {
	Type   string `json:"_type"`
	UserID string `json:"user_id"`
	RID    string `json:"rid"`
}

func (r GetBackpackRequest) getRID() string { return r.RID }

// GetBackpackResponse contains a user's backpack contents
type GetBackpackResponse struct {
	Type     string         `json:"_type"`
	Backpack map[string]int `json:"backpack"`
	RID      string         `json:"rid"`
}

// ChangeBackpackRequest is sent to modify a user's backpack
type ChangeBackpackRequest struct {
	Type    string         `json:"_type"`
	UserID  string         `json:"user_id"`
	Changes map[string]int `json:"changes"`
	RID     string         `json:"rid"`
}

func (r ChangeBackpackRequest) getRID() string { return r.RID }

// GetConversationsRequest is sent to request the conversation list
type GetConversationsRequest struct {
	Type      string  `json:"_type"`
	NotJoined bool    `json:"not_joined"`
	LastID    *string `json:"last_id,omitempty"`
	RID       string  `json:"rid"`
}

func (r GetConversationsRequest) getRID() string { return r.RID }

// GetConversationsResponse contains the conversation list
type GetConversationsResponse struct {
	Type          string         `json:"_type"`
	Conversations []Conversation `json:"conversations"`
	NotJoined     int            `json:"not_joined"`
	RID           string         `json:"rid"`
}

// Conversation represents a DM or group conversation
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

// GetMessagesRequest is sent to request messages in a conversation
type GetMessagesRequest struct {
	Type           string `json:"_type"`
	ConversationID string `json:"conversation_id"`
	LastMessageID  string `json:"last_message_id"`
	RID            string `json:"rid"`
}

func (r GetMessagesRequest) getRID() string { return r.RID }

// GetMessagesResponse contains the message list for a conversation
type GetMessagesResponse struct {
	Type     string    `json:"_type"`
	Messages []Message `json:"messages"`
	RID      string    `json:"rid"`
}

// Message represents a single message in a conversation
type Message struct {
	MessageID      string  `json:"message_id"`
	ConversationID string  `json:"conversation_id"`
	CreatedAt      *string `json:"createdAt,omitempty"`
	Content        string  `json:"content"`
	SenderID       string  `json:"sender_id"`
	Category       string  `json:"category"`
}

// GetInventoryRequest is sent to request the bot's inventory
type GetInventoryRequest struct {
	Type string `json:"_type"`
	RID  string `json:"rid"`
}

func (r GetInventoryRequest) getRID() string { return r.RID }

// GetInventoryResponse contains the bot's inventory items
type GetInventoryResponse struct {
	Type  string `json:"_type"`
	Items []Item `json:"items"`
	RID   string `json:"rid"`
}

// MessageMedia represents media content attached to a message
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

// MessageMediaRequest is sent to upload message media
type MessageMediaRequest struct {
	Type  string       `json:"_type"`
	Media MessageMedia `json:"media"`
	RID   string       `json:"rid"`
}

func (r MessageMediaRequest) getRID() string { return r.RID }

// MessageMediaResponse contains upload URLs for message media
type MessageMediaResponse struct {
	Type               string       `json:"_type"`
	Media              MessageMedia `json:"media"`
	UploadURL          string       `json:"uploadUrl,omitempty"`
	ThumbnailUploadURL string       `json:"thumbnailUploadUrl,omitempty"`
	RID                string       `json:"rid"`
}

// BuyVoiceTimeResponse contains the voice time purchase result
type BuyVoiceTimeResponse struct {
	Type   string `json:"_type"`
	Result string `json:"result"`
	RID    string `json:"rid"`
}

// BuyRoomBoostResponse contains the room boost purchase result
type BuyRoomBoostResponse struct {
	Type   string `json:"_type"`
	Result string `json:"result"`
	RID    string `json:"rid"`
}

// TipUserResponse contains the tip result
type TipUserResponse struct {
	Type   string `json:"_type"`
	Result string `json:"result"`
	RID    string `json:"rid"`
}

// BuyItemResponse contains the item purchase result
type BuyItemResponse struct {
	Type   string `json:"_type"`
	Result string `json:"result"`
	RID    string `json:"rid"`
}
