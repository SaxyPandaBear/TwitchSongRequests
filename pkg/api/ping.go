package api

import (
	"log"
	"net/http"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Pong")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Pong")); err != nil {
		log.Println("failed to write response body", err)
	}
}
