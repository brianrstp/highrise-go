package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/brianrstp/highrise-go/highrise"
)

type MyBot struct {
	highrise.Bot
}

func (b *MyBot) OnStart(ctx context.Context, session *highrise.SessionMetadata) {
	log.Printf("Bot started!")
	log.Printf("  UserID: %s", session.UserID)
	log.Printf("  Room: %s (%s)", session.RoomInfo.RoomName, session.RoomInfo.OwnerID)
	log.Printf("  ConnectionID: %s", session.ConnectionID)
	b.Highrise.Chat(ctx, "Hello everyone! I'm a Go bot!")
}

func (b *MyBot) OnChat(ctx context.Context, user highrise.User, message string) {
	log.Printf("[CHAT] %s: %s", user.Username, message)
	if user.Username == "" {
		return
	}

	args := strings.Fields(message)
	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "!ping":
		b.Highrise.Reply(ctx, user, "pong!")

	case "!users":
		users, err := b.Highrise.GetRoomUsers(ctx)
		if err != nil {
			b.Highrise.Reply(ctx, user, "Error getting users")
			return
		}
		names := make([]string, 0, len(users))
		for _, entry := range users {
			names = append(names, "@"+entry.User.Username)
		}
		b.Highrise.Reply(ctx, user, fmt.Sprintf("Users (%d): %s", len(users), strings.Join(names, " ")))

	case "!wallet":
		wallet, err := b.Highrise.GetWallet(ctx)
		if err != nil {
			b.Highrise.Reply(ctx, user, "Error getting wallet")
			return
		}
		if len(wallet) == 0 {
			b.Highrise.Reply(ctx, user, "Wallet is empty")
			return
		}
		for _, item := range wallet {
			b.Highrise.Reply(ctx, user, fmt.Sprintf("Wallet: %d %s", item.Amount, item.Type))
		}

	case "!myoutfit":
		outfit, err := b.Highrise.GetMyOutfit(ctx)
		if err != nil {
			b.Highrise.Reply(ctx, user, "Error getting outfit")
			return
		}
		b.Highrise.Reply(ctx, user, fmt.Sprintf("Outfit items: %d", len(outfit)))

	case "!backpack":
		if len(args) < 2 {
			b.Highrise.Reply(ctx, user, "Usage: !backpack <user_id>")
			return
		}
		items, err := b.Highrise.GetBackpack(ctx, args[1])
		if err != nil {
			b.Highrise.Reply(ctx, user, "Error getting backpack")
			return
		}
		b.Highrise.Reply(ctx, user, fmt.Sprintf("Backpack: %d items", len(items)))

	case "!inventory":
		inv, err := b.Highrise.GetInventory(ctx)
		if err != nil {
			b.Highrise.Reply(ctx, user, "Error getting inventory")
			return
		}
		b.Highrise.Reply(ctx, user, fmt.Sprintf("Inventory: %d items", len(inv)))

	case "!outfit":
		if len(args) < 2 {
			b.Highrise.Reply(ctx, user, "Usage: !outfit <user_id>")
			return
		}
		outfit, err := b.Highrise.GetUserOutfit(ctx, args[1])
		if err != nil {
			b.Highrise.Reply(ctx, user, "Error getting outfit")
			return
		}
		b.Highrise.Reply(ctx, user, fmt.Sprintf("User outfit: %d items", len(outfit)))

	case "!voice":
		status, err := b.Highrise.GetVoiceStatus(ctx)
		if err != nil {
			b.Highrise.Reply(ctx, user, "Error getting voice status")
			return
		}
		if status == nil {
			b.Highrise.Reply(ctx, user, "Voice chat is not active")
			return
		}
		b.Highrise.Reply(ctx, user, fmt.Sprintf("Voice: %d users, %ds left", len(status.Users), status.SecondsLeft))

	case "!walk":
		if len(args) < 4 {
			b.Highrise.Reply(ctx, user, "Usage: !walk <x> <y> <z>")
			return
		}
		var x, y, z float64
		fmt.Sscanf(args[1], "%f", &x)
		fmt.Sscanf(args[2], "%f", &y)
		fmt.Sscanf(args[3], "%f", &z)
		err := b.Highrise.WalkTo(ctx, highrise.Position{X: x, Y: y, Z: z, Facing: "front"})
		if err != nil {
			b.Highrise.Reply(ctx, user, "Error walking")
			return
		}
		b.Highrise.Reply(ctx, user, fmt.Sprintf("Walking to (%.1f, %.1f, %.1f)", x, y, z))

	case "!emote":
		if len(args) < 2 {
			b.Highrise.Reply(ctx, user, "Usage: !emote <emote_id>")
			return
		}
		err := b.Highrise.SendEmote(ctx, args[1], nil)
		if err != nil {
			b.Highrise.Reply(ctx, user, "Error sending emote")
			return
		}
		b.Highrise.Reply(ctx, user, "Emote sent!")

	case "!help":
		help := strings.Join([]string{
			"Commands: !ping !users !wallet !myoutfit",
			"!backpack <id> !inventory !outfit <id>",
			"!voice !walk <x> <y> <z> !emote <id>",
		}, " | ")
		b.Highrise.Reply(ctx, user, help)
	}
}

