package middlewares

import (
	"maribooru/internal/config"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Middleware struct {
	cfg *config.Config
	db  *gorm.DB
	log *zap.Logger
}

func NewMiddleware(cfg *config.Config, db *gorm.DB, log *zap.Logger) *Middleware {
	return &Middleware{
		cfg: cfg,
		db:  db,
		log: log,
	}
}
