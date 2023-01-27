package main

import (
	"flag"
	"log"
	"sync"

	"github.com/saxypandabear/twitchsongrequests/cmd/songrequests"
)

var port = flag.Int("port", 8000, "port that the HTTP server listens on")

func main() {
	flag.Parse()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		err := songrequests.StartServer(*port)
		log.Println("server stopped ", err)
		wg.Done()
	}()

	wg.Wait()
	log.Println("Shutting down")
}
