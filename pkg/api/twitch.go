package api

import (
	"crypto/cipher"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
)

type TwitchAuthZHandler struct {
	redirectURL string
	state       string
	client      *helix.Client
	userStore   db.UserStore
	gcm         cipher.AEAD
}

var (
	twitchMu sync.Mutex
)

func NewTwitchAuthZHandler(url, state string, c *helix.Client, userStore db.UserStore, gcm cipher.AEAD) *TwitchAuthZHandler {
	return &TwitchAuthZHandler{
		redirectURL: url,
		state:       state,
		client:      c,
		userStore:   userStore,
		gcm:         gcm,
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
		TwitchID:           data.Data.UserID,
		TwitchAccessToken:  token.Data.AccessToken,
		TwitchRefreshToken: token.Data.RefreshToken,
	}

	// encrypt the cookie value before saving to the database in case this fails
	// because we don't want to pollute the database
	encrypted, err := util.EncryptTwitchID(user.TwitchID, h.gcm, util.DefaultNonceGenerator)
	if err != nil {
		log.Printf("failed to encrypt %s: %v\n", user.TwitchID, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to encrypt user ID")
		return
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
		Name:  constants.TwitchIDCookieKey,
		Value: string(encrypted),
	}
	http.SetCookie(w, &twitchCookie)

	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
