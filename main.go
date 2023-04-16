package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/saxypandabear/twitchsongrequests/cmd/songrequests"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultPort = 8000

func main() {
	flag.Parse()

	config := zap.NewProductionConfig()
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // human readable timestamps for logs
	config.EncoderConfig = encoderConfig
	logger := zap.Must(config.Build())

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

	go func() {
		if err := songrequests.StartServer(logger, port); err != nil {
			log.Println("server terminated unexpectedly", err)
			os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
