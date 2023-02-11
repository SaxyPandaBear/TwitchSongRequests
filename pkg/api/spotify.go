package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type SpotifyAuthZHandler struct {
	redirectURL   string
	state         string
	authenticator *spotifyauth.Authenticator
	userStore     db.UserStore
}

func NewSpotifyAuthZHandler(url, state string, auth *spotifyauth.Authenticator, userStore db.UserStore) *SpotifyAuthZHandler {
	return &SpotifyAuthZHandler{
		redirectURL:   url,
		state:         state,
		authenticator: auth,
	}
}

// https://developer.spotify.com/documentation/general/guides/authorization/code-flow/
func (h *SpotifyAuthZHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(constants.TwitchIDCookieKey)
	if err != nil {
		// There is no point in authenticatingwith Spotify because
		// the Twitch user ID is the primary key for a user
		log.Println("Twitch ID is not available")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "missing Twitch ID cookie")
		return
	}

	var userID string
	if err = cookie.Valid(); err == nil {
		userID = cookie.Value
	}

	// TODO: get Twitch user ID, grab Spotify OAuth token, and persist
	token, err := h.authenticator.Token(r.Context(), "", r)
	if err != nil {
		log.Println("failed to get spotify auth token for user", userID)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "failed to get Spotify auth token")
	}

	if token != nil {
		log.Println("successfully got Spotify token")
	}

	log.Println("TODO: store information on user", userID)
	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
