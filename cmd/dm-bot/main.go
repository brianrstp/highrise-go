package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/brianrstp/highrise-go/highrise"
)

type DMBot struct {
	highrise.Bot
}

func (b *DMBot) OnStart(ctx context.Context, session *highrise.SessionMetadata) {
	log.Printf("DM Bot started! UserID: %s", session.UserID)
}

func (b *DMBot) OnMessage(ctx context.Context, userID, conversationID string, isNew bool) {
	log.Printf("Message from %s (conv: %s, new: %v)", userID, conversationID, isNew)

	if isNew {
		b.Highrise.SendMessage(ctx, conversationID,
			"Hi! I'm a DM bot. I can forward your message to the room admin. Type your message!", "text", nil, nil, nil)
	}
}

func (b *DMBot) OnChat(ctx context.Context, user highrise.User, message string) {
	if strings.HasPrefix(message, "!dm ") {
		parts := strings.SplitN(message, " ", 3)
		if len(parts) < 3 {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Usage: !dm <user_id> <message>", user.Username))
			return
		}
		targetID := parts[1]
		dmMsg := parts[2]

		convs, _, err := b.Highrise.GetConversations(ctx, false, nil)
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Error fetching conversations", user.Username))
			return
		}

		for _, conv := range convs {
			for _, memberID := range conv.MemberIDs {
				if memberID == targetID {
					b.Highrise.SendMessage(ctx, conv.ID, dmMsg, "text", nil, nil, nil)
					b.Highrise.Chat(ctx, fmt.Sprintf("@%s Message sent!", user.Username))
					return
				}
			}
		}

		b.Highrise.Chat(ctx, fmt.Sprintf("@%s No shared conversation with that user", user.Username))
	}
}

func main() {
	bot := &DMBot{}
	client := highrise.NewClient(bot)
	bot.SetHighrise(client.Highrise())

	ctx := context.Background()
	if err := client.Run(ctx, os.Args[1], os.Args[2]); err != nil {
		log.Fatal(err)
	}
}
