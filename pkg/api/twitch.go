package api

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/internal/locking"
	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
)

type TwitchAuthZHandler struct {
	redirectURL string
	state       string
	client      *helix.Client
	userStore   db.UserStore
}

func NewTwitchAuthZHandler(url, state string, c *helix.Client, userStore db.UserStore) *TwitchAuthZHandler {
	return &TwitchAuthZHandler{
		redirectURL: url,
		state:       state,
		client:      c,
		userStore:   userStore,
	}
}

func (h *TwitchAuthZHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	// https://dev.twitch.tv/docs/authentication/getting-tokens-oauth/
	if r.URL.Query().Has("error") {
		log.Printf("failed to authorize: %s\n", r.URL.Query().Get("error_description"))
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "failed to authorize")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Println("could not extract access code from redirect")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "failed to authorize")
		return
	}

	// validate state key matches
	state := r.URL.Query().Get("state")
	if state != h.state {
		log.Println("could not verify request state")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "failed to verify state")
		return
	}

	// https://dev.twitch.tv/docs/authentication/getting-tokens-oauth/#use-the-authorization-code-to-get-a-token
	token, err := h.client.RequestUserAccessToken(code)
	if err != nil {
		log.Printf("failed to retrieve user access token: %v\n", err)
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "failed to authorize")
		return
	}
	log.Printf("token response: HTTP %d; %s", token.StatusCode, token.ErrorMessage)

	// authorize for this call
	locking.TwitchClientLock.Lock()
	defer locking.TwitchClientLock.Unlock()
	h.client.SetUserAccessToken(token.Data.AccessToken)

	ok, data, err := h.client.ValidateToken(token.Data.AccessToken)
	if err != nil {
		log.Println("error occurred while validating Twitch OAuth token", err)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "failed to validate Twitch OAuth token")
		return
	} else if !ok {
		log.Printf("failed to validate. Error Status: %d; Message: %s\n", data.ErrorStatus, data.ErrorMessage)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "failed to validate Twitch OAuth token")
		return
	}

	log.Println("validated", data.Data.Login)

	user := users.User{
		TwitchID:           data.Data.UserID,
		TwitchAccessToken:  token.Data.AccessToken,
		TwitchRefreshToken: token.Data.RefreshToken,
	}

	err = h.userStore.AddUser(&user)
	if err != nil {
		log.Printf("failed to store auth details for %s\n", user.TwitchID)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to save details on user")
		return // don't set a cookie on the client
	}

	// add a cookie in the response
	twitchCookie := http.Cookie{
		Name:     constants.TwitchIDCookieKey,
		Path:     "/",
		Value:    base64.StdEncoding.EncodeToString([]byte(user.TwitchID)),
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &twitchCookie)

	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
