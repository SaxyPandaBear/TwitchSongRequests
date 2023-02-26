package api

import (
	"log"
	"net/http"
	"time"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
)

type UserHandler struct {
	data        db.UserStore
	redirectURL string
}

func NewUserHandler(d db.UserStore, redirectURL string) *UserHandler {
	return &UserHandler{
		data:        d,
		redirectURL: redirectURL,
	}
}

func (h *UserHandler) RevokeUserAccesses(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetUserIDFromRequest(r)

	if err != nil {
		log.Println("failed to get Twitch ID from request", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	// TODO: revoke API access
	// u, err := h.data.GetUser(userID)
	// if err != nil {
	// 	log.Println("failed to fetch user", err)
	// 	http.Redirect(w, r, h.redirectURL, http.StatusFound)
	// 	return
	// }

	err = h.data.DeleteUser(userID)
	if err != nil {
		log.Println("failed to delete user", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
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
