package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"go.uber.org/zap"
)

type QueueHandler struct {
	spotify   *util.AuthConfig
	userStore db.UserStore
}

func NewQueueHandler(u db.UserStore, spotify *util.AuthConfig) *QueueHandler {
	return &QueueHandler{
		userStore: u,
		spotify:   spotify,
	}
}

func (h *QueueHandler) GetUserQueue(w http.ResponseWriter, r *http.Request) {
	// This ID parameter is expected to be the b64 encoding for the user's user ID. This isn't great or
	// really opaque, but it's good enough.
	id := chi.URLParam(r, "id")
	if id == "" {
		zap.L().Warn("Unable to associate user ID with request", zap.String("path", r.URL.Path))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		zap.L().Warn("Unable to decode ID", zap.String("encoded", id))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID := string(decoded)

	tok, err := db.FetchSpotifyToken(h.userStore, userID)
	if err != nil {
		zap.L().Error("failed to verify user for spotify access", zap.String("id", userID), zap.Error(err))
		w.WriteHeader(http.StatusOK)
		return
	}

	refreshed, err := util.RefreshSpotifyToken(r.Context(), h.spotify, tok)
	if err != nil {
		zap.L().Error("failed to get valid token", zap.Error(err))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "failed to get Spotify token")
		return
	}

	// try to offload this to see if the response time is better
	defer func() {
		// store the refreshed token
		u, err := h.userStore.GetUser(userID)
		if err == nil {
			u.SpotifyAccessToken = refreshed.AccessToken
			u.SpotifyRefreshToken = refreshed.RefreshToken
			u.SpotifyExpiry = &refreshed.Expiry

			zap.L().Debug("saving updated Spotify credentials", zap.String("id", u.TwitchID))

			if err = h.userStore.UpdateUser(u); err != nil {
				// if we got a valid token but failed to update the DB this is not necessarily fatal.
				zap.L().Error("failed to update user's spotify token", zap.Error(err))
			}
		}
	}()

	c := util.GetNewSpotifyClient(r.Context(), h.spotify, refreshed)

	q, err := c.GetQueue(context.Background())
	if err != nil {
		zap.L().Error("failed to get Spotify queue for user", zap.String("id", userID), zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "failed to get Spotify queue")
		return
	}

	bytes, err := json.Marshal(q)
	if err != nil {
		zap.L().Warn("failed to marshal response JSON", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to generate response")
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(bytes); err != nil {
		zap.L().Warn("failed to write response", zap.Error(err))
	}
}
