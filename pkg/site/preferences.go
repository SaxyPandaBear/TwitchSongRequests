package site

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
)

var preferencesPage = template.Must(template.ParseFiles("pkg/site/preferences.html"))

type PreferencesRenderer struct {
	siteURL string
	pref    db.PreferenceStore
}

type PreferencePageData struct {
	SaveURL         string
	Authenticated   bool
	RewardID        string
	Explicit        bool
	SongLengthLimit int `unit:"seconds"`
}

func NewPreferencesRenderer(p db.PreferenceStore, siteURL string) *PreferencesRenderer {
	return &PreferencesRenderer{
		pref:    p,
		siteURL: siteURL,
	}
}

func (p *PreferencesRenderer) PreferencesPage(w http.ResponseWriter, r *http.Request) {
	d := PreferencePageData{
		SaveURL:       fmt.Sprintf("%s/preference", p.siteURL),
		Authenticated: true,
	}

	id, err := util.GetUserIDFromRequest(r)
	if err != nil {
		log.Println("failed to get Twitch ID from request", err)
		d.Authenticated = false
	}

	if id != "" {
		pref, err := p.pref.GetPreference(id)
		if err != nil {
			log.Println("failed to get user preferences for", id, err)
		} else {
			d.Explicit = pref.ExplicitSongs
			d.RewardID = pref.CustomRewardID
			d.SongLengthLimit = pref.MaxSongLength / 1000 // stored as millis
		}
	}

	log.Println("Serving preferences page to", id)

	if err := preferencesPage.Execute(w, &d); err != nil {
		log.Println("error occurred while executing template:", err)
	}
}
