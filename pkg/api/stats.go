package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"go.uber.org/zap"
)

type StatsHandler struct {
	msgCounter   db.MessageCounter
	NumOnboarded int
	NumAllowed   int
}

// https://shields.io/endpoint schema
type SvgData struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
	Style         string `json:"style"`
	CacheSeconds  int    `json:"cacheSeconds"`
}

func NewStatsHandler(counter db.MessageCounter, onboarded, allowed int) *StatsHandler {
	return &StatsHandler{
		msgCounter:   counter,
		NumOnboarded: onboarded,
		NumAllowed:   allowed,
	}
}

func (h *StatsHandler) TotalMessages(w http.ResponseWriter, r *http.Request) {
	data := SvgData{
		SchemaVersion: 1,
		Label:         "Songs Queued",
		Style:         "for-the-badge",
		Color:         "informational",
		Message:       fmt.Sprintf("%v", h.msgCounter.TotalMessages()),
		CacheSeconds:  60 * 30,
	}

	respondWithSVG(w, &data)
}

func (h *StatsHandler) RunningCount(w http.ResponseWriter, r *http.Request) {
	data := SvgData{
		SchemaVersion: 1,
		Label:         "Queued in the last ? days",
		Style:         "for-the-badge",
		Color:         "informational",
		Message:       "0", // default just in case
		CacheSeconds:  60 * 60,
	}

	daysBack := r.URL.Query().Get("days")

	if i, err := strconv.Atoi(daysBack); err == nil && i > 0 {
		data.Label = fmt.Sprintf("Queued in the last %d days", i)
		data.Message = fmt.Sprintf("%v", h.msgCounter.RunningCount(i))
	}

	respondWithSVG(w, &data)
}

func (h *StatsHandler) Onboarded(w http.ResponseWriter, r *http.Request) {
	var color string
	pct := float32(h.NumOnboarded) / float32(h.NumAllowed)

	if pct < 0.4 {
		color = "green"
	} else if pct < 0.75 {
		color = "yellow"
	} else {
		color = "red"
	}

	data := SvgData{
		SchemaVersion: 1,
		Label:         "Onboarded",
		Style:         "for-the-badge",
		Color:         color,
		Message:       fmt.Sprintf("%d/%d", h.NumOnboarded, h.NumAllowed),
		CacheSeconds:  60 * 60,
	}

	respondWithSVG(w, &data)
}

func respondWithSVG(w http.ResponseWriter, d *SvgData) {
	bytes, err := json.Marshal(d)

	if err != nil {
		zap.L().Error("failed to marshal data to serve in SVG", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to generate message count")
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(bytes); err != nil {
		zap.L().Error("failed to write response", zap.Error(err))
	}
}
