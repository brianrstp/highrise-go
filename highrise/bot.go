package highrise

import "context"

// BotHandler is the interface that bot implementations should implement
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

// Optional interfaces that bots can implement for selective event handling

type HasBeforeStart interface {
	BeforeStart(ctx context.Context)
}

type HasOnStart interface {
	OnStart(ctx context.Context, session *SessionMetadata)
}

type HasOnChat interface {
	OnChat(ctx context.Context, user User, message string)
}

type HasOnWhisper interface {
	OnWhisper(ctx context.Context, user User, message string)
}

type HasOnEmote interface {
	OnEmote(ctx context.Context, user User, emoteID string, receiver *User)
}

type HasOnReaction interface {
	OnReaction(ctx context.Context, user User, reaction string, receiver User)
}

type HasOnUserJoin interface {
	OnUserJoin(ctx context.Context, user User, position PositionOrAnchor)
}

type HasOnUserLeave interface {
	OnUserLeave(ctx context.Context, user User)
}

type HasOnUserMove interface {
	OnUserMove(ctx context.Context, user User, position PositionOrAnchor)
}

type HasOnTip interface {
	OnTip(ctx context.Context, sender, receiver User, item *TipItem)
}

type HasOnVoiceChange interface {
	OnVoiceChange(ctx context.Context, users []UserVoiceStatus, secondsLeft int)
}

type HasOnChannel interface {
	OnChannel(ctx context.Context, senderID, message string, tags []string)
}

type HasOnMessage interface {
	OnMessage(ctx context.Context, userID, conversationID string, isNewConversation bool)
}

type HasOnModerate interface {
	OnModerate(ctx context.Context, moderatorID, targetUserID, moderationType string, duration *int)
}

type HasOnError interface {
	OnError(ctx context.Context, err Error)
}

// Bot is the base bot struct that users should embed
// It provides default no-op implementations for all handlers
type Bot struct {
	Highrise *Highrise
}

func (b *Bot) SetHighrise(h *Highrise) {
	b.Highrise = h
}

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
