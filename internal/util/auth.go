package util

import (
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"golang.org/x/oauth2"
)

type AuthConfig struct {
	ClientID     string
	ClientSecret string
	Scope        string
	RedirectURL  string
	State        string
	APIBaseURL   string
	OAuth        *oauth2.Config
}

func GetUserIDFromRequest(r *http.Request) (string, error) {
	c, err := r.Cookie(constants.TwitchIDCookieKey)
	if err != nil {
		return "", err
	}

	if err = c.Valid(); err != nil {
		return "", err
	}

	idBytes, err := base64.StdEncoding.DecodeString(c.Value)
	if err != nil {
		return "", err
	}

	return string(idBytes), nil
}

func GenerateAuthURL(host, path string, config *AuthConfig) string {
	query := url.Values{
		"client_id":     {config.ClientID},
		"redirect_uri":  {config.RedirectURL},
		"response_type": {"code"},
		"state":         {config.State},
		"scope":         {config.Scope},
	}

	u := url.URL{
		Scheme:   "https",
		Host:     host,
		Path:     path,
		RawQuery: query.Encode(),
	}

	return u.String()
}
