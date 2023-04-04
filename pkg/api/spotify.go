package api

import (
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
)

type SpotifyAuthZHandler struct {
	redirectURL string
	auth        *util.AuthConfig
	userStore   db.UserStore
}

func NewSpotifyAuthZHandler(url string, auth *util.AuthConfig, userStore db.UserStore) *SpotifyAuthZHandler {
	return &SpotifyAuthZHandler{
		redirectURL: url,
		auth:        auth,
		userStore:   userStore,
	}
}

// https://developer.spotify.com/documentation/general/guides/authorization/code-flow/
func (h *SpotifyAuthZHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetUserIDFromRequest(r)
	if err != nil {
		log.Println("failed to get Twitch ID from request", err)
		w.Write([]byte(err.Error()))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	u, err := h.userStore.GetUser(userID)
	if err != nil {
		log.Println("failed to get user", err)
		w.Write([]byte(err.Error()))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	state := r.URL.Query().Get("state")
	if state != h.auth.State {
		log.Println("failed to validate auth state")
		w.Write([]byte("failed to validate auth statee"))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := h.auth.OAuth.Exchange(r.Context(), code)

	if err != nil {
		log.Println("failed to get spotify auth token for user", userID)
		w.Write([]byte("failed to get spotify auth token for user"))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	log.Println("successfully got Spotify token")

	u.SpotifyAccessToken = token.AccessToken
	u.SpotifyRefreshToken = token.RefreshToken
	u.SpotifyExpiry = &token.Expiry

	err = h.userStore.UpdateUser(u)
	if err != nil {
		log.Println("failed to update user", err)
	}

	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
