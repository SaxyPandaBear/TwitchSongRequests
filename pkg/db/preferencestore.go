package db

import "github.com/saxypandabear/twitchsongrequests/pkg/preferences"

type PreferenceStore interface {
	GetPreference(string) (*preferences.Preference, error)
	AddPreference(*preferences.Preference) error
	UpdatePreference(*preferences.Preference) error
	DeletePreference(string) error
}
