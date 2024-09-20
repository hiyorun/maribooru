package helpers

import (
	"log"

	"go.uber.org/zap"
)

func NewZapLogger(development bool) *zap.Logger {
	var logger *zap.Logger
	var err error
	if development {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatal("Failed to initialize zap logger")
	}

	return logger
}
