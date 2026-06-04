package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brianrstp/highrise-go/highrise"
)

type MyBot struct {
	highrise.Bot
}

func (b *MyBot) OnStart(ctx context.Context, session *highrise.SessionMetadata) {
	log.Printf("Bot started! UserID: %s, Room: %s", session.UserID, session.RoomInfo.RoomName)
	b.Highrise.Chat(ctx, "Hello everyone! I'm a Go bot!")
}

func (b *MyBot) OnChat(ctx context.Context, user highrise.User, message string) {
	log.Printf("[CHAT] %s: %s", user.Username, message)

	switch message {
	case "!ping":
		b.Highrise.Reply(ctx, user, "pong!")
	case "!users":
		users, err := b.Highrise.GetRoomUsers(ctx)
		if err != nil {
			log.Printf("Error getting users: %v", err)
			return
		}
		msg := fmt.Sprintf("Users in room (%d):", len(users))
		for _, entry := range users {
			msg += fmt.Sprintf(" @%s", entry.User.Username)
		}
		b.Highrise.Chat(ctx, msg)
	case "!wallet":
		wallet, err := b.Highrise.GetWallet(ctx)
		if err != nil {
			log.Printf("Error getting wallet: %v", err)
			return
		}
		for _, item := range wallet {
			b.Highrise.Chat(ctx, fmt.Sprintf("Wallet: %d %s", item.Amount, item.Type))
		}
	case "!myoutfit":
		outfit, err := b.Highrise.GetMyOutfit(ctx)
		if err != nil {
			log.Printf("Error getting outfit: %v", err)
			return
		}
		b.Highrise.Chat(ctx, fmt.Sprintf("Outfit items: %d", len(outfit)))
	}
}

func (b *MyBot) OnWhisper(ctx context.Context, user highrise.User, message string) {
	log.Printf("[WHISPER] %s: %s", user.Username, message)
	b.Highrise.WhisperReply(ctx, user, fmt.Sprintf("You said: %s", message))
}

func (b *MyBot) OnUserJoin(ctx context.Context, user highrise.User, position highrise.PositionOrAnchor) {
	log.Printf("[JOIN] %s joined the room", user.Username)
	if position.Position != nil {
		log.Printf("  Position: (%.1f, %.1f, %.1f)", position.Position.X, position.Position.Y, position.Position.Z)
	}
	b.Highrise.Reply(ctx, user, "welcome to the room!")
}

func (b *MyBot) OnUserLeave(ctx context.Context, user highrise.User) {
	log.Printf("[LEAVE] %s left the room", user.Username)
}

func (b *MyBot) OnTip(ctx context.Context, sender, receiver highrise.User, item *highrise.TipItem) {
	if item.CurrencyItem != nil {
		log.Printf("[TIP] %s tipped %s: %d %s", sender.Username, receiver.Username, item.CurrencyItem.Amount, item.CurrencyItem.Type)
	} else if item.Item != nil {
		log.Printf("[TIP] %s tipped %s: item %s", sender.Username, receiver.Username, item.Item.ID)
	}
}

func (b *MyBot) OnVoiceChange(ctx context.Context, users []highrise.UserVoiceStatus, secondsLeft int) {
	log.Printf("[VOICE] %d users, %ds left", len(users), secondsLeft)
	for _, u := range users {
		log.Printf("  %s: %s", u.User.Username, u.Status)
	}
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
