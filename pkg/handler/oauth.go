package handler

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/nicklaw5/helix"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var mu sync.Mutex

type OAuthRedirectHandler struct {
	redirectURL string
	spotify     *spotifyauth.Authenticator
	twitch      *helix.Client
}

func NewOAuthRedirectHandler(uri string, spotify *spotifyauth.Authenticator, twitch *helix.Client) *OAuthRedirectHandler {
	return &OAuthRedirectHandler{
		redirectURL: uri,
		spotify:     spotify,
		twitch:      twitch,
	}
}

// https://dev.twitch.tv/docs/authentication/getting-tokens-oauth/
func (h *OAuthRedirectHandler) HandleTwitchRedirect(w http.ResponseWriter, r *http.Request) {
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
	token, err := h.twitch.RequestUserAccessToken(code)
	if err != nil {
		log.Printf("failed to retrieve user access token: %v\n", err)
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "failed to authorize")
		return
	}
	log.Printf("token response: HTTP %d; %s", token.StatusCode, token.ErrorMessage)

	// authorize for this call
	mu.Lock()
	defer mu.Unlock()
	h.twitch.SetUserAccessToken(token.Data.AccessToken)

	ok, data, err := h.twitch.ValidateToken(token.Data.AccessToken)
	if err != nil {
		log.Println("oops", err)
	} else if !ok {
		log.Printf("failed to validate. Error Status: %d; Message: %s\n", data.ErrorStatus, data.ErrorMessage)
	} else if data != nil {
		log.Println("validated", data.Data.Login)
	}

	// TODO: store
	log.Println("store something")
	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}

// https://developer.spotify.com/documentation/general/guides/authorization/code-flow/
func (h *OAuthRedirectHandler) HandleSpotifyRedirect(w http.ResponseWriter, r *http.Request) {
	var success = true
	// TODO: get code
	token, err := h.spotify.Token(r.Context(), "", r)
	if err != nil {
		log.Printf("failed to retrieve Spotify token: %v\n", err)
		success = false
	}

	if token != nil {
		log.Println("successfully retrieved Spotify token")
		// TODO: store
	}

	_, err = w.Write([]byte(fmt.Sprintf("spotify: %v", success)))
	if err != nil {
		log.Println("failed to include payload", err)
	}
	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
