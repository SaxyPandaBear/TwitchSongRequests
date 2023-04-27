package api_test

import (
	"encoding/json"
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

	sh := api.NewStatsHandler(&counter)

	req, err := http.NewRequest("GET", "/count", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(sh.GetMessageCount)

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

	counter.AddMessage(&metrics.Message{})
	assert.Len(t, counter.Msgs, 1)

	rr = httptest.NewRecorder()

	// Check the messages again just to be sure
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
	var res2 api.SvgData
	err = json.Unmarshal(bytes, &res2)
	assert.NoError(t, err)
	assert.Equal(t, "1", res2.Message)
}
