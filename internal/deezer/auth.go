package deezer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type session struct {
	apiToken     string
	licenseToken string
	client       *http.Client
}

func authenticate(ctx context.Context, arl string) (*session, error) {
	jar, _ := cookiejar.New(nil)
	c := &http.Client{Timeout: 20 * time.Second, Jar: jar}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.deezer.com/ajax/gw-light.php?method=deezer.getUserData&input=3&api_version=1.0&api_token=", nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{Name: "arl", Value: arl})

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

	return &session{
		apiToken:     data.Results.CheckForm,
		licenseToken: data.Results.User.Options.LicenseToken,
		client:       c,
	}, nil
}
