package account

import (
	"errors"
	"maribooru/internal/config"
	"maribooru/internal/helpers"
	"maribooru/internal/models"
	"maribooru/internal/structs"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	AdminResponse struct {
		AdminID   uuid.UUID `json:"admin_id"`
		UserID    uuid.UUID `json:"user_id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	AdminHandler struct {
		db        *gorm.DB
		model     *AdminModel
		userModel *UserModel
		cfg       *config.Config
		log       *zap.Logger
	}
)

func (a *Admin) ToResponse() AdminResponse {
	return AdminResponse{
		AdminID:   a.ID,
		UserID:    a.UserID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func (a *AdminHandler) InitialCreateAdmin(c echo.Context) error {
	a.log.Debug("UserHandler: InitialCreateAdmin")
	if a.cfg.AppConfig.AdminCreated {
		return helpers.Response(c, http.StatusForbidden, nil, "Admin already created")
	}

	params := UserParams{
		GenericPagedQuery: helpers.GenericPagedQuery{
			Limit:    10,
			Offset:   0,
			Sort:     "id",
			Keywords: "",
		},
		IsAdmin: true,
	}

	_, count, err := a.userModel.GetAll(params)
	if err != nil {
		a.log.Error("Failed to check if admin exists", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "")
	}

	if count > 0 {
		a.log.Warn("Admin exists, but ADMIN_CREATED is false. Updating ADMIN_CREATED to true")
		a.cfg.AppConfig.AdminCreated = true

		model := models.NewSettingsModel(a.db)
		settings := structs.AppSettings{
			Key:       "ADMIN_CREATED",
			ValueBool: true,
		}
		err := model.Update(settings)
		if err != nil {
			a.log.Error("Failed to update settings", zap.Error(err))
			return helpers.Response(c, http.StatusInternalServerError, nil, "")
		}

		return helpers.Response(c, http.StatusForbidden, nil, "Admin already created")
	}

	return a.CreateAdmin(c)
}

func (a *AdminHandler) CreateAdmin(c echo.Context) error {
	a.log.Debug("UserHandler: CreateAdmin")

	var request SignUp

	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}

	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	hashedPassword, err := helpers.PasswordHash(request.Password)
	if err != nil {
		a.log.Error("Error while hashing password", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while hashing password")
	}
	request.HashedPassword = hashedPassword

	tx := a.db.Begin()
	userModel := NewUserModel(tx)
	adminModel := NewAdminModel(tx)

	data, err := userModel.Create(request.ToTable())
	if err != nil {
		tx.Rollback()
		a.log.Error("Failed to create user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while creating admin")
	}

	admin, err := adminModel.AssignAdmin(data.ID)
	if err != nil {
		tx.Rollback()
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to assign admin")
	}

	if !a.cfg.AppConfig.AdminCreated {
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

	token, err := helpers.GenerateJWT(data.ID, data.Name, a.cfg.JWT.Secret, a.cfg.AppConfig.TokenLifetime)
	if err != nil {
		tx.Rollback()
		a.log.Error("Error while generating token", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to generate token")
	}

	data.Admin = admin

	tx.Commit()
	return helpers.Response(c, http.StatusOK, data.ToResponse(true), token)
}

func (a *AdminHandler) GetAllAdmin(c echo.Context) error {
	a.log.Debug("UserHandler: GetAllAdmin")

	params := UserParams{
		GenericPagedQuery: helpers.GenericPagedQuery{
			Limit:    50,
			Offset:   0,
			Keywords: "",
			Sort:     "",
		},
		IsAdmin: true,
	}
	if err := c.Bind(&params); err != nil {
		a.log.Error("Failed to set limit and offset, defaulting to 50 limit and 0 offset", zap.Error(err))
	}

	data, total, err := a.userModel.GetAll(params)
	if err != nil {
		a.log.Error("Failed to get admins", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to get admins")
	}

	paged := helpers.PageData(data.ToResponse(true), int(total), params.Offset, params.Limit)

	return helpers.Response(c, http.StatusOK, paged, "")
}

func (a *AdminHandler) AssignAdmin(c echo.Context) error {
	a.log.Debug("UserHandler: AssignAdmin")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}

	data, err := a.model.AssignAdmin(id)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return helpers.Response(c, http.StatusConflict, nil, "Already an admin")
		}
		a.log.Error("Failed to assign admin", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to assign admin")
	}

	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}

func (a *AdminHandler) RemoveAdmin(c echo.Context) error {
	a.log.Debug("UserHandler: RemoveAdmin")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}

	data, err := a.model.RemoveAdmin(id)
	if err != nil {
		a.log.Error("Failed to remove admin", zap.Error(err))
	}

	return helpers.Response(c, http.StatusOK, data, "")
}

func (a *AdminHandler) AdministrativeUserUpdate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}

	var request UserUpdate
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}

	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	user, err := a.userModel.Update(request.ToTable(id))
	if err != nil {
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to update user")
	}

	return helpers.Response(c, http.StatusOK, user, "")
}
