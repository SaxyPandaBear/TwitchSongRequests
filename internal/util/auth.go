package util

import (
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
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

func IsUserAuthenticated(userStore db.UserStore, id string) bool {
	u, err := userStore.GetUser(id)
	if err != nil {
		log.Println("failed to look up user", err)
		return false
	} else if u == nil {
		log.Println("nil user found")
		return false
	}

	tAuthed := len(u.TwitchAccessToken) > 0 && len(u.TwitchRefreshToken) > 0
	sAuthed := len(u.SpotifyAccessToken) > 0 && len(u.SpotifyRefreshToken) > 0
	return tAuthed && sAuthed
}

func GetUserIDFromRequest(r *http.Request) (string, error) {
	c, err := r.Cookie(constants.TwitchIDCookieKey)
	if err != nil {
		return "", err
	}

	if !c.Secure || c.SameSite != http.SameSiteStrictMode {
		return "", errors.New("invalid cookie security settings")
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
