package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func StartHandler(ctx context.Context, b *bot.Bot, u *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: u.Message.Chat.ID,
		Text:   "hello everynyan~",
	})
}
