package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/pkg/db"
)

type StatsHandler struct {
	msgCounter db.MessageCounter
}

// https://shields.io/endpoint schema
type SvgData struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
	Style         string `json:"style"`
}

func NewStatsHandler(counter db.MessageCounter) *StatsHandler {
	return &StatsHandler{
		msgCounter: counter,
	}
}

func (h *StatsHandler) GetMessageCount(w http.ResponseWriter, r *http.Request) {
	data := SvgData{
		SchemaVersion: 1,
		Label:         "Songs Queued",
		Style:         "for-the-badge",
		Color:         "informational",
		Message:       fmt.Sprintf("%v", h.msgCounter.CountMessages()),
	}

	bytes, err := json.Marshal(data)

	if err != nil {
		log.Println("failed to marshal data to serve in SVG", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to generate message count")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
