package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/brianrstp/highrise-go/highrise"
)

type EconomyBot struct {
	highrise.Bot
}

func (b *EconomyBot) OnStart(ctx context.Context, session *highrise.SessionMetadata) {
	log.Printf("Economy Bot started! UserID: %s", session.UserID)
}

func (b *EconomyBot) OnChat(ctx context.Context, user highrise.User, message string) {
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "!wallet":
		wallet, err := b.Highrise.GetWallet(ctx)
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Error: %v", user.Username, err))
			return
		}
		msg := fmt.Sprintf("@%s Wallet:", user.Username)
		for _, c := range wallet {
			msg += fmt.Sprintf(" %d %s,", c.Amount, c.Type)
		}
		msg = strings.TrimRight(msg, ",")
		b.Highrise.Chat(ctx, msg)

	case "!tip":
		if len(parts) < 3 {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Usage: !tip <user_id> <gold_bar_type>", user.Username))
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Example: !tip abc123 gold_bar_10", user.Username))
			return
		}
		result, err := b.Highrise.TipUser(ctx, parts[1], parts[2])
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Tip failed: %v", user.Username, err))
		} else {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Tip result: %s", user.Username, result))
		}

	case "!backpack":
		if len(parts) < 2 {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Usage: !backpack <user_id>", user.Username))
			return
		}
		bp, err := b.Highrise.GetBackpack(ctx, parts[1])
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Error: %v", user.Username, err))
			return
		}
		msg := fmt.Sprintf("@%s Backpack:", user.Username)
		for item, qty := range bp {
			msg += fmt.Sprintf(" %s x%d,", item, qty)
		}
		msg = strings.TrimRight(msg, ",")
		b.Highrise.Chat(ctx, msg)

	case "!inventory":
		items, err := b.Highrise.GetInventory(ctx)
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Error: %v", user.Username, err))
			return
		}
		msg := fmt.Sprintf("@%s Inventory (%d items):", user.Username, len(items))
		for _, item := range items {
			msg += fmt.Sprintf(" %s x%d,", item.ID, item.Amount)
		}
		msg = strings.TrimRight(msg, ",")
		b.Highrise.Chat(ctx, msg)

	case "!buy":
		if len(parts) < 2 {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Usage: !buy <item_id>", user.Username))
			return
		}
		result, err := b.Highrise.BuyItem(ctx, parts[1])
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Buy failed: %v", user.Username, err))
		} else {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Purchase result: %s", user.Username, result))
		}

	case "!outfit":
		if len(parts) < 2 {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Usage: !outfit <user_id>", user.Username))
			return
		}
		outfit, err := b.Highrise.GetUserOutfit(ctx, parts[1])
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Error: %v", user.Username, err))
			return
		}
		msg := fmt.Sprintf("@%s Outfit:", user.Username)
		for _, item := range outfit {
			msg += fmt.Sprintf(" %s,", item.ID)
		}
		msg = strings.TrimRight(msg, ",")
		b.Highrise.Chat(ctx, msg)

	case "!myoutfit":
		outfit, err := b.Highrise.GetMyOutfit(ctx)
		if err != nil {
			b.Highrise.Chat(ctx, fmt.Sprintf("@%s Error: %v", user.Username, err))
			return
		}
		msg := fmt.Sprintf("@%s My outfit:", user.Username)
		for _, item := range outfit {
			msg += fmt.Sprintf(" %s,", item.ID)
		}
		msg = strings.TrimRight(msg, ",")
		b.Highrise.Chat(ctx, msg)
	}
}

func (b *EconomyBot) OnTip(ctx context.Context, sender, receiver highrise.User, item *highrise.TipItem) {
	if item.CurrencyItem != nil {
		log.Printf("TIP: %s sent %d %s to %s",
			sender.Username, item.CurrencyItem.Amount, item.CurrencyItem.Type, receiver.Username)
	} else if item.Item != nil {
		log.Printf("TIP: %s sent item %s to %s",
			sender.Username, item.Item.ID, receiver.Username)
	}
}

func main() {
	bot := &EconomyBot{}
	client := highrise.NewClient(bot,
		highrise.WithSDKVersion("1.0.0"),
	)
	bot.SetHighrise(client.Highrise())

	ctx := context.Background()
	if err := client.Run(ctx, os.Args[1], os.Args[2]); err != nil {
		log.Fatal(err)
	}
}
