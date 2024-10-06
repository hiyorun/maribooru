package handlers

import (
	"errors"
	"maribooru/internal/config"
	"maribooru/internal/helpers"
	"maribooru/internal/models"
	"maribooru/internal/structs"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PermissionHandler struct {
	db    *gorm.DB
	model *models.PermissionModel
	cfg   *config.Config
	log   *zap.Logger
}

func NewPermissionHandler(db *gorm.DB, cfg *config.Config, log *zap.Logger) *PermissionHandler {
	return &PermissionHandler{
		db,
		models.NewPermissionModel(db),
		cfg,
		log,
	}
}

func (p *PermissionHandler) GetByUserID(c echo.Context) error {
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

func (p *PermissionHandler) Set(c echo.Context) error {
	p.log.Debug("PermissionHandler: SetPermission")

	var request structs.PermissionRequest
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
