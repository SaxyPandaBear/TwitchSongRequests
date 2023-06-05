package logger_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/saxypandabear/twitchsongrequests/pkg/o11y/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestWriteZapEntry(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	l := zap.New(core)

	requestID := "foo"

	req := httptest.NewRequest("GET", "/ping", strings.NewReader("hello, world"))

	entry := logger.ZapLogEntry{
		L:       l,
		ID:      requestID,
		Request: req,
	}

	entry.Write(400, 0, http.Header{}, 10*time.Nanosecond, nil)

	assert.Equal(t, 1, observed.Len())
	log := observed.All()[0]

	assert.Equal(t, zapcore.InfoLevel, log.Level)
	assert.Equal(t, "Served", log.Message)

	m := make(map[string]interface{})
	for _, field := range log.Context {
		switch field.Type {
		case zapcore.StringType:
			m[field.Key] = field.String
		case zapcore.Int64Type:
			m[field.Key] = field.Integer
		}
	}
	assert.Len(t, m, 7)

	expected := map[string]interface{}{
		"requestID": "foo",
		"method":    "GET",
		"path":      "/ping",
		"status":    int64(400),
		"from":      req.RemoteAddr,
		"size":      int64(0),
		"elapsedNs": int64(10),
	}

	for k, v := range expected {
		t.Run(fmt.Sprintf("Check %s", k), func(t *testing.T) {
			assert.Equal(t, v, m[k])
		})
	}
}
