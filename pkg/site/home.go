package site

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
)

var homePage = template.Must(template.ParseFiles("pkg/site/home.html"))

type HomePageRenderer struct {
	userStore db.UserStore
	twitch    *util.AuthConfig
	spotify   *util.AuthConfig
	siteURL   string
}

type HomePageData struct {
	UserID         string
	TwitchAuthURL  string
	SubscribeURL   string
	UnsubscribeURL string
	SpotifyAuthURL string
	PreferencesURL string
	Authenticated  bool
	Subscribed     bool
	Error          string
}

func NewHomePageRenderer(siteURL string, u db.UserStore, twitch, spotify *util.AuthConfig) *HomePageRenderer {
	return &HomePageRenderer{
		siteURL:   siteURL,
		userStore: u,
		twitch:    twitch,
		spotify:   spotify,
	}
}

func (h *HomePageRenderer) HomePage(w http.ResponseWriter, r *http.Request) {
	data := h.getHomePageData(r)

	log.Printf("Serving home page to %s: %v %v %s\n", data.UserID, data.Authenticated, data.Subscribed, data.Error)

	if err := homePage.Execute(w, data); err != nil {
		log.Println("error occurred while executing template:", err)
	}
}

func (h *HomePageRenderer) getHomePageData(r *http.Request) *HomePageData {
	d := HomePageData{
		SubscribeURL:   fmt.Sprintf("%s/subscribe", h.siteURL),
		UnsubscribeURL: fmt.Sprintf("%s/revoke", h.siteURL),
		PreferencesURL: fmt.Sprintf("%s/preferences", h.siteURL),
		TwitchAuthURL:  util.GenerateAuthURL("id.twitch.tv", "oauth2/authorize", h.twitch),
		SpotifyAuthURL: util.GenerateAuthURL("accounts.spotify.com", "authorize", h.spotify),
	}

	id, err := util.GetUserIDFromRequest(r)
	if err != nil {
		log.Println("failed to get Twitch ID", err)
		return &d
	}

	d.UserID = id

	user, err := h.userStore.GetUser(id)
	if err != nil {
		log.Println("failed to get user", err)
		return &d
	}

	d.Authenticated = user.IsAuthenticated()
	d.Subscribed = user.Subscribed

	// check if there is an error in the request body
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("failed to read request body", err)
	} else {
		if len(body) > 0 {
			d.Error = string(body)
			log.Println("found error in request body", d.Error)
		}
	}

	return &d
}
