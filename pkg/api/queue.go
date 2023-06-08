package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/o11y/metrics"
	"go.uber.org/zap"
)

type QueueHandler struct {
	msgCounter db.MessageCounter
}

type queuedTracks struct {
	ID     string
	Tracks []*metrics.Message
}

func NewQueueHandler(c db.MessageCounter) *QueueHandler {
	return &QueueHandler{
		msgCounter: c,
	}
}

func (h *QueueHandler) GetLatestMessages(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		zap.L().Warn("Unable to associate user ID with request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	msgs := h.msgCounter.MessagesForUser(id)

	payload := queuedTracks{
		ID:     id,
		Tracks: msgs,
	}

	// TODO: redirect once this is tested out
	bytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error("failed to marshall payload", zap.String("id", id), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(bytes); err != nil {
		zap.L().Error("failed to write response", zap.Error(err))
	}
}
