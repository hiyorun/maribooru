package handlers

import (
	"maribooru/internal/config"
	"maribooru/internal/helpers"
	"maribooru/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	SettingsHandler struct {
		db     *gorm.DB
		models *models.SettingsModel
		cfg    *config.Config
		log    *zap.Logger
	}
)

func NewSettingsHandler(db *gorm.DB, cfg *config.Config, log *zap.Logger) *SettingsHandler {
	return &SettingsHandler{
		db,
		models.NewSettingsModel(db),
		cfg,
		log,
	}
}

func (s *SettingsHandler) Get(c echo.Context) error {
	s.log.Debug("SettingsHandler: Get")
	data, err := s.models.Get()
	if err != nil {
		s.log.Error("Failed to get settings", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while getting settings")
	}
	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}
