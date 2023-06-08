package api

import (
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"go.uber.org/zap"
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
		zap.L().Error("failed to get Twitch ID from request", zap.Error(err))
		w.Write([]byte(err.Error()))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	u, err := h.userStore.GetUser(userID)
	if err != nil {
		zap.L().Error("failed to get user", zap.Error(err))
		w.Write([]byte(err.Error()))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	state := r.URL.Query().Get("state")
	if state != h.auth.State {
		zap.L().Error("failed to validate auth state")
		w.Write([]byte("failed to validate auth state"))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := h.auth.OAuth.Exchange(r.Context(), code)

	if err != nil {
		zap.L().Error("failed to get spotify auth token", zap.String("id", userID), zap.Error(err))
		w.Write([]byte("failed to get spotify auth token for user"))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	log.Println("successfully got Spotify token")

	u.SpotifyAccessToken = token.AccessToken
	u.SpotifyRefreshToken = token.RefreshToken
	u.SpotifyExpiry = &token.Expiry

	client := util.GetNewSpotifyClient(r.Context(), h.auth, token)
	user, err := client.CurrentUser(r.Context())
	if err != nil {
		zap.L().Error("failed to get Spotify user", zap.String("id", userID), zap.Error(err))
	} else {
		u.Email = user.Email
	}

	err = h.userStore.UpdateUser(u)
	if err != nil {
		zap.L().Error("failed to update user", zap.Error(err))
	}

	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
