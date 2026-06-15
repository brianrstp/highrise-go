# highrise-go

Go SDK for the [Highrise Bot API](https://highrise.game). Provides a WebSocket client for real-time bot actions (chat, walk, teleport, etc.), event handler interfaces, rate limiting, middleware system, and a REST client for the Highrise WebAPI.

## Features

- **WebSocket Client** - Real-time connection to Highrise rooms with auto-reconnect (exponential backoff + jitter)
- **30+ Bot Actions** - Chat, whisper, walk, teleport, emote, moderate, tip, buy items, manage outfits, and more
- **Event Handlers** - 14 event handlers (chat, join, leave, move, emote, reaction, tip, whisper, voice, channel, DM, moderation)
- **Middleware System** - Wrap every event handler with custom logic (logging, metrics, auth)
- **Rate Limiting** - Sliding window rate limiter automatically applied from server-sent limits
- **REST API Client** - Access Highrise WebAPI for users, rooms, items, grab bags, and posts
- **Context-Aware** - All methods accept `context.Context` for cancellation and timeouts
- **Connection State** - Monitor connection status (Disconnected, Connecting, Connected, Reconnecting)
- **Metrics** - Counters for events, actions, errors, and reconnects
- **Worker Pool** - Bounded goroutine pool (max 64) for event handling
- **JSON Pool** - Buffer reuse via `sync.Pool` to reduce memory allocations

## Installation

```bash
go get github.com/brianrstp/highrise-go
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/brianrstp/highrise-go/highrise"
)

type MyBot struct {
    highrise.Bot
}

func (b *MyBot) OnStart(ctx context.Context, session *highrise.SessionMetadata) {
    log.Printf("Bot started! Room: %s", session.RoomInfo.RoomName)
    b.Highrise.Chat(ctx, "Hello everyone!")
}

func (b *MyBot) OnChat(ctx context.Context, user highrise.User, message string) {
    if message == "!ping" {
        b.Highrise.Reply(ctx, user, "pong!")
    }
}

func main() {
    bot := &MyBot{}
    client := highrise.NewClient(bot)
    bot.SetHighrise(client.Highrise())

    ctx := context.Background()
    if err := client.Run(ctx, os.Args[1], os.Args[2]); err != nil {
        log.Fatal(err)
    }
}
```

Run it:

```bash
go run . <room_id> <api_token>
```

## Architecture

```
┌──────────────────────────────────────────────────────┐
│  Your Bot (embed highrise.Bot)                       │
│  Override OnChat, OnUserJoin, etc.                   │
├──────────────────────────────────────────────────────┤
│  Client                                              │
│  ├── WebSocket connection management                 │
│  ├── Auto-reconnect (exponential backoff)            │
│  ├── Event routing → BotHandler methods              │
│  ├── Middleware chain                                 │
│  ├── Rate limiter (auto-applied from server)         │
│  ├── Request-response correlation (RID)              │
│  └── Worker pool (bounded goroutines)                │
├──────────────────────────────────────────────────────┤
│  Highrise (Action Facade)                            │
│  ├── Chat, Whisper, Emote, React                     │
│  ├── Walk, Teleport, Anchor                          │
│  ├── Moderate, ChangePrivilege, MoveUser             │
│  ├── GetRoomUsers, GetWallet, GetInventory           │
│  ├── Tip, BuyItem, BuyVoiceTime, BuyRoomBoost        │
│  ├── SendMessage, GetConversations, GetMessages      │
│  └── SetOutfit, GetUserOutfit, GetBackpack, etc.     │
├──────────────────────────────────────────────────────┤
│  WebAPI (REST Client)                                │
│  ├── GetUser / GetUsers                              │
│  ├── GetRoom / GetRooms                              │
│  ├── GetItem / GetItems                              │
│  ├── GetGrab / GetGrabs                              │
│  └── GetPost / GetPosts                              │
└──────────────────────────────────────────────────────┘
```

## Event Handlers

### Bot Struct (Embed for Default No-Op)

Embed `highrise.Bot` in your bot struct. All event handlers have default no-op implementations, so you only need to override the methods you care about.

```go
type MyBot struct {
    highrise.Bot
}

// Override only what you need
func (b *MyBot) OnChat(ctx context.Context, user highrise.User, message string) {
    // handle chat
}
```

### All Event Handlers

| Event | Method | Description |
|-------|--------|-------------|
| Connect | `BeforeStart(ctx)` | Called before connection (optional, implement `HasBeforeStart`) |
| Session | `OnStart(ctx, *SessionMetadata)` | Connection established, session metadata received |
| Chat | `OnChat(ctx, User, message)` | User sent a public chat message |
| Whisper | `OnWhisper(ctx, User, message)` | User sent a private whisper |
| Emote | `OnEmote(ctx, User, emoteID, *User)` | User performed an emote |
| Reaction | `OnReaction(ctx, User, reaction, User)` | User reacted to another user |
| User Join | `OnUserJoin(ctx, User, PositionOrAnchor)` | User entered the room |
| User Leave | `OnUserLeave(ctx, User)` | User left the room |
| User Move | `OnUserMove(ctx, User, PositionOrAnchor)` | User changed position |
| Tip | `OnTip(ctx, sender, receiver, *TipItem)` | A tip was sent |
| Voice | `OnVoiceChange(ctx, []UserVoiceStatus, secondsLeft)` | Voice chat status changed |
| Channel | `OnChannel(ctx, senderID, message, tags)` | Channel message received |
| DM | `OnMessage(ctx, userID, convID, isNew)` | Inbox message received |
| Moderate | `OnModerate(ctx, modID, targetID, modType, *duration)` | A moderation action occurred |
| Error | `OnError(ctx, Error)` | Server sent an error message |
| Any Event | `OnAnyEvent(ctx, eventType, data)` | Every event (raw JSON payload) |
| State | `OnConnectionChange(ctx, ConnectionState)` | Connection state changed |

### How to Implement

Implement the `Has*` interface to register a handler:

```go
// Implement directly on your struct (auto-detected)
func (b *MyBot) OnChat(ctx context.Context, user highrise.User, message string) { ... }

// Or via interface assertion
var _ highrise.HasOnChat = (*MyBot)(nil)
```

## Bot Actions

All actions are accessed via `b.Highrise` (or `client.Highrise()`).

### Chat & Messaging

```go
// Send a public chat message
b.Highrise.Chat(ctx, "Hello!")

// Send a whisper to a specific user
b.Highrise.SendWhisper(ctx, userID, "Secret message")

// Reply to a user's chat with @username prefix
b.Highrise.Reply(ctx, user, "pong!")

// Reply via whisper
b.Highrise.WhisperReply(ctx, user, "Secret reply")

// Set indicator icon (nil to remove)
icon := "star"
b.Highrise.SetIndicator(ctx, &icon)
b.Highrise.SetIndicator(ctx, nil)

// Send a channel message with tags
b.Highrise.SendChannel(ctx, "channel message", []string{"tag1", "tag2"})
```

### Movement

```go
// Walk to a position
b.Highrise.WalkTo(ctx, highrise.Position{X: 5, Y: 0, Z: 3, Facing: "FrontRight"})

// Walk to an anchor point
b.Highrise.WalkToAnchor(ctx, highrise.AnchorPosition{EntityID: "ent_123", AnchorIx: 0})

// Teleport a user
b.Highrise.Teleport(ctx, userID, highrise.Position{X: 0, Y: 0, Z: 0, Facing: "Front"})
```

### Emotes & Reactions

```go
// Send an emote (targetUserID is optional)
b.Highrise.SendEmote(ctx, "emoji_laugh", nil)
b.Highrise.SendEmote(ctx, "emoji_wave", &targetUserID)

// React to a user
b.Highrise.React(ctx, "heart", targetUserID)
```

### Moderation

```go
// Kick a user
b.Highrise.ModerateRoom(ctx, userID, "kick", nil)

// Ban a user (duration in seconds, nil = permanent)
duration := 3600 // 1 hour
b.Highrise.ModerateRoom(ctx, userID, "ban", &duration)

// Mute a user
duration = 300 // 5 minutes
b.Highrise.ModerateRoom(ctx, userID, "mute", &duration)

// Unban a user
b.Highrise.ModerateRoom(ctx, userID, "unban", nil)

// Grant/revoke moderator privileges
modTrue := true
modFalse := false
b.Highrise.ChangeRoomPrivilege(ctx, userID, highrise.RoomPermissions{Moderator: &modTrue})
b.Highrise.ChangeRoomPrivilege(ctx, userID, highrise.RoomPermissions{Moderator: &modFalse})

// Move a user to another room
b.Highrise.MoveUserToRoom(ctx, userID, roomID)
```

### Voice Chat

```go
// Check voice chat status
status, err := b.Highrise.GetVoiceStatus(ctx)
// status.SecondsLeft, status.Users, status.AutoSpeakers

// Invite/remove a speaker
b.Highrise.InviteSpeaker(ctx, userID)
b.Highrise.RemoveSpeaker(ctx, userID)

// Purchase voice time
result, err := b.Highrise.BuyVoiceTime(ctx)
```

### Economy

```go
// Check wallet balance
wallet, err := b.Highrise.GetWallet(ctx)
for _, item := range wallet {
    fmt.Printf("%s: %d\n", item.Type, item.Amount)
}

// Tip a user
result, err := b.Highrise.TipUser(ctx, userID, "gold_bar_10")

// Buy an item from the shop
result, err := b.Highrise.BuyItem(ctx, itemID)

// Purchase a room boost
result, err := b.Highrise.BuyRoomBoost(ctx, 1)
```

### Items & Outfit

```go
// Get the bot's inventory
items, err := b.Highrise.GetInventory(ctx)

// Set the bot's outfit
b.Highrise.SetOutfit(ctx, []highrise.Item{
    {Type: "shirt", ID: "item_123", Amount: 1, AccountBound: false},
})

// Get a user's outfit
outfit, err := b.Highrise.GetUserOutfit(ctx, userID)

// Get the bot's own outfit
outfit, err := b.Highrise.GetMyOutfit(ctx)

// Get a user's backpack
bp, err := b.Highrise.GetBackpack(ctx, userID)
// bp = map[string]int{"gold_bar_10": 5}

// Modify a user's backpack
b.Highrise.ChangeBackpack(ctx, userID, map[string]int{
    "gold_bar_10": -1, // decrease by 1
})
```

### Room Info

```go
// Get all users in the room
users, err := b.Highrise.GetRoomUsers(ctx)
for _, entry := range users {
    fmt.Printf("%s at (%.1f, %.1f, %.1f)\n",
        entry.User.Username,
        entry.Position.Position.X,
        entry.Position.Position.Y,
        entry.Position.Position.Z,
    )
}

// Get a user's room privileges
perms, err := b.Highrise.GetRoomPrivilege(ctx, userID)
// perms.Moderator, perms.Designer
```

### Direct Messages (DM)

```go
// Get conversation list
convs, notJoined, err := b.Highrise.GetConversations(ctx, false, nil)

// Send a text message
b.Highrise.SendMessage(ctx, convID, "Hello!", "text", nil, nil, nil)

// Send a message with media
b.Highrise.SendMessage(ctx, convID, "Check this out!", "text", nil, nil, &mediaID)

// Send a bulk message to multiple users
b.Highrise.SendBulkMessage(ctx, []string{"uid1", "uid2"}, "Broadcast!", "text", nil, nil)

// Get messages in a conversation
msgs, err := b.Highrise.GetMessages(ctx, convID, lastMessageID)

// Leave a conversation
b.Highrise.LeaveConversation(ctx, convID)

// Upload media for use in a message
media := highrise.MessageMedia{Type: "image", Width: 100, Height: 100}
resp, err := b.Highrise.MessageMediaUpload(ctx, media)
// resp.UploadURL, resp.ThumbnailUploadURL
```

## Client Options

```go
// Custom WebSocket URL
client := highrise.NewClient(bot, highrise.WithURL("wss://custom.url"))

// Custom logger
client := highrise.NewClient(bot, highrise.WithLogger(myLogger))

// Custom SDK version (sent to server)
client := highrise.NewClient(bot, highrise.WithSDKVersion("1.0.0"))

// Custom event subscription
client := highrise.NewClient(bot, highrise.WithEvents("chat,user_joined,user_left"))

// Custom action timeout (default: 10s, 0 to disable)
client := highrise.NewClient(bot, highrise.WithActionTimeout(30*time.Second))
```

## Middleware

Middleware wraps every event handler call. Useful for logging, metrics, or auth checks.

```go
// Timing middleware
client.Use(func(next func()) {
    start := time.Now()
    next()
    log.Printf("Event handled in %v", time.Since(start))
})

// Logging middleware
client.Use(func(next func()) {
    log.Println("Before event handler")
    next()
    log.Println("After event handler")
})

// Multiple middleware (executed in order added)
client.Use(middleware1)
client.Use(middleware2)
// Execution order: middleware1 → middleware2 → actual handler
```

## Connection State

```go
// Via interface
func (b *MyBot) OnConnectionChange(ctx context.Context, state highrise.ConnectionState) {
    switch state {
    case highrise.StateConnected:
        log.Println("Connected!")
    case highrise.StateDisconnected:
        log.Println("Disconnected")
    case highrise.StateReconnecting:
        log.Println("Reconnecting...")
    }
}

// Via client methods
if client.IsConnected() { ... }
if client.IsStopped() { ... }
```

## Metrics

```go
metrics := client.Metrics()
// metrics["events"]     - total events received
// metrics["actions"]    - total actions sent
// metrics["errors"]     - total errors received
// metrics["reconnects"] - total reconnections
```

## Graceful Shutdown

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Handle OS signals
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
go func() {
    <-sigCh
    cancel()
}()

// Run blocks until context is cancelled
if err := client.Run(ctx, roomID, apiToken); err != nil {
    log.Fatal(err)
}
```

Or use `client.Stop()`:

```go
go client.Run(ctx, roomID, apiToken)
// ... later
client.Stop() // graceful shutdown
```

## Error Handling

```go
// ResponseError - error returned by the server
var respErr *highrise.ResponseError
if errors.As(err, &respErr) {
    log.Printf("Server error: %s", respErr.Message)
}

// ConnectionError - error from the WebSocket layer
var connErr *highrise.ConnectionError
if errors.As(err, &connErr) {
    log.Printf("Connection error [%s]: %v", connErr.ReqType, connErr.Err)
}
```

### Server Error with DoNotReconnect

If the server sends an error with `DoNotReconnect: true`, the client will automatically stop. Implement `HasOnError` to handle it:

```go
func (b *MyBot) OnError(ctx context.Context, err highrise.Error) {
    log.Printf("Error: %s (reconnect: %v)", err.Message, !err.DoNotReconnect)
}
```

## WebAPI (REST Client)

For read-only data without a WebSocket connection.

```go
api := highrise.NewWebAPI()

// Users
user, err := api.GetUser(ctx, userID)
users, err := api.GetUsers(ctx, highrise.UsersListParams{
    Username: "alice",
    Limit:    10,
})

// Rooms
room, err := api.GetRoom(ctx, roomID)
rooms, err := api.GetRooms(ctx, highrise.RoomsListParams{
    RoomName: "My Room",
    Limit:    20,
})

// Items
item, err := api.GetItem(ctx, itemID)
items, err := api.GetItems(ctx, highrise.ItemsListParams{
    Rarity:   "rare",
    Category: "top",
    Limit:    50,
})

// Grab Bags
grab, err := api.GetGrab(ctx, grabID)
grabs, err := api.GetGrabs(ctx, highrise.GrabsListParams{Limit: 10})

// Posts
post, err := api.GetPost(ctx, postID)
posts, err := api.GetPosts(ctx, highrise.PostsListParams{
    AuthorID: userID,
    Limit:    20,
})
```

### Pagination

All list endpoints support pagination:

```go
// First page
resp1, _ := api.GetRooms(ctx, highrise.RoomsListParams{Limit: 20})

// Next page
resp2, _ := api.GetRooms(ctx, highrise.RoomsListParams{
    Limit:       20,
    StartsAfter: resp1.LastID,
})

// Previous page
resp3, _ := api.GetRooms(ctx, highrise.RoomsListParams{
    Limit:      20,
    EndsBefore: resp1.FirstID,
})
```

## Model Types

### Core

```go
type User struct {
    ID       string `json:"id"`
    Username string `json:"username"`
}

type Position struct {
    X      float64 `json:"x"`
    Y      float64 `json:"y"`
    Z      float64 `json:"z"`
    Facing string  `json:"facing"`
}

type AnchorPosition struct {
    EntityID string `json:"entity_id"`
    AnchorIx int    `json:"anchor_ix"`
}

type PositionOrAnchor struct {
    Position       *Position
    AnchorPosition *AnchorPosition
}

type Item struct {
    Type          string `json:"type"`
    Amount        int    `json:"amount"`
    ID            string `json:"id"`
    AccountBound  bool   `json:"account_bound"`
    ActivePalette *int   `json:"active_palette,omitempty"`
}

type CurrencyItem struct {
    Type   string `json:"type"`
    Amount int    `json:"amount"`
}

type TipItem struct {
    CurrencyItem *CurrencyItem
    Item         *Item
}
```

### Polymorphic Types

Some types have custom JSON marshaling:

- **`PositionOrAnchor`** - Can be either a `Position` or `AnchorPosition` (determined by the presence of the `entity_id` field)
- **`TipItem`** - Can be either a `CurrencyItem` or `Item` (determined by the presence of the `id` field)
- **`UserVoiceStatus`** - Tuple `[User, string]` as a JSON array

## Project Structure

```
highrise-go/
├── highrise/                    # Core library
│   ├── client.go               # WebSocket client, event routing, reconnection
│   ├── actions.go              # Bot action methods (30+ methods)
│   ├── models.go               # Data types, JSON marshaling, request/response types
│   ├── bot.go                  # BotHandler interface, Bot base struct
│   ├── webapi.go               # REST client for Highrise WebAPI
│   ├── ratelimit.go            # Sliding window rate limiter
│   ├── errors.go               # Error types
│   └── *_test.go               # Tests & benchmarks
├── cmd/                         # Example bots
│   ├── highrise-bot/           # Full-featured demo bot
│   ├── economy-bot/            # Economy commands (!wallet, !tip, !buy)
│   ├── moderation-bot/         # Moderation commands (!kick, !ban, !mute)
│   └── dm-bot/                 # Direct message forwarder
├── go.mod
├── go.sum
└── .github/workflows/ci.yml   # CI: build, vet, test, benchmark
```

## Requirements

- Go 1.22+
- Dependency: [gorilla/websocket](https://github.com/gorilla/websocket) v1.5.3

## CI

Automated pipeline on GitHub Actions:

- **Build & Vet** - Compilation and static analysis
- **Test** - Unit tests (Linux with race detector, Windows without)
- **Benchmark** - Performance benchmarks (Linux only)

Matrix: `ubuntu-latest` + `windows-latest`

## License

MIT
