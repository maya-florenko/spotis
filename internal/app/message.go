package app

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/maya-florenko/spotis/internal/deezer"
	"github.com/maya-florenko/spotis/internal/songlink"
)

func MessageHandler(ctx context.Context, b *bot.Bot, u *models.Update) {
	if u.Message == nil {
		return
	}

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: u.Message.Chat.ID,
		Action: models.ChatActionRecordVoice,
	})

	res, err := songlink.GetLink(ctx, u.Message.Text)
	if err != nil {
		return
	}

	file, err := deezer.DownloadTrackFromURL(ctx, res.URL)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: u.Message.Chat.ID,
			Text:   err.Error(),
		})
	}

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: u.Message.Chat.ID,
		Action: models.ChatActionUploadVoice,
	})

	b.SendAudio(ctx, &bot.SendAudioParams{
		ChatID: u.Message.Chat.ID,
		Audio: &models.InputFileUpload{
			Filename: res.Artist + " - " + res.Title,
			Data:     file,
		},
		Title:     res.Artist + " - " + res.Title,
		Thumbnail: cover(ctx, res.Cover),
	})
}
