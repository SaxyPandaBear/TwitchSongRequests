package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
)

type UserHandler struct {
	data db.UserStore
}

type userResponse struct {
	TwitchID string `json:"twitch_id"`
}

func NewUserHandler(d db.UserStore) *UserHandler {
	return &UserHandler{
		data: d,
	}
}

// TODO: this is just to try to test CORS
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
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
	u, err := h.data.GetUser(userID)
	if err != nil {
		log.Println("failed to get user", err)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "failed to get user")
		return
	}

	resp := userResponse{TwitchID: u.TwitchID}

	bytes, err = json.Marshal(resp)
	if err != nil {
		log.Println("failed to marshal response", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to marshal response")
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = fmt.Fprintln(w, string(bytes)); err != nil {
		log.Println("failed to write response", err)
	}
}
