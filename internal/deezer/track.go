package deezer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Song struct {
	ID         string `json:"SNG_ID"`
	Artist     string `json:"ART_NAME"`
	Title      string `json:"SNG_TITLE"`
	Version    string `json:"VERSION"`
	Cover      string `json:"ALB_PICTURE"`
	TrackToken string `json:"TRACK_TOKEN"`
	Duration   string `json:"DURATION"`
}

type TrackResponse struct {
	Results struct {
		Data *Song `json:"DATA"`
	} `json:"results"`
}

type Media struct {
	Data []struct {
		Media []struct {
			Format  string `json:"format"`
			Sources []struct {
				URL string `json:"url"`
			} `json:"sources"`
		} `json:"media"`
		Errors []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	} `json:"data"`
	Errors []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func FetchTrack(ctx context.Context, session *Session, trackID string) (*Song, error) {
	payload := map[string]any{
		"sng_id": trackID,
		"nb":     10000,
		"start":  0,
		"lang":   "en",
	}

	jsonData, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://www.deezer.com/ajax/gw-light.php?method=deezer.pageTrack&input=3&api_version=1.0&api_token=%s", session.APIToken)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	res, err := session.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	var track TrackResponse
	json.Unmarshal(body, &track)

	if track.Results.Data == nil {
		return nil, fmt.Errorf("track not found")
	}

	return track.Results.Data, nil
}

func FetchMediaURL(ctx context.Context, session *Session, song *Song, quality string) (*Media, error) {
	formats := fmt.Sprintf(`[{"cipher":"BF_CBC_STRIPE","format":"%s"}]`, quality)
	reqBody := fmt.Sprintf(`{"license_token":"%s","media":[{"type":"FULL","formats":%s}],"track_tokens":["%s"]}`,
		session.LicenseToken, formats, song.TrackToken)

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://media.deezer.com/v1/get_url", bytes.NewBuffer([]byte(reqBody)))
	res, err := session.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	var media Media
	json.Unmarshal(body, &media)

	if len(media.Errors) > 0 {
		return nil, fmt.Errorf("error: %s", media.Errors[0].Message)
	}

	if len(media.Data[0].Errors) > 0 {
		return nil, fmt.Errorf("error: %s", media.Data[0].Errors[0].Message)
	}

	return &media, nil
}

func (m *Media) GetURL() string {
	return m.Data[0].Media[0].Sources[0].URL
}

func (m *Media) GetFormat() string {
	return m.Data[0].Media[0].Format
}
