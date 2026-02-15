package spotify

import (
	"context"
	"os"

	"github.com/zmb3/spotify/v2"
	auth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var SpotifyClient *spotify.Client

func Init(ctx context.Context) error {
	conf := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		TokenURL:     auth.TokenURL,
	}
	token, err := conf.Token(ctx)
	if err != nil {
		return err
	}

	httpClient := auth.New().Client(ctx, token)
	SpotifyClient = spotify.New(httpClient)

	return nil
}

func Get(ctx context.Context, url string) (*spotify.FullTrack, error) {
	track, err := SpotifyClient.GetTrack(ctx, spotify.ID(extractID(url)))
	if err != nil {
		return nil, err
	}

	return track, err
}
