package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nicklaw5/helix"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

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
	var success = true

	// TODO: debugging
	log.Println("form values")
	e := r.ParseForm()
	if e != nil {
		log.Println("oopsie", e)
	} else {
		for k, v := range r.Form {
			log.Printf("%s: %v\n", k, v)
		}
	}

	if r.URL.Query().Has("error") {
		log.Printf("failed to authorize: %s\n", r.URL.Query().Get("error_description"))
		success = false
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Println("could not extract access code from redirect")
		success = false
	} else {
		// https://dev.twitch.tv/docs/authentication/getting-tokens-oauth/#use-the-authorization-code-to-get-a-token
		token, err := h.twitch.RequestUserAccessToken(code)
		if err != nil {
			log.Printf("failed to retrieve user access token: %v\n", err)
			success = false
		}
		if token != nil {
			log.Println("successfully got user access token")
			// TODO: store
		}
	}

	_, err := w.Write([]byte(fmt.Sprintf("twitch: %v", success)))
	if err != nil {
		log.Println("failed to include payload", err)
	}
	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}

// https://developer.spotify.com/documentation/general/guides/authorization/code-flow/
func (h *OAuthRedirectHandler) HandleSpotifyRedirect(w http.ResponseWriter, r *http.Request) {
	var success = true
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
