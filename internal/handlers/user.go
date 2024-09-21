package handlers

import (
	"maribooru/internal/config"
	"maribooru/internal/helpers"
	"maribooru/internal/models"
	"maribooru/internal/structs"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db    *gorm.DB
	model *models.UserModel
	cfg   *config.Config
	log   *zap.Logger
}

func NewUserHandler(db *gorm.DB, cfg *config.Config, log *zap.Logger) *UserHandler {
	return &UserHandler{
		db,
		models.NewUserModel(db),
		cfg,
		log,
	}
}

func (u *UserHandler) Create(c echo.Context) error {
	u.log.Debug("UserHandler: Create")
	var request structs.SignUp
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}
	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	data, err := u.model.Create(request.ToTable())
	if err != nil {
		u.log.Debug("Failed to create user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while creating user")
	}
	return helpers.Response(c, http.StatusOK, data, "")
}

func (u *UserHandler) GetByID(c echo.Context) error {
	u.log.Debug("UserHandler: GetByID")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}
	data, err := u.model.GetByID(id)
	if err != nil {
		u.log.Debug("Failed to get user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while getting user")
	}
	return helpers.Response(c, http.StatusOK, data, "")
}

func (u *UserHandler) Update(c echo.Context) error {
	u.log.Debug("UserHandler: Update")
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
	if request.Name == "" && request.Email == "" {
		return helpers.Response(c, http.StatusBadRequest, nil, "Nothing to update")
	}

	user := request.ToTable()
	user.ID = id

	data, err := u.model.Update(user)
	if err != nil {
		u.log.Debug("Failed to update user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while updating user")
	}

	return helpers.Response(c, http.StatusOK, data, "")
}

func (u *UserHandler) SignUp(c echo.Context) error {
	u.log.Debug("UserHandler: SignUp")
	var request structs.SignUp

	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}

	if u.cfg.AppConfig.EnforceEmail {
		if request.Email == "" {
			return helpers.Response(c, http.StatusBadRequest, nil, "Email is enforced by administrator")
		}
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

	user := request.ToTable()
	permission := structs.Permission{}
	if u.cfg.AppConfig.EnforceEmail {
		permission.Permission = structs.Read
	} else {
		permission.Permission = structs.Write + structs.Read
	}

	user.Permission = permission

	data, err := u.model.Create(user)
	if err != nil {
		u.log.Error("Failed to create user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while creating user")
	}

	token, err := helpers.GenerateJWT(data.ID, data.Name, u.cfg.JWT.Secret)
	if err != nil {
		u.log.Error("Error while generating token", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to generate token")
	}

	return helpers.Response(c, http.StatusOK, data.ToAuthResponse(), token)
}

func (u *UserHandler) SignIn(c echo.Context) error {
	u.log.Debug("UserHandler: SignIn")
	var request structs.SignIn

	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}

	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	data, err := u.model.GetByNameOrEmail(request.NameOrEmail)
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "User not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(data.Password), []byte(request.Password)); err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Invalid credentials")
	}

	token, err := helpers.GenerateJWT(data.ID, data.Name, u.cfg.JWT.Secret)
	if err != nil {
		u.log.Error("Failed to generate token", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to generate token")
	}

	return helpers.Response(c, http.StatusOK, token, "")
}

func (u *UserHandler) ChangePassword(c echo.Context) error {
	u.log.Debug("UserHandler: ChangePassword")
	userID, err := helpers.GetUserID(c, u.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
	}

	var request structs.UserPassword
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

	user := structs.User{
		ID:       userID,
		Password: hashedPassword,
	}

	data, err := u.model.Update(user)
	if err != nil {
		u.log.Error("Failed to update user's password", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to update user's password")
	}

	return helpers.Response(c, http.StatusOK, data, "")
}

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

	token, err := helpers.GenerateJWT(data.ID, data.Name, u.cfg.JWT.Secret)
	if err != nil {
		u.log.Error("Error while generating token", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to generate token")
	}

	return helpers.Response(c, http.StatusOK, admin, token)
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
