package metrics

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshallMessageAsJSON(t *testing.T) {
	date := time.Date(2012, time.April, 1, 0, 0, 0, 0, time.UTC)
	m := Message{
		CreatedAt:     &date,
		Success:       1,
		BroadcasterID: "foo",
		SpotifyTrack:  "bar",
	}

	bytes, err := json.Marshal(m)
	require.NoError(t, err)
	jsonStr := string(bytes)
	assert.True(t, strings.Contains(jsonStr, "created_at"))
	assert.True(t, strings.Contains(jsonStr, "success"))
	assert.True(t, strings.Contains(jsonStr, "broadcaster_id"))
	assert.True(t, strings.Contains(jsonStr, "spotify_track"))
}
