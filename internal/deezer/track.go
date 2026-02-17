package deezer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const quality = "MP3_128"

type song struct {
	ID         string `json:"SNG_ID"`
	TrackToken string `json:"TRACK_TOKEN"`
}

func fetchTrack(ctx context.Context, s *session, id string) (*song, error) {
	body, _ := json.Marshal(map[string]any{"sng_id": id})
	url := fmt.Sprintf("https://www.deezer.com/ajax/gw-light.php?method=deezer.pageTrack&input=3&api_version=1.0&api_token=%s", s.apiToken)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var resp struct {
		Results struct {
			Data *song `json:"DATA"`
		} `json:"results"`
	}

	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if resp.Results.Data == nil {
		return nil, fmt.Errorf("track not found")
	}

	return resp.Results.Data, nil
}

func fetchMediaURL(ctx context.Context, s *session, track *song) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"license_token": s.licenseToken,
		"media": []map[string]any{{
			"type":    "FULL",
			"formats": []map[string]string{{"cipher": "BF_CBC_STRIPE", "format": quality}},
		}},
		"track_tokens": []string{track.TrackToken},
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://media.deezer.com/v1/get_url", bytes.NewReader(body))
	res, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var resp struct {
		Data []struct {
			Media []struct {
				Sources []struct {
					URL string `json:"url"`
				} `json:"sources"`
			} `json:"media"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return "", err
	}

	if len(resp.Errors) > 0 {
		return "", fmt.Errorf("media error: %s", resp.Errors[0].Message)
	}

	if len(resp.Data[0].Errors) > 0 {
		return "", fmt.Errorf("media error: %s", resp.Data[0].Errors[0].Message)
	}

	return resp.Data[0].Media[0].Sources[0].URL, nil
}
