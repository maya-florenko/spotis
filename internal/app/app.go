package app

import (
	"context"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/maya-florenko/spotis/internal/app/handlers"
)

func Init(ctx context.Context) error {
	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(os.Getenv("TELEGRAM_TOKEN"), opts...)
	if err != nil {
		return err
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommand, handlers.StartHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "get", bot.MatchTypeCommand, handlers.GetHandler)
	b.Start(ctx)

	return nil
}

func handler(ctx context.Context, b *bot.Bot, u *models.Update) {
	if u.InlineQuery != nil {
		handlers.InlineHandler(ctx, b, u)
		return
	}

	handlers.MessageHandler(ctx, b, u)
}
