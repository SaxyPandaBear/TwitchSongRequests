package site

import (
	"html/template"
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/pkg/db"
)

var preferencesPage = template.Must(template.ParseFiles("pkg/site/home.html"))

type PreferencesRenderer struct {
	userStore db.UserStore
	siteURL   string
}

func NewPreferencesRenderer(siteURL string, u db.UserStore) *PreferencesRenderer {
	return &PreferencesRenderer{
		userStore: u,
		siteURL:   siteURL,
	}
}

func (p *PreferencesRenderer) PreferencesPage(w http.ResponseWriter, r *http.Request) {
	// id, err := util.GetUserIDFromRequest(r)
	// if err != nil {
	// 	log.Println("failed to get Twitch ID from request", err)
	// 	http.Redirect(w, r, p.siteURL, http.StatusFound)
	// 	return
	// }

	if err := preferencesPage.Execute(w, nil); err != nil {
		log.Println("error occurred while executing template:", err)
	}
}
