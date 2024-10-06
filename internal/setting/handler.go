package setting

import (
	"maribooru/internal/config"
	"maribooru/internal/helpers"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	AppSettingResponse struct {
		AdminCreated bool `json:"admin_created"`
	}

	SettingHandler struct {
		db     *gorm.DB
		models *SettingModel
		cfg    *config.Config
		log    *zap.Logger
	}
)

func (a AppSettingSlice) ToResponse() AppSettingResponse {
	response := AppSettingResponse{}
	for _, setting := range a {
		if setting.Key == "ADMIN_CREATED" {
			response.AdminCreated = setting.ValueBool
		}
	}
	return response
}

func NewSettingHandler(db *gorm.DB, cfg *config.Config, log *zap.Logger) *SettingHandler {
	return &SettingHandler{
		db,
		NewSettingModel(db),
		cfg,
		log,
	}
}

func (s *SettingHandler) Get(c echo.Context) error {
	s.log.Debug("SettingHandler: Get")
	data, err := s.models.Get()
	if err != nil {
		s.log.Error("Failed to get settings", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while getting settings")
	}
	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}
