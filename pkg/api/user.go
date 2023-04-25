package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/saxypandabear/twitchsongrequests/internal/constants"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
)

type UserHandler struct {
	users       db.UserStore
	prefs       db.PreferenceStore
	redirectURL string
	twitch      *util.AuthConfig
	spotify     *util.AuthConfig
}

func NewUserHandler(d db.UserStore, p db.PreferenceStore, redirectURL string, twitch, spotify *util.AuthConfig) *UserHandler {
	return &UserHandler{
		users:       d,
		prefs:       p,
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

	c, err := util.GetNewTwitchClient(h.twitch)
	if err != nil {
		log.Println("failed to get Twitch client", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	u, err := h.users.GetUser(userID)
	if err != nil {
		log.Println("failed to get user", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	// revoke subscription first
	appToken, err := c.RequestAppAccessToken(strings.Split(h.twitch.Scope, " "))
	if err != nil {
		log.Println("failed to get app access token", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}
	c.SetAppAccessToken(appToken.Data.AccessToken)
	res, err := c.RemoveEventSubSubscription(u.SubscriptionID)
	if err != nil {
		log.Println("failed to remove eventsub subscription for", userID, err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	} else if len(res.ErrorMessage) > 0 {
		log.Printf("failed to remove eventsub subscription for %s: HTTP %d | %s | %s", userID, res.ErrorStatus, res.Error, res.ErrorMessage)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	tok, err := db.FetchTwitchToken(h.users, userID)
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

	err = h.users.DeleteUser(userID)
	if err != nil {
		log.Println("failed to delete user", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	log.Println("successfully deleted user", userID)

	err = h.prefs.DeletePreference(userID)
	if err != nil {
		// I'm not sure if this is fatal or not.
		log.Println("failed to delete user", err)
	}

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
