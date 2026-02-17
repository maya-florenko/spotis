package downloader

import (
	"bytes"
	"context"
	"crypto/cipher"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"

	"github.com/maya-florenko/spotis/internal/songlink"
	"golang.org/x/crypto/blowfish"
)

const (
	chunkSize = 2048
	quality   = "MP3_128"
)

var iv = []byte{0, 1, 2, 3, 4, 5, 6, 7}

// DeezerDownloader implements the Downloader interface for Deezer
type DeezerDownloader struct {
	arl    string
	secret string
}

// NewDeezerDownloader creates a new Deezer downloader
func NewDeezerDownloader(arl, secret string) *DeezerDownloader {
	return &DeezerDownloader{
		arl:    arl,
		secret: secret,
	}
}

// NewDeezerDownloaderFromEnv creates a new Deezer downloader from environment variables
func NewDeezerDownloaderFromEnv() *DeezerDownloader {
	return NewDeezerDownloader(
		os.Getenv("DEEZER_ARL"),
		os.Getenv("DEEZER_SECRET"),
	)
}

// Platform returns the platform this downloader supports
func (d *DeezerDownloader) Platform() songlink.Platform {
	return songlink.PlatformDeezer
}

// CanDownload checks if this downloader can handle the given URL
func (d *DeezerDownloader) CanDownload(url string) bool {
	return strings.Contains(url, "deezer.com")
}

// Download downloads a track from Deezer
func (d *DeezerDownloader) Download(ctx context.Context, url string) (*bytes.Buffer, error) {
	trackID := extractTrackID(url)
	if trackID == "" {
		return nil, fmt.Errorf("could not extract track ID from URL")
	}

	session, err := d.authenticate(ctx)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	track, err := d.fetchTrack(ctx, session, trackID)
	if err != nil {
		return nil, fmt.Errorf("fetch track: %w", err)
	}

	mediaURL, err := d.fetchMediaURL(ctx, session, track)
	if err != nil {
		return nil, fmt.Errorf("fetch media URL: %w", err)
	}

	return d.downloadTrack(ctx, session, mediaURL, track)
}

type deezerSession struct {
	apiToken     string
	licenseToken string
	client       *http.Client
}

type song struct {
	ID         string `json:"SNG_ID"`
	TrackToken string `json:"TRACK_TOKEN"`
}

func (d *DeezerDownloader) authenticate(ctx context.Context) (*deezerSession, error) {
	jar, _ := cookiejar.New(nil)
	c := &http.Client{Timeout: 20 * time.Second, Jar: jar}

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.deezer.com/ajax/gw-light.php?method=deezer.getUserData&input=3&api_version=1.0&api_token=", nil)
	req.AddCookie(&http.Cookie{Name: "arl", Value: d.arl})

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data struct {
		Results struct {
			CheckForm string `json:"checkForm"`
			User      struct {
				ID      int `json:"USER_ID"`
				Options struct {
					LicenseToken string `json:"license_token"`
				} `json:"OPTIONS"`
			} `json:"USER"`
		} `json:"results"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Results.User.ID == 0 {
		return nil, fmt.Errorf("invalid arl cookie")
	}

	return &deezerSession{
		apiToken:     data.Results.CheckForm,
		licenseToken: data.Results.User.Options.LicenseToken,
		client:       c,
	}, nil
}

func (d *DeezerDownloader) fetchTrack(ctx context.Context, s *deezerSession, id string) (*song, error) {
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

func (d *DeezerDownloader) fetchMediaURL(ctx context.Context, s *deezerSession, track *song) (string, error) {
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

func (d *DeezerDownloader) downloadTrack(ctx context.Context, s *deezerSession, url string, track *song) (*bytes.Buffer, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", res.StatusCode)
	}

	key := d.decryptionKey(track.ID)
	chunk := make([]byte, chunkSize)
	buf := new(bytes.Buffer)

	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		n, err := io.ReadFull(res.Body, chunk)
		if n > 0 {
			if i%3 == 0 && n == chunkSize {
				dec, decErr := decrypt(chunk, key)
				if decErr != nil {
					return nil, decErr
				}
				buf.Write(dec)
			} else {
				buf.Write(chunk[:n])
			}
		}

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return buf, nil
}

func (d *DeezerDownloader) decryptionKey(id string) []byte {
	hash := md5.Sum([]byte(id))
	hex := fmt.Sprintf("%x", hash)

	key := []byte(d.secret)
	for i := range hash {
		key[i] ^= hex[i] ^ hex[i+16]
	}

	return key
}

func decrypt(data, key []byte) ([]byte, error) {
	block, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}

	out := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, data)

	return out, nil
}

func extractTrackID(url string) string {
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] < '0' || url[i] > '9' {
			return url[i+1:]
		}
	}

	return ""
}
