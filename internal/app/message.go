package app

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/maya-florenko/spotis/internal/metadata"
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

	// Get track information from songlink
	trackInfo, err := songlink.GetLink(ctx, u.Message.Text)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: u.Message.Chat.ID,
			Text:   fmt.Sprintf("Failed to find track: %v", err),
		})
		return
	}

	// Download track using the download manager
	trackData, err := downloadManager.Download(ctx, trackInfo)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: u.Message.Chat.ID,
			Text:   fmt.Sprintf("Failed to download track: %v", err),
		})
		return
	}

	// Add metadata to the audio file
	audioWithMetadata, err := metadata.AddMetadata(trackData.Audio, metadata.Metadata{
		Title:  trackData.Title,
		Artist: trackData.Artist,
		Cover:  trackData.Cover,
	})
	if err != nil {
		// If metadata fails, use original audio
		fmt.Printf("Warning: failed to add metadata: %v\n", err)
		audioWithMetadata = trackData.Audio
	}

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: u.Message.Chat.ID,
		Action: models.ChatActionUploadVoice,
	})

	// Send the audio file with metadata
	filename := trackData.Artist + " - " + trackData.Title + ".mp3"
	b.SendAudio(ctx, &bot.SendAudioParams{
		ChatID: u.Message.Chat.ID,
		Audio: &models.InputFileUpload{
			Filename: filename,
			Data:     audioWithMetadata,
		},
		Title:     trackData.Title,
		Performer: trackData.Artist,
		Thumbnail: cover(ctx, trackData.Cover),
	})
}
