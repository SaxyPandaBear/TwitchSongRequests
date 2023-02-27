package users

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUserAuthenticated(t *testing.T) {
	testCases := map[User]bool{
		{
			TwitchID:            "123",
			TwitchAccessToken:   "foo",
			TwitchRefreshToken:  "foo",
			SpotifyAccessToken:  "bar",
			SpotifyRefreshToken: "bar",
		}: true,
		{ // missing spotify auth
			TwitchID:           "234",
			TwitchAccessToken:  "foo",
			TwitchRefreshToken: "foo",
		}: false,
	}

	for u, shouldBeAuthenticated := range testCases {
		t.Run(u.TwitchID, func(t *testing.T) {
			assert.Equal(t, shouldBeAuthenticated, u.IsAuthenticated())
		})
	}
}
