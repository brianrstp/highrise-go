# highrise-go

Highrise Bot SDK for Go. Port of the official [python-bot-sdk](https://github.com/pocketzworld/python-bot-sdk).

## Installation

```bash
go get github.com/brianrstp/highrise-go@latest
```

## Quick Start

Create `main.go`:

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
    log.Printf("Bot started! UserID: %s, Room: %s", session.UserID, session.RoomInfo.RoomName)
    b.Highrise.Chat(ctx, "Hello everyone!")
}

func (b *MyBot) OnChat(ctx context.Context, user highrise.User, message string) {
    log.Printf("%s: %s", user.Username, message)

    switch message {
    case "!ping":
        b.Highrise.Chat(ctx, "pong!")
    case "!users":
        users, _ := b.Highrise.GetRoomUsers(ctx)
        msg := "Users:"
        for _, u := range users {
            msg += " @" + u.User.Username
        }
        b.Highrise.Chat(ctx, msg)
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

Run:

```bash
go run main.go <room_id> <api_token>
```

## Project Structure

```
highrise-go/
├── .github/workflows/
│   └── ci.yml            # GitHub Actions CI (build, vet, test, benchmark)
├── highrise/
│   ├── bot.go            # Bot struct + handler interfaces
│   ├── client.go         # WebSocket client + event routing + middleware + metrics
│   ├── models.go         # All data types, requests, responses
│   ├── actions.go        # Action methods (chat, walk, teleport, etc.)
│   ├── webapi.go         # REST API client
│   ├── errors.go         # Error types (ResponseError, ConnectionError)
│   ├── ratelimit.go      # Sliding window rate limiter
│   ├── actions_test.go   # Integration tests for actions
│   ├── client_test.go    # WebSocket client tests (OnAnyEvent, routing, etc.)
│   ├── models_test.go    # JSON serialization tests
│   ├── ratelimit_test.go # Rate limiter unit tests
│   ├── webapi_test.go    # REST API client tests
│   └── benchmark_test.go # Benchmarks (marshal, unmarshal, rate limiter, etc.)
├── cmd/
│   ├── dm-bot/           # DM bot example
│   ├── moderation-bot/   # Moderation bot example
│   └── economy-bot/      # Economy bot example
├── DOCUMENTS.md           # Blog-format documentation
└── README.md
```

## Event Handlers

All handlers are **optional** — override only the methods you need:

| Handler | Event |
|---|---|---|
| `BeforeStart()` | Called before connection, useful for setup |
| `OnStart(session)` | Connection established, session metadata received |
| `OnAnyEvent(eventType, data)` | **Every** event (raw type + JSON bytes), fires before typed handler |
| `OnChat(user, message)` | Public room chat |
| `OnWhisper(user, message)` | Private whisper |
| `OnUserJoin(user, position)` | User entered the room |
| `OnUserLeave(user)` | User left the room |
| `OnUserMove(user, position)` | User moved |
| `OnEmote(user, emoteID, receiver)` | Emote received |
| `OnReaction(user, reaction, receiver)` | Reaction received |
| `OnTip(sender, receiver, item)` | Tip received |
| `OnVoiceChange(users, secondsLeft)` | Voice chat status changed |
| `OnChannel(senderID, message, tags)` | Hidden channel message |
| `OnMessage(userID, conversationID, isNew)` | Inbox message |
| `OnModerate(moderatorID, targetUserID, type, duration)` | Room moderation |
| `OnError(err)` | Server error received |
| `OnConnectionChange(state)` | Connection state changed |

Example:

```go
func (b *MyBot) OnUserJoin(ctx context.Context, user highrise.User, pos highrise.PositionOrAnchor) {
    b.Highrise.Chat(ctx, fmt.Sprintf("Welcome @%s!", user.Username))
}

func (b *MyBot) OnTip(ctx context.Context, sender, receiver highrise.User, item *highrise.TipItem) {
    if item.CurrencyItem != nil {
        log.Printf("Tipped %d %s", item.CurrencyItem.Amount, item.CurrencyItem.Type)
    }
}
```

## Actions

All actions are available via `b.Highrise.*`:

### Chat & Interaction
```go
b.Highrise.Chat(ctx, "Hello!")                          // Broadcast chat
b.Highrise.SendWhisper(ctx, userID, "secret")            // Whisper a user
b.Highrise.Reply(ctx, user, "hello!")                    // Chat with @username prefix
b.Highrise.WhisperReply(ctx, user, "secret")             // Whisper reply to user
b.Highrise.SendEmote(ctx, "emoji_laugh", nil)            // Send emote
b.Highrise.React(ctx, "heart", userID)                   // Send reaction
b.Highrise.SetIndicator(ctx, strPtr("mood_happy"))       // Set indicator icon
b.Highrise.SendChannel(ctx, "hidden msg", []string{"tag1"}) // Hidden channel message
```

### Movement
```go
b.Highrise.WalkTo(ctx, highrise.Position{X: 10, Y: 0, Z: 5, Facing: "FrontRight"})
b.Highrise.WalkToAnchor(ctx, highrise.AnchorPosition{EntityID: "ent_1", AnchorIx: 0})
b.Highrise.Teleport(ctx, userID, highrise.Position{X: 0, Y: 0, Z: 0, Facing: "FrontRight"})
```

### Moderation & Room Privileges
```go
b.Highrise.ModerateRoom(ctx, userID, "kick", nil)
b.Highrise.ModerateRoom(ctx, userID, "ban", intPtr(3600))
b.Highrise.ChangeRoomPrivilege(ctx, userID, highrise.RoomPermissions{Moderator: boolPtr(true)})
b.Highrise.MoveUserToRoom(ctx, userID, roomID)
b.Highrise.GetRoomPrivilege(ctx, userID)                  // Get user's room role
```

### Voice Chat
```go
b.Highrise.InviteSpeaker(ctx, userID)
b.Highrise.RemoveSpeaker(ctx, userID)
status, _ := b.Highrise.GetVoiceStatus(ctx)
b.Highrise.BuyVoiceTime(ctx)
```

### Economy & Backpack
```go
wallet, _ := b.Highrise.GetWallet(ctx)
result, _ := b.Highrise.TipUser(ctx, userID, "gold_bar_10")
result, _ := b.Highrise.BuyItem(ctx, itemID)
b.Highrise.BuyRoomBoost(ctx, 1)
bp, _ := b.Highrise.GetBackpack(ctx, userID)              // Get user's backpack
b.Highrise.ChangeBackpack(ctx, userID, map[string]int{"gold_bar_10": 5})
```

### Inventory & Outfit
```go
items, _ := b.Highrise.GetInventory(ctx)
outfit, _ := b.Highrise.GetUserOutfit(ctx, userID)
outfit, _ := b.Highrise.GetMyOutfit(ctx)
b.Highrise.SetOutfit(ctx, []highrise.Item{...})
```

### Room Users
```go
users, _ := b.Highrise.GetRoomUsers(ctx)
for _, entry := range users {
    fmt.Printf("%s at (%.1f, %.1f, %.1f)\n",
        entry.User.Username,
        entry.Position.Position.X,
        entry.Position.Position.Y,
        entry.Position.Position.Z,
    )
}
```

### Inbox Messages
```go
convs, _, _ := b.Highrise.GetConversations(ctx, false, nil)
b.Highrise.SendMessage(ctx, convID, "Hello!", "text", nil, nil, nil)
b.Highrise.SendBulkMessage(ctx, []string{userID}, "Hello!", "text", nil, nil)
msgs, _ := b.Highrise.GetMessages(ctx, convID, "")
b.Highrise.LeaveConversation(ctx, convID)
```

### Media Upload
```go
resp, _ := b.Highrise.MessageMediaUpload(ctx, highrise.MessageMedia{
    Type: "image", Width: 800, Height: 600,
})
// Upload to resp.UploadURL, thumbnail to resp.ThumbnailUploadURL
```

### REST API (WebAPI)
```go
webapi := highrise.NewWebAPI()
user, _ := webapi.GetUser(ctx, userID)

// List endpoints return paginated responses
usersResp, _ := webapi.GetUsers(ctx, highrise.UsersListParams{Limit: 50})
fmt.Printf("Total: %d, First: %s, Last: %s", usersResp.Total, usersResp.FirstID, usersResp.LastID)

room, _ := webapi.GetRoom(ctx, roomID)
roomsResp, _ := webapi.GetRooms(ctx, highrise.RoomsListParams{RoomName: "cozy", Limit: 10})

item, _ := webapi.GetItem(ctx, itemID)
itemsResp, _ := webapi.GetItems(ctx, highrise.ItemsListParams{Category: "watch", Limit: 20})

grab, _ := webapi.GetGrab(ctx, grabID)
grabsResp, _ := webapi.GetGrabs(ctx, highrise.GrabsListParams{Limit: 10})

post, _ := webapi.GetPost(ctx, postID)
postsResp, _ := webapi.GetPosts(ctx, highrise.PostsListParams{AuthorID: userID, Limit: 10})
```

## Data Types

### Position & AnchorPosition

```go
// Regular floor position
pos := highrise.Position{X: 10.5, Y: 0, Z: 3.2, Facing: "FrontRight"}

// Anchor position (entity-based, e.g. sitting on a chair)
anchor := highrise.AnchorPosition{EntityID: "ent_123", AnchorIx: 2}

// Event position (can be either)
var pos highrise.PositionOrAnchor
if pos.Position != nil {
    // Regular position
}
if pos.AnchorPosition != nil {
    // Anchor position
}
```

### TipItem

```go
var item *highrise.TipItem
if item.CurrencyItem != nil {
    fmt.Printf("%d %s\n", item.CurrencyItem.Amount, item.CurrencyItem.Type)
}
if item.Item != nil {
    fmt.Printf("Item: %s\n", item.Item.ID)
}
```

### RoomPermissions

```go
perms := highrise.RoomPermissions{
    Moderator: boolPtr(true),
    Designer:  boolPtr(false),
}
```

## Client Options

Customize the client with functional options:

```go
client := highrise.NewClient(bot,
    highrise.WithURL("wss://custom.highrise.game/web/botapi"),
    highrise.WithLogger(myLogger),
    highrise.WithSDKVersion("1.0.0"),
)
```

## Custom Logger

Implement the `Logger` interface to use your own logger:

```go
type Logger interface {
    Printf(format string, v ...any)
}

// Example with log/slog:
client.SetLogger(slog.Default())

// Example with zerolog:
client.SetLogger(zerologLogger)
```

## Middleware

Middleware wraps every event handler call. Useful for logging, metrics, or filtering:

```go
client.Use(func(next func()) {
    log.Println("before event")
    next()
    log.Println("after event")
})
```

Multiple middleware run in FIFO order.

## Connection State Tracking

Implement `HasOnConnectionChange` to track connection state:

```go
type MyBot struct {
    highrise.Bot
}

func (b *MyBot) OnConnectionChange(ctx context.Context, state highrise.ConnectionState) {
    switch state {
    case highrise.StateConnected:
        log.Println("Bot connected!")
    case highrise.StateDisconnected:
        log.Println("Bot disconnected")
    case highrise.StateConnecting:
        log.Println("Connecting...")
    case highrise.StateReconnecting:
        log.Println("Reconnecting...")
    }
}
```

States: `StateDisconnected`, `StateConnecting`, `StateConnected`, `StateReconnecting`.

## Event Handlers

For selective event handling without implementing all `BotHandler` methods,
implement one of the optional interfaces (`HasOnChat`, `HasOnStart`, etc.)
directly instead of embedding `Bot`.

## Testing

```bash
# Run all tests
go test -v -count=1 ./highrise/

# Run benchmarks
go test -bench=. -benchmem -count=1 ./highrise/

# Run with race detector (Linux only — requires gcc)
go test -race -count=1 ./highrise/
```

## Metrics

Built-in counters via `client.Metrics()`:

```go
m := client.Metrics()
fmt.Printf("Events: %d, Actions: %d, Errors: %d, Reconnects: %d\n",
    m["events"], m["actions"], m["errors"], m["reconnects"])
```

## OnAnyEvent

Log every event without implementing specific handlers:

```go
func (b *MyBot) OnAnyEvent(ctx context.Context, eventType string, data []byte) {
    log.Printf("Event: %s — %s", eventType, string(data))
}
```

Fires for all events (including unknown types) before the typed handler.

## Notes

- WebSocket URL: `wss://highrise.game/web/botapi`
- Keepalive every 15 seconds
- Reconnection with exponential backoff (1s → 30s max) + jitter
- TCP write deadline (10s), read deadline (20s), pong handler extend
- Rate limiter: sliding window per action + global limit
- Panic recovery: all handler panics are caught and logged
- Event semaphore: non-blocking dispatch, drops when pool is full
- `ConnectionError` includes `ReqType` and `RID` for debugging
- Module name: `github.com/brianrstp/highrise-go`
