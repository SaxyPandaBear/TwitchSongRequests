package site

import (
	"html/template"
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
)

var preferencesPage = template.Must(template.ParseFiles("pkg/site/preferences.html"))

type PreferencesRenderer struct {
	pref    db.PreferenceStore
	siteURL string
}

type PreferencePageData struct {
	Authenticated bool
	RewardID      string
	Explicit      bool
}

func NewPreferencesRenderer(siteURL string, p db.PreferenceStore) *PreferencesRenderer {
	return &PreferencesRenderer{
		pref:    p,
		siteURL: siteURL,
	}
}

func (p *PreferencesRenderer) PreferencesPage(w http.ResponseWriter, r *http.Request) {
	d := PreferencePageData{
		Authenticated: true,
	}

	id, err := util.GetUserIDFromRequest(r)
	if err != nil {
		log.Println("failed to get Twitch ID from request", err)
		d.Authenticated = false
	}

	pref, err := p.pref.GetPreference(id)
	if err != nil {
		log.Println("failed to get user preferences", err)
	} else {
		d.Explicit = pref.ExplicitSongs
		d.RewardID = pref.CustomRewardID
	}

	if err := preferencesPage.Execute(w, &d); err != nil {
		log.Println("error occurred while executing template:", err)
	}
}