func (b *MyBot) OnWhisper(ctx context.Context, user highrise.User, message string) {
	log.Printf("[WHISPER] %s: %s", user.Username, message)
	b.Highrise.WhisperReply(ctx, user, fmt.Sprintf("You whispered: %s", message))
}

func (b *MyBot) OnEmote(ctx context.Context, user highrise.User, emoteID string, receiver *highrise.User) {
	if receiver != nil {
		log.Printf("[EMOTE] %s -> %s: %s", user.Username, receiver.Username, emoteID)
	} else {
		log.Printf("[EMOTE] %s: %s", user.Username, emoteID)
	}
}

func (b *MyBot) OnReaction(ctx context.Context, user highrise.User, reaction string, receiver highrise.User) {
	log.Printf("[REACTION] %s -> %s: %s", user.Username, receiver.Username, reaction)
}

func (b *MyBot) OnUserJoin(ctx context.Context, user highrise.User, position highrise.PositionOrAnchor) {
	log.Printf("[JOIN] %s entered the room", user.Username)
	if position.Position != nil {
		log.Printf("  at (%.1f, %.1f, %.1f)", position.Position.X, position.Position.Y, position.Position.Z)
	}
	b.Highrise.Reply(ctx, user, "welcome to the room!")
}

func (b *MyBot) OnUserLeave(ctx context.Context, user highrise.User) {
	log.Printf("[LEAVE] %s left the room", user.Username)
}

func (b *MyBot) OnUserMove(ctx context.Context, user highrise.User, position highrise.PositionOrAnchor) {
	if position.Position != nil {
		log.Printf("[MOVE] %s -> (%.1f, %.1f, %.1f)", user.Username, position.Position.X, position.Position.Y, position.Position.Z)
	}
}

func (b *MyBot) OnTip(ctx context.Context, sender, receiver highrise.User, item *highrise.TipItem) {
	if item.CurrencyItem != nil {
		log.Printf("[TIP] %s tipped %s: %d %s", sender.Username, receiver.Username, item.CurrencyItem.Amount, item.CurrencyItem.Type)
	} else if item.Item != nil {
		log.Printf("[TIP] %s tipped %s: item %s", sender.Username, receiver.Username, item.Item.ID)
	}
}

func (b *MyBot) OnVoiceChange(ctx context.Context, users []highrise.UserVoiceStatus, secondsLeft int) {
	log.Printf("[VOICE] %d users active, %ds remaining", len(users), secondsLeft)
	for _, u := range users {
		log.Printf("  %s: %s", u.User.Username, u.Status)
	}
}

func (b *MyBot) OnChannel(ctx context.Context, senderID, message string, tags []string) {
	log.Printf("[CHANNEL] from %s: %s (tags: %v)", senderID, message, tags)
}

func (b *MyBot) OnMessage(ctx context.Context, userID, conversationID string, isNew bool) {
	log.Printf("[MESSAGE] from %s in %s (new: %v)", userID, conversationID, isNew)
}

func (b *MyBot) OnModerate(ctx context.Context, moderatorID, targetUserID, modType string, duration *int) {
	if duration != nil {
		log.Printf("[MODERATE] %s moderated %s: %s (%ds)", moderatorID, targetUserID, modType, *duration)
	} else {
		log.Printf("[MODERATE] %s moderated %s: %s", moderatorID, targetUserID, modType)
	}
}

func (b *MyBot) OnError(ctx context.Context, err highrise.Error) {
	log.Printf("[ERROR] %s (reconnect: %v)", err.Message, !err.DoNotReconnect)
}

func (b *MyBot) OnConnectionChange(ctx context.Context, state highrise.ConnectionState) {
	log.Printf("[STATE] %s", state)
}

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("Usage: highrise-bot <room_id> <api_token>")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  highrise-bot 641b78a7ad0857b5099f55b6 your-api-token-here")
		os.Exit(1)
	}

	roomID := args[0]
	apiToken := args[1]

	bot := &MyBot{}
	client := highrise.NewClient(bot)
	bot.SetHighrise(client.Highrise())

	client.Use(func(next func()) {
		start := time.Now()
		next()
		log.Printf("[EVENT] handled in %v", time.Since(start))
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	log.Printf("Starting bot...")
	log.Printf("Room ID: %s", roomID)

	if err := client.Run(ctx, roomID, apiToken); err != nil {
		log.Fatalf("Bot error: %v", err)
	}

	log.Printf("Bot stopped gracefully")
}
