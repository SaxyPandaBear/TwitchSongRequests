package db_test

import (
	"sync"
	"testing"

	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
	"github.com/stretchr/testify/assert"
)

var userOnce sync.Once

func TestPostgresGetUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integ test")
	}

	userOnce.Do(connect)

	store := db.NewPostgresUserStore(pool)
	u, err := store.GetUser("12345")
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "12345", u.TwitchID)
	assert.Equal(t, "a", u.TwitchAccessToken)
	assert.Equal(t, "b", u.TwitchRefreshToken)
	assert.Equal(t, "c", u.SpotifyAccessToken)
	assert.Equal(t, "d", u.SpotifyRefreshToken)
	assert.False(t, u.Subscribed)
	assert.Equal(t, "abc-123", u.SubscriptionID)
	assert.Equal(t, "foo@bar", u.Email)
}

func TestPostgresGetUserMissing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integ test")
	}

	userOnce.Do(connect)

	store := db.NewPostgresUserStore(pool)
	u, err := store.GetUser("98765")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestPostgresAddUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integ test")
	}

	userOnce.Do(connect)

	store := db.NewPostgresUserStore(pool)

	u := users.User{
		TwitchID: "555",
	}

	err := store.AddUser(&u)
	assert.NoError(t, err)

	fetched, err := store.GetUser("555")
	assert.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, "555", fetched.TwitchID)
}

func TestPostgresUpdateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integ test")
	}

	userOnce.Do(connect)

	store := db.NewPostgresUserStore(pool)

	u := &users.User{
		TwitchID: "22",
		Email:    "abc@123",
	}

	err := store.AddUser(u)
	assert.NoError(t, err)

	u, err = store.GetUser("22")
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "22", u.TwitchID)
	assert.Equal(t, "abc@123", u.TwitchID)

	u.Email = "123@abc"
	err = store.UpdateUser(u)
	assert.NoError(t, err)

	u, err = store.GetUser("22")
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "123@abc", u.Email)
}

func TestPostgresDeleteUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integ test")
	}

	userOnce.Do(connect)

	store := db.NewPostgresUserStore(pool)

	u := &users.User{
		TwitchID: "555",
	}

	err := store.AddUser(u)
	assert.NoError(t, err)

	err = store.DeleteUser("555")
	assert.NoError(t, err)

	u, err = store.GetUser("555")
	assert.Error(t, err)
	assert.Nil(t, u)
}
