package api

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
)

type UserHandler struct {
	data        db.UserStore
	redirectURL string
}

type userResponse struct {
	TwitchID string `json:"twitch_id"`
}

func NewUserHandler(d db.UserStore, redirectURL string) *UserHandler {
	return &UserHandler{
		data:        d,
		redirectURL: redirectURL,
	}
}

func (h *UserHandler) RevokeUserAccesses(w http.ResponseWriter, r *http.Request) {
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
		return
	}

	userID := string(bytes)
	err = h.data.DeleteUser(userID)
	if err != nil {
		log.Println("failed to delete user", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to delete user")
		return
	}

	log.Println("successfully deleted user", userID)

	twitchCookie := http.Cookie{
		Name:     constants.TwitchIDCookieKey,
		Path:     "/",
		Value:    "",
		MaxAge:   -1, // delete the cookie
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &twitchCookie)
	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
