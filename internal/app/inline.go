package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/maya-florenko/spotis/internal/metadata"
	"github.com/maya-florenko/spotis/internal/songlink"
)

func InlineHandler(ctx context.Context, b *bot.Bot, u *models.Update) {
	if u.InlineQuery == nil {
		return
	}

	id, err := download(ctx, b, u.InlineQuery.Query)
	if err != nil {
		return
	}

	results := []models.InlineQueryResult{
		&models.InlineQueryResultCachedAudio{
			ID:          "1",
			AudioFileID: id,
		},
	}

	b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
		InlineQueryID: u.InlineQuery.ID,
		Results:       results,
		CacheTime:     0,
	})
}

func download(ctx context.Context, b *bot.Bot, url string) (string, error) {
	// Get track information from songlink
	trackInfo, err := songlink.GetLink(ctx, url)
	if err != nil {
		return "", err
	}

	// Download track using the download manager
	trackData, err := downloadManager.Download(ctx, trackInfo)
	if err != nil {
		return "", err
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

	filename := trackData.Artist + " - " + trackData.Title + ".mp3"
	msg, err := b.SendAudio(ctx, &bot.SendAudioParams{
		ChatID: os.Getenv("TELEGRAM_CHAT_ID"),
		Audio: &models.InputFileUpload{
			Filename: filename,
			Data:     audioWithMetadata,
		},
		Title:     trackData.Title,
		Performer: trackData.Artist,
		Thumbnail: cover(ctx, trackData.Cover),
	})
	if err != nil {
		return "", err
	}

	return msg.Audio.FileID, err
}

func cover(ctx context.Context, url string) models.InputFile {
	if url == "" {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	return &models.InputFileUpload{
		Filename: "cover.jpg",
		Data:     bytes.NewReader(data),
	}
}
