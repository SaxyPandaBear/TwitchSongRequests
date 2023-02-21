package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/saxypandabear/twitchsongrequests/cmd/songrequests"
	"github.com/saxypandabear/twitchsongrequests/internal/logger"
	"go.uber.org/zap"
)

const defaultPort = 8000

func main() {
	flag.Parse()

	zap.RedirectStdLog(logger.L())
	defer logger.L().Sync()

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
		err := songrequests.StartServer(port)
		log.Println("server stopped ", err)
		wg.Done()
	}()

	wg.Wait()
	log.Println("Shutting down")
}
