package api

import (
	"log"
	"net/http"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type SpotifyAuthZHandler struct {
	redirectURL   string
	authenticator *spotifyauth.Authenticator
}

func NewSpotifyAuthZHandler(url string, auth *spotifyauth.Authenticator) *SpotifyAuthZHandler {
	return &SpotifyAuthZHandler{
		redirectURL:   url,
		authenticator: auth,
	}
}

// https://developer.spotify.com/documentation/general/guides/authorization/code-flow/
func (h *SpotifyAuthZHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	// TODO: implement and get code
	token, err := h.authenticator.Token(r.Context(), "", r)
	if err != nil {
		log.Println("failed to get spotify auth token")
	}

	client := h.authenticator.Client(r.Context(), token)
	if client == nil {
		log.Println("failed to create spotify client")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
