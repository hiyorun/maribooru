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
	Response struct {
		AdminCreated bool `json:"admin_created"`
	}

	Handler struct {
		db     *gorm.DB
		models *Model
		cfg    *config.Config
		log    *zap.Logger
	}
)

func (a AppSettingSlice) ToResponse() Response {
	response := Response{}
	for _, setting := range a {
		if setting.Key == "ADMIN_CREATED" {
			response.AdminCreated = setting.ValueBool
		}
	}
	return response
}

func NewHandler(db *gorm.DB, cfg *config.Config, log *zap.Logger) *Handler {
	return &Handler{
		db,
		NewModel(db),
		cfg,
		log,
	}
}

func (s *Handler) Get(c echo.Context) error {
	s.log.Debug("Handler: Get")
	data, err := s.models.Get()
	if err != nil {
		s.log.Error("Failed to get settings", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while getting settings")
	}
	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}
