# highrise-go

Highrise Bot SDK for Go. Port of the official [python-bot-sdk](https://github.com/pocketzworld/python-bot-sdk).

## Installation

```bash
go get github.com/gorilla/websocket
```

## Quick Start

Create `main.go`:

```go
package main

import (
    "context"
    "log"
    "os"

    "highrise-go/highrise"
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
├── highrise/
│   ├── bot.go        # Bot struct + handler interfaces
│   ├── client.go     # WebSocket client + event routing
│   ├── models.go     # All data types, requests, responses
│   ├── actions.go    # Action methods (chat, walk, teleport, etc.)
│   ├── webapi.go     # REST API client
│   └── errors.go     # Error types
├── cmd/highrise-bot/
│   └── main.go       # Example bot
└── README.md
```

## Event Handlers

All handlers are **optional** — override only the methods you need:

| Handler | Event |
|---|---|
| `OnStart(session)` | Connection established, session metadata received |
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

### Moderation
```go
b.Highrise.ModerateRoom(ctx, userID, "kick", nil)
b.Highrise.ModerateRoom(ctx, userID, "ban", intPtr(3600))
b.Highrise.ChangeRoomPrivilege(ctx, userID, highrise.RoomPermissions{Moderator: boolPtr(true)})
b.Highrise.MoveUserToRoom(ctx, userID, roomID)
```

### Voice Chat
```go
b.Highrise.InviteSpeaker(ctx, userID)
b.Highrise.RemoveSpeaker(ctx, userID)
status, _ := b.Highrise.GetVoiceStatus(ctx)
b.Highrise.BuyVoiceTime(ctx)
```

### Economy
```go
wallet, _ := b.Highrise.GetWallet(ctx)
result, _ := b.Highrise.TipUser(ctx, userID, "gold_bar_10")
result, _ := b.Highrise.BuyItem(ctx, itemID)
b.Highrise.BuyRoomBoost(ctx, 1)
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
msgs, _ := b.Highrise.GetMessages(ctx, convID, "")
b.Highrise.LeaveConversation(ctx, convID)
```

### REST API (WebAPI)
```go
webapi := highrise.NewWebAPI()
user, _ := webapi.GetUser(ctx, userID)
room, _ := webapi.GetRoom(ctx, roomID)
item, _ := webapi.GetItem(ctx, itemID)
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

## Testing

```bash
go test -v -count=1 ./highrise/
```

## Notes

- WebSocket URL: `wss://highrise.game/web/botapi`
- Keepalive every 15 seconds
- Reconnection with exponential backoff (1s → 30s max)
- Module name: `highrise-go` (local module, no external domain dependency)
