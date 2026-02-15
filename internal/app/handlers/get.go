package handlers

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func GetHandler(ctx context.Context, b *bot.Bot, u *models.Update) {
	log.Print("meow")
}
