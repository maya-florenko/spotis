package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/maya-florenko/spotis/internal/spotify"
)

func InlineHandler(ctx context.Context, b *bot.Bot, u *models.Update) {
	if u.InlineQuery == nil {
		return
	}

	res, err := spotify.Get(ctx, u.InlineQuery.Query)
	if err != nil {
		return
	}

	results := []models.InlineQueryResult{
		&models.InlineQueryResultArticle{
			ID:    "1",
			Title: res.Artists[0].Name + " - " + res.Name,
			InputMessageContent: &models.InputTextMessageContent{
				MessageText: res.Artists[0].Name + " - " + res.Name,
			},
			ThumbnailURL: res.Album.Images[0].URL,
		},
	}

	b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
		InlineQueryID: u.InlineQuery.ID,
		Results:       results,
		CacheTime:     0,
	})
}
