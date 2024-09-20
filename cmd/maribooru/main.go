package main

import (
	"log"
	"maribooru/api"
	"maribooru/internal/config"
	"maribooru/internal/db"
	"maribooru/internal/helpers"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	log := helpers.NewZapLogger(cfg.AppConfig.Development)
	log.Info("Starting API Server")

	db, err := db.InitDatabase(cfg, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	e := api.NewHTTPServer(cfg, db, log)
	e.RunHTTPServer()
}
