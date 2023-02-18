package site

import (
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
)

var (
	homePage = template.Must(template.ParseFiles("pkg/site/home.html"))
)

type AuthConfig struct {
	ClientID    string
	RedirectURL string
	State       string
}

type SiteRenderer struct {
	userStore db.UserStore
	twitch    AuthConfig
}

type HomePageData struct {
	TwitchAuthURL string
	Authenticated bool
}

func NewSiteRenderer(u db.UserStore, twitch AuthConfig) *SiteRenderer {
	return &SiteRenderer{
		userStore: u,
		twitch:    twitch,
	}
}

func (h *SiteRenderer) HomePage(w http.ResponseWriter, r *http.Request) {
	data := h.getHomePageData(r)

	if err := homePage.Execute(w, data); err != nil {
		log.Println("error occurred while executing template:", err)
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404) // TODO: use a custom 404 page HTML template
	if _, err := w.Write([]byte("page does not exist")); err != nil {
		log.Println("error occurred while writing response:", err)
	}
}

func NotAllowed(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(405)
	if _, err := w.Write([]byte("method is not valid")); err != nil {
		log.Println("error occurred while writing response:", err)
	}
}

func (h *SiteRenderer) getHomePageData(r *http.Request) *HomePageData {
	d := HomePageData{
		TwitchAuthURL: GenerateTwitchAuthorizeURL(h.twitch),
		Authenticated: false,
	}

	c, err := r.Cookie(constants.TwitchIDCookieKey)
	if err != nil {
		log.Println("cookie not found", err)
		return &d
	}

	if err = c.Valid(); err != nil {
		log.Println("cookie expired", err)
		return &d
	}

	idBytes, err := base64.StdEncoding.DecodeString(c.Value)
	if err != nil {
		log.Println("failed to decode cookie value", err)
		return &d
	}
	id := string(idBytes)

	var u *users.User
	u, err = h.userStore.GetUser(id)
	if err != nil {
		log.Println("failed to look up user", err)
		return &d
	} else if u == nil {
		log.Println("nil user found")
		return &d
	}

	tAuthed := len(u.TwitchAccessToken) > 0 && len(u.TwitchRefreshToken) > 0
	sAuthed := len(u.SpotifyAccessToken) > 0 && len(u.SpotifyRefreshToken) > 0
	d.Authenticated = tAuthed && sAuthed

	return &d
}

func GenerateTwitchAuthorizeURL(config AuthConfig) string {
	query := url.Values{
		"client_id":     {config.ClientID},
		"redirect_uri":  {config.RedirectURL},
		"response_type": {"code"},
		"state":         {config.State},
		"scope":         {"channel-read-redemptions"},
	}

	u := url.URL{
		Scheme:   "https",
		Host:     "id.twitch.tv",
		Path:     "oauth2/authorize",
		RawQuery: query.Encode(),
	}

	return u.String()
}
