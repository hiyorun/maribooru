package handlers

import (
	"errors"
	"maribooru/internal/helpers"
	"maribooru/internal/models"
	"maribooru/internal/structs"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (u *UserHandler) InitialCreateAdmin(c echo.Context) error {
	u.log.Debug("UserHandler: InitialCreateAdmin")
	if u.cfg.AppConfig.AdminCreated {
		return helpers.Response(c, http.StatusForbidden, nil, "Admin already created")
	}

	bounds := structs.PagedRequest{
		Limit:    10,
		Offset:   0,
		Sort:     "id",
		Keywords: "",
	}

	_, count, err := u.model.GetAllAdmin(bounds)
	if err != nil {
		u.log.Error("Failed to check if admin exists", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "")
	}

	if count > 0 {
		u.log.Warn("Admin exists, but ADMIN_CREATED is false. Updating ADMIN_CREATED to true")
		u.cfg.AppConfig.AdminCreated = true

		model := models.NewSettingsModel(u.db)
		settings := structs.AppSettings{
			Key:       "ADMIN_CREATED",
			ValueBool: true,
		}
		err := model.Update(settings)
		if err != nil {
			u.log.Error("Failed to update settings", zap.Error(err))
			return helpers.Response(c, http.StatusInternalServerError, nil, "")
		}

		return helpers.Response(c, http.StatusForbidden, nil, "Admin already created")
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

	tx := u.db.Begin()
	userModel := models.NewUserModel(tx)

	data, err := userModel.Create(request.ToTable())
	if err != nil {
		tx.Rollback()
		u.log.Error("Failed to create user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while creating admin")
	}

	admin, err := userModel.AssignAdmin(data.ID)
	if err != nil {
		tx.Rollback()
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to assign admin")
	}

	if !u.cfg.AppConfig.AdminCreated {
		settingsModel := models.NewSettingsModel(tx)
		adminSettings := structs.AppSettings{
			Key:       "ADMIN_CREATED",
			ValueBool: true,
		}

		if err := settingsModel.Update(adminSettings); err != nil {
			tx.Rollback()
			return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to update admin settings")
		}
	}

	token, err := helpers.GenerateJWT(data.ID, data.Name, u.cfg.JWT.Secret, u.cfg.AppConfig.TokenLifetime)
	if err != nil {
		tx.Rollback()
		u.log.Error("Error while generating token", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to generate token")
	}

	data.Admin = admin

	tx.Commit()
	return helpers.Response(c, http.StatusOK, data.ToResponse(true), token)
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

	data, total, err := u.model.GetAllAdmin(bounds)
	if err != nil {
		u.log.Error("Failed to get admins", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to get admins")
	}

	paged := helpers.PageData(data.ToResponse(true), int(total), bounds.Offset, bounds.Limit)

	return helpers.Response(c, http.StatusOK, paged, "")
}

func (u *UserHandler) AssignAdmin(c echo.Context) error {
	u.log.Debug("UserHandler: AssignAdmin")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}

	data, err := u.model.AssignAdmin(id)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return helpers.Response(c, http.StatusConflict, nil, "Already an admin")
		}
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
