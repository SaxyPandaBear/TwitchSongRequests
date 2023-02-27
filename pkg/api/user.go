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
	twitch      *util.AuthConfig
	spotify     *util.AuthConfig
}

func NewUserHandler(d db.UserStore, redirectURL string, twitch, spotify *util.AuthConfig) *UserHandler {
	return &UserHandler{
		data:        d,
		redirectURL: redirectURL,
		twitch:      twitch,
		spotify:     spotify,
	}
}

func (h *UserHandler) RevokeUserAccesses(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetUserIDFromRequest(r)

	if err != nil {
		log.Println("failed to get Twitch ID from request", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	// u, err := h.data.GetUser(userID)
	// if err != nil {
	// 	log.Println("failed to fetch user", err)
	// 	http.Redirect(w, r, h.redirectURL, http.StatusFound)
	// 	return
	// }
	c, err := util.GetNewTwitchClient(h.twitch)
	if err != nil {
		log.Println("failed to get Twitch client", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	tok, err := db.FetchTwitchToken(h.data, userID)
	if err != nil {
		log.Println("failed to get user token", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	c.SetUserAccessToken(tok.AccessToken)
	tokenResp, err := c.RefreshUserAccessToken(tok.RefreshToken)
	if err != nil {
		log.Println("failed to refresh twitch token", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	// make sure that the token is fresh before revoking
	_, err = c.RevokeUserAccessToken(tokenResp.Data.AccessToken)
	if err != nil {
		log.Println("failed to revoke access", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

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
