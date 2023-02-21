package logger

import "go.uber.org/zap"

var logger *zap.Logger

func L() *zap.Logger {
	return logger
}

func init() {
	logger, _ = zap.NewProduction()
}
