package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/brianrstp/highrise-go/highrise"
)

type ModBot struct {
	highrise.Bot
}

func (b *ModBot) OnStart(ctx context.Context, session *highrise.SessionMetadata) {
	log.Printf("Mod Bot started! Room: %s", session.RoomInfo.RoomName)
}

func (b *ModBot) OnChat(ctx context.Context, user highrise.User, message string) {
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "!kick":
		if len(parts) < 2 {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Usage: !kick <user_id>", user.Username))
			return
		}
		err := b.Highrise.ModerateRoom(ctx, parts[1], "kick", nil)
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Kick failed: %v", user.Username, err))
		} else {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s User %s kicked", user.Username, parts[1]))
		}

	case "!ban":
		if len(parts) < 2 {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Usage: !ban <user_id> [hours]", user.Username))
			return
		}
		hours := 24
		if len(parts) >= 3 {
			fmt.Sscanf(parts[2], "%d", &hours)
		}
		duration := hours * 3600
		err := b.Highrise.ModerateRoom(ctx, parts[1], "ban", &duration)
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Ban failed: %v", user.Username, err))
		} else {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s User %s banned for %d hours", user.Username, parts[1], hours))
		}

	case "!mute":
		if len(parts) < 2 {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Usage: !mute <user_id> [minutes]", user.Username))
			return
		}
		mins := 5
		if len(parts) >= 3 {
			fmt.Sscanf(parts[2], "%d", &mins)
		}
		duration := mins * 60
		err := b.Highrise.ModerateRoom(ctx, parts[1], "mute", &duration)
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Mute failed: %v", user.Username, err))
		} else {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s User %s muted for %d minutes", user.Username, parts[1], mins))
		}

	case "!unban":
		if len(parts) < 2 {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Usage: !unban <user_id>", user.Username))
			return
		}
		err := b.Highrise.ModerateRoom(ctx, parts[1], "unban", nil)
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Unban failed: %v", user.Username, err))
		} else {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s User %s unbanned", user.Username, parts[1]))
		}

	case "!setmod":
		if len(parts) < 2 {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Usage: !setmod <user_id>", user.Username))
			return
		}
		trueVal := true
		err := b.Highrise.ChangeRoomPrivilege(ctx, parts[1], highrise.RoomPermissions{Moderator: &trueVal})
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Failed: %v", user.Username, err))
		} else {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s User %s is now moderator", user.Username, parts[1]))
		}

	case "!unmod":
		if len(parts) < 2 {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Usage: !unmod <user_id>", user.Username))
			return
		}
		falseVal := false
		err := b.Highrise.ChangeRoomPrivilege(ctx, parts[1], highrise.RoomPermissions{Moderator: &falseVal})
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Failed: %v", user.Username, err))
		} else {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s User %s is no longer moderator", user.Username, parts[1]))
		}

	case "!move":
		if len(parts) < 3 {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Usage: !move <user_id> <room_id>", user.Username))
			return
		}
		err := b.Highrise.MoveUserToRoom(ctx, parts[1], parts[2])
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Move failed: %v", user.Username, err))
		} else {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s User moved to room", user.Username))
		}
	}
}

func (b *ModBot) OnModerate(ctx context.Context, moderatorID, targetUserID, modType string, duration *int) {
	durStr := "permanent"
	if duration != nil {
		durStr = fmt.Sprintf("%d seconds", *duration)
	}
	log.Printf("[AUDIT] %s performed %s on %s for %s", moderatorID, modType, targetUserID, durStr)
}

func (b *ModBot) OnError(ctx context.Context, err highrise.Error) {
	log.Printf("[ERROR] %s (reconnect=%v)", err.Message, !err.DoNotReconnect)
}

func main() {
	bot := &ModBot{}
	client := highrise.NewClient(bot,
		highrise.WithSDKVersion("1.0.0"),
		highrise.WithActionTimeout(30*time.Second),
	)
	bot.SetHighrise(client.Highrise())

	ctx := context.Background()
	if err := client.Run(ctx, os.Args[1], os.Args[2]); err != nil {
		log.Fatal(err)
	}
}
