package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

func (h *StatsHandler) TotalMessages(w http.ResponseWriter, r *http.Request) {
	data := SvgData{
		SchemaVersion: 1,
		Label:         "Songs Queued",
		Style:         "for-the-badge",
		Color:         "informational",
		Message:       fmt.Sprintf("%v", h.msgCounter.TotalMessages()),
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

func (h *StatsHandler) RunningCount(w http.ResponseWriter, r *http.Request) {
	data := SvgData{
		SchemaVersion: 1,
		Label:         "Queued in the last ? days",
		Style:         "for-the-badge",
		Color:         "informational",
		Message:       "0", // default just in case
	}

	daysBack := r.URL.Query().Get("days")

	if i, err := strconv.Atoi(daysBack); err == nil {
		data.Label = fmt.Sprintf("Queued in the last %d days", i)
		data.Message = fmt.Sprintf("%v", h.msgCounter.RunningCount(i))
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
