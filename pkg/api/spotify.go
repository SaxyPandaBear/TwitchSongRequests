package api

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"golang.org/x/oauth2"
)

type SpotifyAuthZHandler struct {
	redirectURL   string
	state         string
	authenticator *oauth2.Config
	userStore     db.UserStore
}

func NewSpotifyAuthZHandler(url, state string, auth *oauth2.Config, userStore db.UserStore) *SpotifyAuthZHandler {
	return &SpotifyAuthZHandler{
		redirectURL:   url,
		state:         state,
		authenticator: auth,
		userStore:     userStore,
	}
}

// https://developer.spotify.com/documentation/general/guides/authorization/code-flow/
func (h *SpotifyAuthZHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(constants.TwitchIDCookieKey)
	if err != nil {
		// There is no point in authenticating with Spotify because
		// the Twitch user ID is the primary key for a user
		log.Println("Twitch ID is not available", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "missing Twitch ID cookie")
		return
	}

	if err = cookie.Valid(); err != nil {
		log.Println("cookie expired", err)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "cookie expired")
		return
	}

	bytes, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		log.Println("failed to decode cookie value", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to decode cookie value")
	}

	userID := string(bytes)
	u, err := h.userStore.GetUser(userID)
	if err != nil {
		log.Println("failed to get user", err)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "failed to get user")
		return
	}

	state := r.URL.Query().Get("state")
	if state != h.state {
		log.Println("failed to validate auth state")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "failed to validate auth state")
		return
	}

	code := r.Form.Get("code")
	token, err := h.authenticator.Exchange(r.Context(), code)

	if err != nil {
		log.Println("failed to get spotify auth token for user", userID)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "failed to get Spotify auth token")
		return
	}

	log.Println("successfully got Spotify token")

	u.SpotifyAccessToken = token.AccessToken
	u.SpotifyRefreshToken = token.RefreshToken
	u.SpotifyExpiry = &token.Expiry

	err = h.userStore.UpdateUser(u)
	if err != nil {
		log.Println("failed to update user", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to save spotify auth")
	}

	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
