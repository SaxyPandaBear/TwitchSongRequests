package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"go.uber.org/zap"
)

const (
	PrefFormExplicitKey   = "explicit"
	PrefFormSongLengthKey = "song-length"
)

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
		zap.L().Error("failed to get Twitch ID from request", zap.Error(err))
		w.Write([]byte(err.Error()))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	p, err := h.prefs.GetPreference(userID)
	if err != nil {
		zap.L().Error("failed to get user preferences", zap.String("id", userID), zap.Error(err))
		w.Write([]byte(err.Error()))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	err = r.ParseForm()
	if err != nil {
		zap.L().Error("failed to parse HTML form", zap.Error(err))
		w.Write([]byte(err.Error()))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	// leaving the checkbox unchecked omits it from the form,
	// so need to always compare the value from the checkbox
	p.ExplicitSongs = r.Form.Get(PrefFormExplicitKey) == "true"

	// if the song length value exists, update the preference with it
	if length := r.Form.Get(PrefFormSongLengthKey); length != "" {
		var l int
		l, err = strconv.Atoi(length)
		if err == nil && l >= 0 {
			p.MaxSongLength = l * 1000 // the value is expected to be in seconds, but we store millis
		} else {
			zap.L().Error("failed to convert", zap.String("input", length), zap.Error(err))
		}
	}

	err = h.prefs.UpdatePreference(p)
	if err != nil {
		zap.L().Error("failed to update user preferences", zap.String("id", userID), zap.Error(err))
		w.Write([]byte(err.Error()))
		http.Redirect(w, r, h.redirectURL, http.StatusFound)
		return
	}

	log.Println("successfully saved user preferences for", userID)
	// redirect this back to the home page.
	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
