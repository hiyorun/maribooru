package routes

import (
	"maribooru/api/middlewares"
	"maribooru/internal/config"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type VersionOne struct {
	e   *echo.Echo
	db  *gorm.DB
	cfg *config.Config
	api *echo.Group
	mw  *middlewares.Middleware
	log *zap.Logger
}

func InitVersionOne(e *echo.Echo, db *gorm.DB, cfg *config.Config, log *zap.Logger) *VersionOne {
	return &VersionOne{
		e,
		db,
		cfg,
		e.Group("/api/v1"),
		middlewares.NewMiddleware(cfg, db, log),
		log,
	}
}
