package db

import "github.com/saxypandabear/twitchsongrequests/pkg/preferences"

type PreferenceStore interface {
	GetPreference(string) (*preferences.Preference, error)
	AddPreference(*preferences.Preference) error
	UpdatePreference(*preferences.Preference) error
	DeletePreference(string) error
}

type NoopPreferenceStore struct{}

// AddPreference implements PreferenceStore.
func (n *NoopPreferenceStore) AddPreference(*preferences.Preference) error {
	return nil
}

// DeletePreference implements PreferenceStore.
func (n *NoopPreferenceStore) DeletePreference(string) error {
	return nil
}

// GetPreference implements PreferenceStore.
func (n *NoopPreferenceStore) GetPreference(string) (*preferences.Preference, error) {
	return nil, nil
}

// UpdatePreference implements PreferenceStore.
func (n *NoopPreferenceStore) UpdatePreference(*preferences.Preference) error {
	return nil
}

var _ PreferenceStore = (*NoopPreferenceStore)(nil)
