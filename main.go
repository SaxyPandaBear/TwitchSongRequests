package main

import (
	"flag"
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
	config.Level.SetLevel(zapcore.DebugLevel)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // human readable timestamps for logs
	config.EncoderConfig = encoderConfig
	logger := zap.Must(config.Build())

	zap.RedirectStdLog(logger) // log.PrintX() functions will log at an info level, always
	zap.ReplaceGlobals(logger) // Not recommended, but I'm lazy
	defer logger.Sync()

	var port = defaultPort
	portEnv, ok := os.LookupEnv("PORT")
	if ok {
		p, err := strconv.Atoi(portEnv)
		if err != nil {
			zap.L().Error("Configured PORT environment variable is not valid", zap.String("port", portEnv))
			os.Exit(1)
		}

		port = p
	}

	go func() {
		if err := songrequests.StartServer(logger, port); err != nil {
			zap.L().Error("server terminated unexpectedly", zap.Error(err))
			os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
