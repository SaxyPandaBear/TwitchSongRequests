package api

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
)

type TwitchAuthZHandler struct {
	redirectURL string
	client      *helix.Client
}

var twitchMu sync.Mutex

func NewTwitchAuthZHandler(url string, c *helix.Client) *TwitchAuthZHandler {
	return &TwitchAuthZHandler{
		redirectURL: url,
		client:      c,
	}
}

func (h *TwitchAuthZHandler) SubscribeToTopic(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "failed to authorize")
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
	twitchMu.Lock()
	defer twitchMu.Unlock()
	h.client.SetUserAccessToken(token.Data.AccessToken)

	ok, data, err := h.client.ValidateToken(token.Data.AccessToken)
	if err != nil {
		log.Println("oops", err)
	} else if !ok {
		log.Printf("failed to validate. Error Status: %d; Message: %s\n", data.ErrorStatus, data.ErrorMessage)
	} else if data != nil {
		log.Println("validated", data.Data.Login)
	}

	// TODO: store
	user := users.User{
		TwitchID: data.Data.UserID,
		TwitchAuth: users.TwitchAuth{
			AccessToken:  token.Data.AccessToken,
			RefreshToken: token.Data.RefreshToken,
		},
	}
	log.Println("TODO: store information on", user.TwitchID)
	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
