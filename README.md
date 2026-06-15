# highrise-go

Go SDK untuk [Highrise Bot API](https://highrise.game). Menyediakan WebSocket client untuk real-time bot actions (chat, walk, teleport, dll), event handler interfaces, rate limiting, middleware system, dan REST client untuk Highrise WebAPI.

## Fitur

- **WebSocket Client** - Koneksi real-time ke Highrise rooms dengan auto-reconnect (exponential backoff + jitter)
- **30+ Bot Actions** - Chat, whisper, walk, teleport, emote, moderate, tip, buy items, kelola outfit, dan banyak lagi
- **Event Handler** - 14 event handlers (chat, join, leave, move, emote, reaction, tip, whisper, voice, channel, DM, moderation)
- **Middleware System** - Wrap setiap event handler dengan custom logic (logging, metrics, auth)
- **Rate Limiting** - Sliding window rate limiter otomatis berdasarkan server-sent limits
- **REST API Client** - Akses Highrise WebAPI untuk users, rooms, items, grab bags, dan posts
- **Context-Aware** - Semua methods mendukung `context.Context` untuk cancellation dan timeout
- **Connection State** - Monitor status koneksi (Disconnected, Connecting, Connected, Reconnecting)
- **Metrics** - Counter events, actions, errors, dan reconnects
- **Worker Pool** - Bounded goroutine pool (max 64) untuk event handling
- **JSON Pool** - Buffer reuse via `sync.Pool` untuk mengurangi memory allocations

## Instalasi

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

Jalankan:

```bash
go run . <room_id> <api_token>
```

## Arsitektur

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

### Bot Struct (Embed untuk Default No-Op)

Embed `highrise.Bot` di struct bot kamu. Semua event handler akan memiliki default no-op, jadi kamu hanya perlu override method yang diinginkan.

```go
type MyBot struct {
    highrise.Bot
}

// Override hanya yang kamu butuhkan
func (b *MyBot) OnChat(ctx context.Context, user highrise.User, message string) {
    // handle chat
}
```

### Semua Event Handlers

| Event | Method | Deskripsi |
|-------|--------|-----------|
| Connect | `BeforeStart(ctx)` | Dipanggil sebelum koneksi (opsional, implement `HasBeforeStart`) |
| Session | `OnStart(ctx, *SessionMetadata)` | Koneksi established, session metadata diterima |
| Chat | `OnChat(ctx, User, message)` | User mengirim chat publik |
| Whisper | `OnWhisper(ctx, User, message)` | User mengirim whisper |
| Emote | `OnEmote(ctx, User, emoteID, *User)` | User melakukan emote |
| Reaction | `OnReaction(ctx, User, reaction, User)` | User memberikan reaction |
| User Join | `OnUserJoin(ctx, User, PositionOrAnchor)` | User masuk room |
| User Leave | `OnUserLeave(ctx, User)` | User keluar room |
| User Move | `OnUserMove(ctx, User, PositionOrAnchor)` | User berpindah posisi |
| Tip | `OnTip(ctx, sender, receiver, *TipItem)` | User memberikan tip |
| Voice | `OnVoiceChange(ctx, []UserVoiceStatus, secondsLeft)` | Status voice chat berubah |
| Channel | `OnChannel(ctx, senderID, message, tags)` | Pesan channel diterima |
| DM | `OnMessage(ctx, userID, convID, isNew)` | Pesan inbox diterima |
| Moderate | `OnModerate(ctx, modID, targetID, modType, *duration)` | Moderasi terjadi |
| Error | `OnError(ctx, Error)` | Server mengirim error |
| Any Event | `OnAnyEvent(ctx, eventType, data)` | Setiap event (raw JSON) |
| State | `OnConnectionChange(ctx, ConnectionState)` | Status koneksi berubah |

### Cara Implement

Gunakan interface `Has*` untuk mendaftarkan handler secara optional:

```go
// Implement langsung di struct (otomatis terdeteksi)
func (b *MyBot) OnChat(ctx context.Context, user highrise.User, message string) { ... }

// Atau via interface check
var _ highrise.HasOnChat = (*MyBot)(nil)
```

## Bot Actions

Semua actions diakses melalui `b.Highrise` (atau `client.Highrise()`).

### Chat & Messaging

```go
// Kirim chat publik
b.Highrise.Chat(ctx, "Hello!")

// Kirim whisper ke user tertentu
b.Highrise.SendWhisper(ctx, userID, "Secret message")

// Balas chat user dengan @username
b.Highrise.Reply(ctx, user, "pong!")

// Balas whisper
b.Highrise.WhisperReply(ctx, user, "Secret reply")

// Set indicator icon (nil untuk remove)
icon := "star"
b.Highrise.SetIndicator(ctx, &icon)
b.Highrise.SetIndicator(ctx, nil)

// Kirim ke channel
b.Highrise.SendChannel(ctx, "channel message", []string{"tag1", "tag2"})
```

### Movement

```go
// Jalan ke posisi
b.Highrise.WalkTo(ctx, highrise.Position{X: 5, Y: 0, Z: 3, Facing: "FrontRight"})

// Jalan ke anchor
b.Highrise.WalkToAnchor(ctx, highrise.AnchorPosition{EntityID: "ent_123", AnchorIx: 0})

// Teleport user
b.Highrise.Teleport(ctx, userID, highrise.Position{X: 0, Y: 0, Z: 0, Facing: "Front"})
```

### Emotes & Reactions

```go
// Kirim emote (targetUserID optional)
b.Highrise.SendEmote(ctx, "emoji_laugh", nil)
b.Highrise.SendEmote(ctx, "emoji_wave", &targetUserID)

// React ke user
b.Highrise.React(ctx, "heart", targetUserID)
```

### Moderation

```go
// Kick user
b.Highrise.ModerateRoom(ctx, userID, "kick", nil)

// Ban user (duration dalam detik, nil = permanent)
duration := 3600 // 1 jam
b.Highrise.ModerateRoom(ctx, userID, "ban", &duration)

// Mute user
duration = 300 // 5 menit
b.Highrise.ModerateRoom(ctx, userID, "mute", &duration)

// Unban user
b.Highrise.ModerateRoom(ctx, userID, "unban", nil)

// Set/remove moderator
modTrue := true
modFalse := false
b.Highrise.ChangeRoomPrivilege(ctx, userID, highrise.RoomPermissions{Moderator: &modTrue})
b.Highrise.ChangeRoomPrivilege(ctx, userID, highrise.RoomPermissions{Moderator: &modFalse})

// Pindahkan user ke room lain
b.Highrise.MoveUserToRoom(ctx, userID, roomID)
```

### Voice Chat

```go
// Cek status voice chat
status, err := b.Highrise.GetVoiceStatus(ctx)
// status.SecondsLeft, status.Users, status.AutoSpeakers

// Invite/remove speaker
b.Highrise.InviteSpeaker(ctx, userID)
b.Highrise.RemoveSpeaker(ctx, userID)

// Beli waktu voice
result, err := b.Highrise.BuyVoiceTime(ctx)
```

### Economy

```go
// Cek wallet
wallet, err := b.Highrise.GetWallet(ctx)
for _, item := range wallet {
    fmt.Printf("%s: %d\n", item.Type, item.Amount)
}

// Tip user
result, err := b.Highrise.TipUser(ctx, userID, "gold_bar_10")

// Beli item
result, err := b.Highrise.BuyItem(ctx, itemID)

// Beli room boost
result, err := b.Highrise.BuyRoomBoost(ctx, 1)
```

### Items & Outfit

```go
// Get inventory bot
items, err := b.Highrise.GetInventory(ctx)

// Set outfit bot
b.Highrise.SetOutfit(ctx, []highrise.Item{
    {Type: "shirt", ID: "item_123", Amount: 1, AccountBound: false},
})

// Get outfit user
outfit, err := b.Highrise.GetUserOutfit(ctx, userID)

// Get outfit bot sendiri
outfit, err := b.Highrise.GetMyOutfit(ctx)

// Get backpack user
bp, err := b.Highrise.GetBackpack(ctx, userID)
// bp = map[string]int{"gold_bar_10": 5}

// Change backpack
b.Highrise.ChangeBackpack(ctx, userID, map[string]int{
    "gold_bar_10": -1, // kurangi 1
})
```

### Room Info

```go
// Get semua user di room
users, err := b.Highrise.GetRoomUsers(ctx)
for _, entry := range users {
    fmt.Printf("%s at (%.1f, %.1f, %.1f)\n",
        entry.User.Username,
        entry.Position.Position.X,
        entry.Position.Position.Y,
        entry.Position.Position.Z,
    )
}

// Get privilege user
perms, err := b.Highrise.GetRoomPrivilege(ctx, userID)
// perms.Moderator, perms.Designer
```

### Direct Messages (DM)

```go
// Get conversations
convs, notJoined, err := b.Highrise.GetConversations(ctx, false, nil)

// Kirim pesan
b.Highrise.SendMessage(ctx, convID, "Hello!", "text", nil, nil, nil)

// Kirim pesan dengan media
b.Highrise.SendMessage(ctx, convID, "Check this out!", "text", nil, nil, &mediaID)

// Kirim bulk message
b.Highrise.SendBulkMessage(ctx, []string{"uid1", "uid2"}, "Broadcast!", "text", nil, nil)

// Get messages
msgs, err := b.Highrise.GetMessages(ctx, convID, lastMessageID)

// Leave conversation
b.Highrise.LeaveConversation(ctx, convID)

// Upload media
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

// Custom SDK version (dikirim ke server)
client := highrise.NewClient(bot, highrise.WithSDKVersion("1.0.0"))

// Custom event subscription
client := highrise.NewClient(bot, highrise.WithEvents("chat,user_joined,user_left"))

// Custom action timeout (default: 10s, 0 = disable)
client := highrise.NewClient(bot, highrise.WithActionTimeout(30*time.Second))
```

## Middleware

Middleware membungkus setiap event handler call. Berguna untuk logging, metrics, atau auth checks.

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

// Multiple middleware (dijalankan berurutan)
client.Use(middleware1)
client.Use(middleware2)
// Urutan: middleware1 → middleware2 → actual handler
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

// Via client
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

// Handle signal
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

Atau gunakan `client.Stop()`:

```go
go client.Run(ctx, roomID, apiToken)
// ... later
client.Stop() // graceful shutdown
```

## Error Handling

```go
// ResponseError - error dari server
var respErr *highrise.ResponseError
if errors.As(err, &respErr) {
    log.Printf("Server error: %s", respErr.Message)
}

// ConnectionError - error dari WebSocket layer
var connErr *highrise.ConnectionError
if errors.As(err, &connErr) {
    log.Printf("Connection error [%s]: %v", connErr.ReqType, connErr.Err)
}
```

### Server Error dengan DoNotReconnect

Jika server mengirim error dengan `DoNotReconnect: true`, client akan otomatis stop. Implement `HasOnError` untuk menanganinya:

```go
func (b *MyBot) OnError(ctx context.Context, err highrise.Error) {
    log.Printf("Error: %s (reconnect: %v)", err.Message, !err.DoNotReconnect)
}
```

## WebAPI (REST Client)

Untuk data read-only tanpa koneksi WebSocket.

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

Semua list endpoints mendukung pagination:

```go
// Halaman pertama
resp1, _ := api.GetRooms(ctx, highrise.RoomsListParams{Limit: 20})

// Halaman berikutnya
resp2, _ := api.GetRooms(ctx, highrise.RoomsListParams{
    Limit:       20,
    StartsAfter: resp1.LastID,
})

// Halaman sebelumnya
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
    Type         string `json:"type"`
    Amount       int    `json:"amount"`
    ID           string `json:"id"`
    AccountBound bool   `json:"account_bound"`
    ActivePalette *int  `json:"active_palette,omitempty"`
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

Beberapa type memiliki custom JSON marshaling:

- **`PositionOrAnchor`** - Bisa berupa `Position` atau `AnchorPosition` (ditentukan oleh keberadaan field `entity_id`)
- **`TipItem`** - Bisa berupa `CurrencyItem` atau `Item` (ditentukan oleh keberadaan field `id`)
- **`UserVoiceStatus`** - Tuple `[User, string]` dalam JSON array

## Struktur Project

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

Pipeline otomatis di GitHub Actions:

- **Build & Vet** - Compile dan static analysis
- **Test** - Unit tests (Linux dengan race detector, Windows tanpa)
- **Benchmark** - Performance benchmarks (Linux only)

Matrix: `ubuntu-latest` + `windows-latest`

## License

MIT
