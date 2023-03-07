package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/saxypandabear/twitchsongrequests/cmd/songrequests"
	"go.uber.org/zap"
)

const defaultPort = 8000

func main() {
	flag.Parse()

	logger, _ := zap.NewProduction()
	zap.RedirectStdLog(logger)
	defer logger.Sync()

	var port = defaultPort
	portEnv, ok := os.LookupEnv("PORT")
	if ok {
		p, err := strconv.Atoi(portEnv)
		if err != nil {
			log.Printf("Configured PORT environment variable %s is not valid\n", portEnv)
			os.Exit(1)
		}

		port = p
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		err := songrequests.StartServer(logger, port)
		log.Println("server stopped ", err)
		wg.Done()
	}()

	wg.Wait()
	log.Println("Shutting down")
}
