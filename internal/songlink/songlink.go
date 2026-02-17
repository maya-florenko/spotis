package songlink

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Platform represents a music streaming platform
type Platform string

const (
	PlatformDeezer       Platform = "deezer"
	PlatformYandex       Platform = "yandex"
	PlatformTidal        Platform = "tidal"
	PlatformYouTube      Platform = "youtube"
	PlatformYouTubeMusic Platform = "youtubeMusic"
	PlatformSpotify      Platform = "spotify"
	PlatformAppleMusic   Platform = "appleMusic"
)

// PlatformLink contains information about a track on a specific platform
type PlatformLink struct {
	Platform Platform `json:"platform"`
	URL      string   `json:"url"`
}

// TrackInfo contains comprehensive information about a track
type TrackInfo struct {
	Artist             string         `json:"artist"`
	Title              string         `json:"title"`
	Cover              string         `json:"cover"`
	Platform           Platform       `json:"platform"`            // The chosen platform
	URL                string         `json:"url"`                 // URL from the chosen platform
	AvailablePlatforms []PlatformLink `json:"available_platforms"` // All available platforms
}

type apiResponse struct {
	LinksByPlatform    map[string]platformLink           `json:"linksByPlatform"`
	EntitiesByUniqueId map[string]entitiesByUniqueIdItem `json:"entitiesByUniqueId"`
}

type platformLink struct {
	EntityUniqueId string `json:"entityUniqueId"`
	URL            string `json:"url"`
}

type entitiesByUniqueIdItem struct {
	Title        string `json:"title,omitempty"`
	ArtistName   string `json:"artistName,omitempty"`
	ThumbnailUrl string `json:"thumbnailUrl,omitempty"`
}

// GetLink fetches track information from song.link API
// It returns the track info with the highest priority available platform
func GetLink(ctx context.Context, raw string) (*TrackInfo, error) {
	u, err := url.Parse("https://api.song.link/v1-alpha.1/links")
	if err != nil {
		return nil, fmt.Errorf("parse API URL: %w", err)
	}

	q := u.Query()
	q.Set("url", raw)
	q.Set("userCountry", "US")
	q.Set("songIfSingle", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("song.link API returned status %d", resp.StatusCode)
	}

	var body apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Priority order for downloading platforms
	priority := []Platform{
		PlatformDeezer,
		PlatformYandex,
		PlatformTidal,
		PlatformYouTube,
		PlatformYouTubeMusic,
		PlatformSpotify,
		PlatformAppleMusic,
	}

	// Collect all available platforms
	availablePlatforms := make([]PlatformLink, 0, len(body.LinksByPlatform))
	for platformStr, link := range body.LinksByPlatform {
		if link.URL != "" {
			availablePlatforms = append(availablePlatforms, PlatformLink{
				Platform: Platform(platformStr),
				URL:      link.URL,
			})
		}
	}

	if len(availablePlatforms) == 0 {
		return nil, fmt.Errorf("no platforms available for this track")
	}

	// Find the highest priority platform
	var chosen platformLink
	var chosenPlatform Platform
	found := false

	for _, p := range priority {
		if pl, ok := body.LinksByPlatform[string(p)]; ok && pl.URL != "" {
			chosen = pl
			chosenPlatform = p
			found = true
			break
		}
	}

	// If no priority platform found, use the first available
	if !found {
		for platformStr, pl := range body.LinksByPlatform {
			if pl.URL != "" {
				chosen = pl
				chosenPlatform = Platform(platformStr)
				found = true
				break
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("no valid platform link found")
	}

	ti := &TrackInfo{
		URL:                chosen.URL,
		Platform:           chosenPlatform,
		AvailablePlatforms: availablePlatforms,
	}

	// Extract metadata from entities
	if ent, ok := body.EntitiesByUniqueId[chosen.EntityUniqueId]; ok {
		ti.Title = ent.Title
		ti.Artist = ent.ArtistName
		ti.Cover = ent.ThumbnailUrl
	}

	return ti, nil
}
