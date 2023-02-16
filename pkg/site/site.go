package site

import (
	"html/template"
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
)

var (
	homePage = template.Must(template.ParseFiles("pkg/site/home.html"))
)

type SiteRenderer struct {
	userStore db.UserStore
}

type HomePageData struct {
	Authenticated bool
}

func NewSiteRenderer(u db.UserStore) *SiteRenderer {
	return &SiteRenderer{
		userStore: u,
	}
}

func (h *SiteRenderer) HomePage(w http.ResponseWriter, r *http.Request) {
	data := HomePageData{}

	c, err := r.Cookie(constants.TwitchIDCookieKey)
	if err == nil {
		if err = c.Valid(); err == nil {
			var u *users.User
			u, err = h.userStore.GetUser(c.Value)
			if err == nil {
				tAuthed := len(u.TwitchAccessToken) > 0 && len(u.TwitchRefreshToken) > 0
				sAuthed := len(u.SpotifyAccessToken) > 0 && len(u.SpotifyRefreshToken) > 0
				data.Authenticated = tAuthed && sAuthed
			}
		}
	}

	if err = homePage.Execute(w, data); err != nil {
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
