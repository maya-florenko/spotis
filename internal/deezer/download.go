package deezer

import (
	"bytes"
	"context"
	"crypto/cipher"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/crypto/blowfish"
)

const chunkSize = 2048

var iv = []byte{0, 1, 2, 3, 4, 5, 6, 7}

func DownloadTrackFromURL(ctx context.Context, url string) (*bytes.Buffer, error) {
	trackID := extractTrackID(url)
	if trackID == "" {
		return nil, fmt.Errorf("invalid track URL")
	}

	s, err := authenticate(ctx, os.Getenv("DEEZER_ARL"))
	if err != nil {
		return nil, err
	}

	track, err := fetchTrack(ctx, s, trackID)
	if err != nil {
		return nil, err
	}

	mediaURL, err := fetchMediaURL(ctx, s, track)
	if err != nil {
		return nil, err
	}

	return downloadTrack(ctx, s, mediaURL, track)
}

func downloadTrack(ctx context.Context, s *session, url string, track *song) (*bytes.Buffer, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", res.StatusCode)
	}

	key, err := decryptionKey(os.Getenv("DEEZER_SECRET"), track.ID)
	if err != nil {
		return nil, err
	}

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

func decryptionKey(secret, id string) ([]byte, error) {
	if len(secret) < 16 {
		return nil, fmt.Errorf("DEEZER_SECRET is too short")
	}

	hash := md5.Sum([]byte(id))
	hex := fmt.Sprintf("%x", hash)

	key := make([]byte, 16)
	copy(key, secret)
	for i := range hash {
		key[i] ^= hex[i] ^ hex[i+16]
	}

	return key, nil
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
