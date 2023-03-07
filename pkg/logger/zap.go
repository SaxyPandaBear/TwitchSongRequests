package logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type ZapFormatter struct {
	L *zap.Logger
}

func (z *ZapFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	reqID := middleware.GetReqID(r.Context())

	return &ZapLogEntry{
		L:       z.L,
		ID:      reqID,
		Request: r,
	}
}

var _ middleware.LogFormatter = (*ZapFormatter)(nil)

type ZapLogEntry struct {
	L       *zap.Logger
	ID      string
	Request *http.Request
}

func (z *ZapLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	z.L.Info("Served",
		zap.String("requestID", z.ID),
		zap.String("method", z.Request.Method),
		zap.String("path", z.Request.RequestURI),
		zap.Int("status", status),
		zap.String("from", z.Request.RemoteAddr),
		zap.Int("size", bytes),
		zap.Duration("elapsed", elapsed))
}

func (z *ZapLogEntry) Panic(v interface{}, stack []byte) {
	middleware.PrintPrettyStack(v) // TODO: maybe implement this myself?
}

var _ middleware.LogEntry = (*ZapLogEntry)(nil)
