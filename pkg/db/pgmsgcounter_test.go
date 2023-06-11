package db_test

import (
	"context"
	"sync"
	"testing"

	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/o11y/metrics"
	"github.com/stretchr/testify/assert"
)

var msgOnce sync.Once

func TestPostgresAddMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integ test")
	}

	msgOnce.Do(connect)

	store := db.NewPostgresMessageCounter(pool)

	var count uint64
	err := pool.QueryRow(context.Background(), "select count(id) from messages").Scan(&count)
	assert.NoError(t, err)

	assert.Greater(t, count, uint64(0))

	store.AddMessage(&metrics.Message{
		Success:       1,
		BroadcasterID: "12345",
		SpotifyTrack:  "xyz",
	})

	var count2 uint64
	err = pool.QueryRow(context.Background(), "select count(id) from messages").Scan(&count2)
	assert.NoError(t, err)

	assert.Equal(t, count+1, count2)

	var m metrics.Message
	err = pool.QueryRow(context.Background(), "select success, broadcaster_id, spotify_track from messages where spotify_track = 'xyz'").Scan(&m.Success, &m.BroadcasterID, &m.SpotifyTrack)
	assert.NoError(t, err)
	assert.Equal(t, 1, m.Success)
	assert.Equal(t, "12345", m.BroadcasterID)
	assert.Equal(t, "xyz", m.SpotifyTrack)
}

func TestPostgresGetMessagesForUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integ test")
	}

	msgOnce.Do(connect)

	store := db.NewPostgresMessageCounter(pool)

	var count uint64
	err := pool.QueryRow(context.Background(), "select count(id) from messages where broadcaster_id = '12345'").Scan(&count)
	assert.NoError(t, err)

	assert.Greater(t, count, uint64(0))

	msgs := store.MessagesForUser("12345")
	assert.NotEmpty(t, msgs)
	assert.Equal(t, count, uint64(len(msgs)))
}

func TestPostgresTotalMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integ test")
	}

	msgOnce.Do(connect)

	store := db.NewPostgresMessageCounter(pool)

	var count uint64
	err := pool.QueryRow(context.Background(), "select count(id) from messages").Scan(&count)
	assert.NoError(t, err)

	assert.Greater(t, count, uint64(0))

	actual := store.TotalMessages()
	assert.Greater(t, count, actual)
	assert.NotZero(t, actual)
}

func TestPostgresRunningCount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integ test")
	}

	msgOnce.Do(connect)

	store := db.NewPostgresMessageCounter(pool)

	total := store.TotalMessages()

	count := store.RunningCount(5)
	assert.Greater(t, total, count)
	assert.NotZero(t, count)
}
