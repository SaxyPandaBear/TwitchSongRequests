package logger_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
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

	attributes := getZapAttributes(t, log)
	assert.Len(t, attributes, 7)

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
			assert.Equal(t, v, attributes[k])
		})
	}
}

func TestZapLogEntryPanic(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	l := zap.New(core)

	requestID := "foo"

	req := httptest.NewRequest("GET", "/ping", strings.NewReader("hello, world"))

	entry := logger.ZapLogEntry{
		L:       l,
		ID:      requestID,
		Request: req,
	}

	entry.Panic(struct{}{}, []byte("abc123"))

	assert.Equal(t, 1, observed.Len())
	log := observed.All()[0]

	assert.Equal(t, zapcore.ErrorLevel, log.Level)
	assert.Equal(t, "Panicked", log.Message)

	attributes := getZapAttributes(t, log)

	assert.Len(t, attributes, 1)

	expected := map[string]interface{}{
		"stack": "abc123",
	}

	for k, v := range expected {
		t.Run(fmt.Sprintf("Check %s", k), func(t *testing.T) {
			assert.Equal(t, v, attributes[k])
		})
	}
}

func TestNewLogEntry(t *testing.T) {
	ctx := context.WithValue(context.TODO(), middleware.RequestIDKey, "foo")

	core, _ := observer.New(zapcore.DebugLevel)
	l := zap.New(core)

	req := httptest.NewRequest("GET", "/ping", strings.NewReader("hello, world")).WithContext(ctx)

	f := logger.ZapFormatter{
		L: l,
	}

	e := f.NewLogEntry(req)
	entry, ok := e.(*logger.ZapLogEntry)
	assert.True(t, ok)
	assert.Equal(t, "foo", entry.ID)
}

func getZapAttributes(t *testing.T, log observer.LoggedEntry) map[string]interface{} {
	t.Helper()
	m := make(map[string]interface{})
	//nolint: exhaustive
	for _, field := range log.Context {
		switch field.Type {
		case zapcore.StringType:
			m[field.Key] = field.String
		case zapcore.Int64Type:
			m[field.Key] = field.Integer
		}
	}

	return m
}
