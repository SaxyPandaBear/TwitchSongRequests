package api

import (
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
)

const PrefFormExplicitKey = "explicit"

type PreferenceHandler struct {
	prefs       db.PreferenceStore
	redirectURL string
}

func NewPreferenceHandler(d db.PreferenceStore, redirectURL string) *PreferenceHandler {
	return &PreferenceHandler{
		prefs:       d,
		redirectURL: redirectURL,
	}
}

func (h *PreferenceHandler) SavePreferences(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetUserIDFromRequest(r)

	if err != nil {
		log.Println("failed to get Twitch ID from request", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	p, err := h.prefs.GetPreference(userID)
	if err != nil {
		log.Println("failed to get user preferences for", userID, err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Println("failed to parse HTML form", err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	// leaving the checkbox unchecked omits it from the form,
	// so need to always compare the value from the checkbox
	p.ExplicitSongs = r.Form.Get(PrefFormExplicitKey) == "true"

	err = h.prefs.UpdatePreference(p)
	if err != nil {
		log.Println("failed to update user preferences for", userID, err)
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	log.Println("successfully saved user preferences for", userID)
	// redirect this back to the home page.
	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
