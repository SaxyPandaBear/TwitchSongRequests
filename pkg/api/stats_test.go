package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/saxypandabear/twitchsongrequests/internal/testutil"
	"github.com/saxypandabear/twitchsongrequests/pkg/api"
	"github.com/saxypandabear/twitchsongrequests/pkg/o11y/metrics"
	"github.com/stretchr/testify/assert"
)

func TestCountMessages(t *testing.T) {
	counter := testutil.InMemoryMessageCounter{
		Msgs: make([]*metrics.Message, 0, 1),
	}

	sh := api.NewStatsHandler(&counter, 0, 100)

	req, err := http.NewRequest("GET", "/count", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(sh.TotalMessages)

	assert.Len(t, counter.Msgs, 0)

	ready := make(chan struct{})

	go func() {
		handler.ServeHTTP(rr, req)
		ready <- struct{}{}
	}()

	select {
	case <-ready:
		t.Log("completed request")
	case <-time.After(time.Millisecond * 100):
		t.Error("failed to complete request in time")
	}

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	bytes, err := io.ReadAll(rr.Result().Body)
	assert.NoError(t, err)
	assert.NotEmpty(t, bytes)
	var res api.SvgData
	err = json.Unmarshal(bytes, &res)
	assert.NoError(t, err)
	assert.Equal(t, "0", res.Message)
	assert.Equal(t, "Songs Queued", res.Label)
	assert.Equal(t, "for-the-badge", res.Style)
	assert.Equal(t, 1, res.SchemaVersion)
	assert.Equal(t, "informational", res.Color)

	counter.AddMessage(&metrics.Message{})
	assert.Len(t, counter.Msgs, 1)

	rr = httptest.NewRecorder()

	// Check the messages again just to be sure, add query parameter
	req, err = http.NewRequest("GET", "/count", nil)
	assert.NoError(t, err)

	go func() {
		handler.ServeHTTP(rr, req)
		ready <- struct{}{}
	}()

	select {
	case <-ready:
		t.Log("completed request")
	case <-time.After(time.Millisecond * 100):
		t.Error("failed to complete request in time")
	}

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	bytes, err = io.ReadAll(rr.Result().Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bytes, &res)
	assert.NoError(t, err)
	assert.Equal(t, "1", res.Message)
}

func TestOnboardedCount(t *testing.T) {
	sh := api.NewStatsHandler(&testutil.InMemoryMessageCounter{}, 1, 2)

	req, err := http.NewRequest("GET", "/onboarded", nil)
	assert.NoError(t, err)

	handler := http.HandlerFunc(sh.Onboarded)

	ready := make(chan struct{})

	tests := []struct {
		onboarded     int
		allowed       int
		expectedColor string
	}{
		{
			onboarded:     1,
			allowed:       10,
			expectedColor: "green",
		},
		{
			onboarded:     2,
			allowed:       4,
			expectedColor: "yellow",
		},
		{
			onboarded:     1,
			allowed:       1,
			expectedColor: "red",
		},
	}

	for _, test := range tests {
		rr := httptest.NewRecorder()

		sh.NumOnboarded = test.onboarded
		sh.NumAllowed = test.allowed

		go func() {
			handler.ServeHTTP(rr, req)
			ready <- struct{}{}
		}()

		select {
		case <-ready:
			t.Log("completed request")
		case <-time.After(time.Millisecond * 100):
			t.Error("failed to complete request in time")
		}

		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
		bytes, err := io.ReadAll(rr.Result().Body)
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		var res api.SvgData
		err = json.Unmarshal(bytes, &res)
		assert.NoError(t, err)
		assert.Equal(t, test.expectedColor, res.Color)
		assert.Equal(t, "Onboarded", res.Label)
		assert.Equal(t, "for-the-badge", res.Style)
		assert.Equal(t, fmt.Sprintf("%d/%d", test.onboarded, test.allowed), res.Message)
	}
}
