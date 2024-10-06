package permission

import (
	"errors"
	"maribooru/internal/config"
	"maribooru/internal/helpers"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	Level   int
	Request struct {
		UserID     uuid.UUID `json:"user_id" validate:"required"`
		Permission Level     `json:"permission_level" validate:"required"`
	}

	Response struct {
		UserID     uuid.UUID `json:"user_id"`
		Permission Level     `json:"permission_level"`
		UpdatedAt  time.Time `json:"updated_at"`
	}

	Handler struct {
		db    *gorm.DB
		model *Model
		cfg   *config.Config
		log   *zap.Logger
	}
)

const (
	Read Level = 1 << iota
	Write
	Approve
	Moderate
)

func (p *Request) ToTable() Permission {
	return Permission{
		Permission: p.Permission,
		UserID:     p.UserID,
	}
}

func (p *Permission) ToResponse() Response {
	return Response{
		UserID:     p.UserID,
		Permission: p.Permission,
		UpdatedAt:  p.UpdatedAt,
	}
}

func NewHandler(db *gorm.DB, cfg *config.Config, log *zap.Logger) *Handler {
	return &Handler{
		db,
		NewModel(db),
		cfg,
		log,
	}
}

func (p *Handler) GetByUserID(c echo.Context) error {
	p.log.Debug("PermissionHandler: GetByUserID")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}
	data, err := p.model.GetByUserID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "User not found")
		}
		p.log.Debug("Failed to get user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while getting user")
	}
	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}

func (p *Handler) Set(c echo.Context) error {
	p.log.Debug("PermissionHandler: SetPermission")

	var request Request
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}
	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}
	data, err := p.model.SetPermission(request.ToTable())
	if err != nil {
		p.log.Debug("Failed to set permission", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while setting permission")
	}
	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}
