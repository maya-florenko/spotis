package downloader

import (
	"bytes"
	"context"
	"fmt"

	"github.com/maya-florenko/spotis/internal/songlink"
)

// TrackData contains the downloaded audio data and metadata
type TrackData struct {
	Audio  *bytes.Buffer
	Title  string
	Artist string
	Cover  string
}

// Downloader is an interface for downloading tracks from different platforms
type Downloader interface {
	// Platform returns the platform this downloader supports
	Platform() songlink.Platform

	// CanDownload checks if this downloader can handle the given URL
	CanDownload(url string) bool

	// Download downloads a track from the given URL
	Download(ctx context.Context, url string) (*bytes.Buffer, error)
}

// Manager manages multiple downloaders and attempts to download from available sources
type Manager struct {
	downloaders map[songlink.Platform]Downloader
}

// NewManager creates a new download manager
func NewManager() *Manager {
	return &Manager{
		downloaders: make(map[songlink.Platform]Downloader),
	}
}

// Register registers a downloader for a specific platform
func (m *Manager) Register(d Downloader) {
	m.downloaders[d.Platform()] = d
}

// Download attempts to download a track using available downloaders
// It tries platforms in the order they appear in trackInfo.AvailablePlatforms
func (m *Manager) Download(ctx context.Context, trackInfo *songlink.TrackInfo) (*TrackData, error) {
	// First, try the chosen platform
	if downloader, ok := m.downloaders[trackInfo.Platform]; ok {
		if downloader.CanDownload(trackInfo.URL) {
			audio, err := downloader.Download(ctx, trackInfo.URL)
			if err == nil {
				return &TrackData{
					Audio:  audio,
					Title:  trackInfo.Title,
					Artist: trackInfo.Artist,
					Cover:  trackInfo.Cover,
				}, nil
			}
			// Log error but continue to try other platforms
			fmt.Printf("Failed to download from %s: %v\n", trackInfo.Platform, err)
		}
	}

	// Try other available platforms
	for _, platformLink := range trackInfo.AvailablePlatforms {
		// Skip the already tried platform
		if platformLink.Platform == trackInfo.Platform {
			continue
		}

		downloader, ok := m.downloaders[platformLink.Platform]
		if !ok {
			continue
		}

		if !downloader.CanDownload(platformLink.URL) {
			continue
		}

		audio, err := downloader.Download(ctx, platformLink.URL)
		if err == nil {
			return &TrackData{
				Audio:  audio,
				Title:  trackInfo.Title,
				Artist: trackInfo.Artist,
				Cover:  trackInfo.Cover,
			}, nil
		}

		// Log error but continue to next platform
		fmt.Printf("Failed to download from %s: %v\n", platformLink.Platform, err)
	}

	return nil, fmt.Errorf("failed to download track from any available platform")
}
