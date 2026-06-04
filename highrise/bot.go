package highrise

import "context"

// BotHandler is the interface that bot implementations must implement.
// All methods must be defined; embed Bot to get no-op defaults for all handlers.
type BotHandler interface {
	// OnStart is called when the connection is established and session metadata is received
	OnStart(ctx context.Context, session *SessionMetadata)
	// OnChat is called when a user sends a public chat message
	OnChat(ctx context.Context, user User, message string)
	// OnUserJoin is called when a user joins the room
	OnUserJoin(ctx context.Context, user User, position PositionOrAnchor)
	// OnUserLeave is called when a user leaves the room
	OnUserLeave(ctx context.Context, user User)
	// OnUserMove is called when a user moves in the room
	OnUserMove(ctx context.Context, user User, position PositionOrAnchor)
	// OnEmote is called when a user sends an emote
	OnEmote(ctx context.Context, user User, emoteID string, receiver *User)
	// OnReaction is called when a user reacts
	OnReaction(ctx context.Context, user User, reaction string, receiver User)
	// OnTip is called when a tip is sent
	OnTip(ctx context.Context, sender, receiver User, item *TipItem)
	// OnWhisper is called when a user sends a whisper
	OnWhisper(ctx context.Context, user User, message string)
	// OnVoiceChange is called when voice chat status changes
	OnVoiceChange(ctx context.Context, users []UserVoiceStatus, secondsLeft int)
	// OnChannel is called when a channel message is received
	OnChannel(ctx context.Context, senderID, message string, tags []string)
	// OnMessage is called when a inbox message is received
	OnMessage(ctx context.Context, userID, conversationID string, isNewConversation bool)
	// OnModerate is called when a room moderation action occurs
	OnModerate(ctx context.Context, moderatorID, targetUserID, moderationType string, duration *int)
}

// The following interfaces are optional: implement only the ones you need.
// If your bot embeds Bot, all handlers already have no-op defaults.

type HasBeforeStart interface {
	BeforeStart(ctx context.Context)
}

// HasOnStart is called when the connection is established and session metadata is received.
type HasOnStart interface {
	OnStart(ctx context.Context, session *SessionMetadata)
}

// HasOnChat is called when a user sends a public chat message.
type HasOnChat interface {
	OnChat(ctx context.Context, user User, message string)
}

// HasOnWhisper is called when a user sends a private whisper.
type HasOnWhisper interface {
	OnWhisper(ctx context.Context, user User, message string)
}

// HasOnEmote is called when a user sends an emote.
type HasOnEmote interface {
	OnEmote(ctx context.Context, user User, emoteID string, receiver *User)
}

// HasOnReaction is called when a user reacts to another user.
type HasOnReaction interface {
	OnReaction(ctx context.Context, user User, reaction string, receiver User)
}

// HasOnUserJoin is called when a user enters the room.
type HasOnUserJoin interface {
	OnUserJoin(ctx context.Context, user User, position PositionOrAnchor)
}

// HasOnUserLeave is called when a user leaves the room.
type HasOnUserLeave interface {
	OnUserLeave(ctx context.Context, user User)
}

// HasOnUserMove is called when a user changes position in the room.
type HasOnUserMove interface {
	OnUserMove(ctx context.Context, user User, position PositionOrAnchor)
}

// HasOnTip is called when a user sends a tip (currency or item).
type HasOnTip interface {
	OnTip(ctx context.Context, sender, receiver User, item *TipItem)
}

// HasOnVoiceChange is called when voice chat status changes.
type HasOnVoiceChange interface {
	OnVoiceChange(ctx context.Context, users []UserVoiceStatus, secondsLeft int)
}

// HasOnChannel is called when a hidden channel message is received.
type HasOnChannel interface {
	OnChannel(ctx context.Context, senderID, message string, tags []string)
}

// HasOnMessage is called when an inbox message is received.
type HasOnMessage interface {
	OnMessage(ctx context.Context, userID, conversationID string, isNewConversation bool)
}

// HasOnModerate is called when a room moderation action occurs.
type HasOnModerate interface {
	OnModerate(ctx context.Context, moderatorID, targetUserID, moderationType string, duration *int)
}

// HasOnError is called when the server returns an error message.
type HasOnError interface {
	OnError(ctx context.Context, err Error)
}

// HasOnConnectionChange is called when the connection state changes.
type HasOnConnectionChange interface {
	OnConnectionChange(ctx context.Context, state ConnectionState)
}

// Bot is the base bot struct that users should embed in their bot.
// It provides default no-op implementations for all event handlers
// so you only need to override the methods you care about.
type Bot struct {
	// Highrise provides access to all bot actions (chat, walk, teleport, etc.).
	// Set via SetHighrise before calling client.Run.
	Highrise *Highrise
}

func (b *Bot) SetHighrise(h *Highrise) { b.Highrise = h }

func (b *Bot) BeforeStart(ctx context.Context)                                                      {}
func (b *Bot) OnStart(ctx context.Context, session *SessionMetadata)                                {}
func (b *Bot) OnChat(ctx context.Context, user User, message string)                                {}
func (b *Bot) OnWhisper(ctx context.Context, user User, message string)                             {}
func (b *Bot) OnUserJoin(ctx context.Context, user User, position PositionOrAnchor)                 {}
func (b *Bot) OnUserLeave(ctx context.Context, user User)                                           {}
func (b *Bot) OnUserMove(ctx context.Context, user User, position PositionOrAnchor)                 {}
func (b *Bot) OnEmote(ctx context.Context, user User, emoteID string, receiver *User)               {}
func (b *Bot) OnReaction(ctx context.Context, user User, reaction string, receiver User)            {}
func (b *Bot) OnTip(ctx context.Context, sender, receiver User, item *TipItem)                      {}
func (b *Bot) OnVoiceChange(ctx context.Context, users []UserVoiceStatus, secondsLeft int)          {}
func (b *Bot) OnChannel(ctx context.Context, senderID, message string, tags []string)               {}
func (b *Bot) OnMessage(ctx context.Context, userID, conversationID string, isNewConversation bool) {}
func (b *Bot) OnModerate(ctx context.Context, moderatorID, targetUserID, moderationType string, duration *int) {
}
func (b *Bot) OnConnectionChange(ctx context.Context, state ConnectionState) {}
