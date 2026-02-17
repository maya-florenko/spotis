package app

import (
	"context"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/maya-florenko/spotis/internal/downloader"
)

var downloadManager *downloader.Manager

func Init(ctx context.Context) error {
	// Initialize download manager with available downloaders
	downloadManager = downloader.NewManager()
	downloadManager.Register(downloader.NewDeezerDownloaderFromEnv())

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, CommandStart),
	}

	b, err := bot.New(os.Getenv("TELEGRAM_TOKEN"), opts...)
	if err != nil {
		return err
	}

	b.Start(ctx)

	return nil
}

func handler(ctx context.Context, b *bot.Bot, u *models.Update) {
	MessageHandler(ctx, b, u)
	InlineHandler(ctx, b, u)
}
