package db_test

import (
	"sync"
	"testing"

	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/stretchr/testify/assert"
)

var prefOnce sync.Once

func TestPostgresGetPreference(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integ test")
	}

	prefOnce.Do(connect)

	store := db.NewPostgresPreferenceStore(pool)

	p, err := store.GetPreference("12345")
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "abc-123", p.CustomRewardID)
	assert.False(t, p.ExplicitSongs)
	assert.Equal(t, "12345", p.TwitchID)
	assert.Zero(t, p.MaxSongLength)
}

func TestPostgresGetPreferenceMissing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integ test")
	}

	prefOnce.Do(connect)

	store := db.NewPostgresPreferenceStore(pool)

	p, err := store.GetPreference("545678")
	assert.Error(t, err)
	assert.Nil(t, p)
}
