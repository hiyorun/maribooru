package handlers

import (
	"maribooru/internal/helpers"
	"maribooru/internal/models"
	"maribooru/internal/structs"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (u *UserHandler) InitialCreateAdmin(c echo.Context) error {
	u.log.Debug("UserHandler: InitialCreateAdmin")
	if u.cfg.AppConfig.AdminCreated {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Admin already created")
	}
	return u.CreateAdmin(c)
}

func (u *UserHandler) CreateAdmin(c echo.Context) error {
	u.log.Debug("UserHandler: CreateAdmin")

	var request structs.SignUp

	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}

	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	hashedPassword, err := helpers.PasswordHash(request.Password)
	if err != nil {
		u.log.Error("Error while hashing password", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while hashing password")
	}
	request.HashedPassword = hashedPassword

	data, err := u.model.Create(request.ToTable())
	if err != nil {
		u.log.Error("Failed to create user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while creating admin")
	}

	admin, err := u.model.AssignAdmin(data.ID)
	if err != nil {
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to assign admin")
	}

	if !u.cfg.AppConfig.AdminCreated {
		settingsModel := models.NewSettingsModel(u.db)
		adminSettings := structs.AppSettings{
			Key:       "ADMIN_CREATED",
			ValueBool: true,
		}

		if err := settingsModel.Update(adminSettings); err != nil {
			return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to update admin settings")
		}
	}

	token, err := helpers.GenerateJWT(data.ID, data.Name, u.cfg.JWT.Secret, u.cfg.AppConfig.TokenLifetime)
	if err != nil {
		u.log.Error("Error while generating token", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to generate token")
	}

	return helpers.Response(c, http.StatusOK, admin, token)
}

func (u *UserHandler) GetAllAdmin(c echo.Context) error {
	u.log.Debug("UserHandler: GetAllAdmin")

	bounds := structs.PagedRequest{
		Limit:    50,
		Offset:   0,
		Keywords: "",
		Sort:     "",
	}
	if err := c.Bind(&bounds); err != nil {
		u.log.Error("Failed to set limit and offset, defaulting to 50 limit and 0 offset", zap.Error(err))
	}

	data, err := u.model.GetAllAdmin(bounds)
	if err != nil {
		u.log.Error("Failed to get admins", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to get admins")
	}

	return helpers.Response(c, http.StatusOK, data.ToResponse(true), "")
}

func (u *UserHandler) AssignAdmin(c echo.Context) error {
	u.log.Debug("UserHandler: AssignAdmin")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}

	data, err := u.model.AssignAdmin(id)
	if err != nil {
		u.log.Error("Failed to assign admin", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to assign admin")
	}

	return helpers.Response(c, http.StatusOK, data, "")
}

func (u *UserHandler) RemoveAdmin(c echo.Context) error {
	u.log.Debug("UserHandler: RemoveAdmin")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}

	data, err := u.model.RemoveAdmin(id)
	if err != nil {
		u.log.Error("Failed to remove admin", zap.Error(err))
	}

	return helpers.Response(c, http.StatusOK, data, "")
}

func (u *UserHandler) AdministrativeUserUpdate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}

	var request structs.UserUpdate
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}

	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	return u.update(c, request.ToTable(id))
}
