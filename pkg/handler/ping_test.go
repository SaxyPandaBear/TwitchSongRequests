package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saxypandabear/twitchsongrequests/pkg/handler"
	"github.com/stretchr/testify/assert"
)

func TestPingHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/health", nil)
	assert.NoError(t, err)

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handler.PingHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Pong", rr.Body.String())
}
