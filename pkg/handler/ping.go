package handler

import (
	"log"
	"net/http"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Pong")
	w.WriteHeader(http.StatusOK)
}
