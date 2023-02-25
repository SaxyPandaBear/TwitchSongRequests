package api

import (
	"encoding/base64"
	"log"
	"net/http"
	"time"

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
		log.Println("Twitch ID is not available", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
	}

	if cookie.SameSite != http.SameSiteStrictMode || !cookie.HttpOnly {
		log.Println("invalid cookie")
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
	}

	if err = cookie.Valid(); err != nil {
		log.Println("cookie expired", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
	}

	bytes, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		log.Println("failed to decode cookie value", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
	}

	userID := string(bytes)
	err = h.data.DeleteUser(userID)
	if err != nil {
		log.Println("failed to delete user", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
	}

	log.Println("successfully deleted user", userID)

	twitchCookie := http.Cookie{
		Name:     constants.TwitchIDCookieKey,
		Path:     "/",
		Value:    "",
		MaxAge:   -1, // delete the cookie
		Expires:  time.Now().Add(-1 * time.Hour),
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &twitchCookie)
	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
