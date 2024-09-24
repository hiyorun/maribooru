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

// INTERNAL CRUD ------------------------------- //

func (u *UserHandler) create(c echo.Context, user structs.User) error {
	u.log.Debug("UserHandler: Create")

	data, err := u.model.Create(user)
	if err != nil {
		u.log.Debug("Failed to create user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while creating user")
	}

	token, err := helpers.GenerateJWT(data.ID, data.Name, u.cfg.JWT.Secret, u.cfg.AppConfig.TokenLifetime)
	if err != nil {
		u.log.Error("Error while generating token", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to generate token")
	}

	return helpers.Response(c, http.StatusOK, data, token)
}

func (u *UserHandler) getAllUser(c echo.Context) error {
	u.log.Debug("UserHandler: GetAll")

	bounds := structs.PagedRequest{
		Limit:  50,
		Offset: 0,
	}
	if err := c.Bind(&bounds); err != nil {
		u.log.Error("Failed to set limit and offset, defaulting to 50 limit and 0 offset", zap.Error(err))
	}

	data, err := u.model.GetAll(bounds)
	if err != nil {
		u.log.Error("Failed to get users", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while getting users")
	}

	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}

func (u *UserHandler) getByID(c echo.Context, id uuid.UUID) error {
	u.log.Debug("UserHandler: GetByID")

	data, err := u.model.GetByID(id)
	if err != nil {
		u.log.Debug("Failed to get user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while getting user")
	}

	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}

func (u *UserHandler) update(c echo.Context, user structs.User) error {
	u.log.Debug("UserHandler: Update")

	data, err := u.model.Update(user)
	if err != nil {
		u.log.Debug("Failed to update user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while updating user")
	}

	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}

func (u *UserHandler) delete(c echo.Context, id uuid.UUID) error {
	u.log.Debug("UserHandler: Delete")

	if err := u.model.Delete(id); err != nil {
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while deleting user")
	}
	return helpers.Response(c, http.StatusOK, nil, "")
}

// END OF INTERNAL CRUD ------------------------ //

func (u *UserHandler) SignUp(c echo.Context) error {
	u.log.Debug("UserHandler: SignUp")
	var request structs.SignUp
	u.bindAndValidate(c, &request)

	if u.cfg.AppConfig.EnforceEmail {
		if request.Email == "" {
			return helpers.Response(c, http.StatusBadRequest, nil, "Email is enforced by administrator")
		}
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
		permission.Permission = structs.Write | structs.Read
	}

	user.Permission = permission

	return u.create(c, user)
}

func (u *UserHandler) SignIn(c echo.Context) error {
	u.log.Debug("UserHandler: SignIn")
	var request structs.SignIn
	u.bindAndValidate(c, &request)

	data, err := u.model.GetByNameOrEmail(request.NameOrEmail)
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "User not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(data.Password), []byte(request.Password)); err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Invalid credentials")
	}

	token, err := helpers.GenerateJWT(data.ID, data.Name, u.cfg.JWT.Secret, u.cfg.AppConfig.TokenLifetime)
	if err != nil {
		u.log.Error("Failed to generate token", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to generate token")
	}

	return helpers.Response(c, http.StatusOK, token, "")
}

func (u *UserHandler) GetAllUsers(c echo.Context) error {
	u.log.Debug("UserHandler: GetAllUsers")
	return u.getAllUser(c)
}

func (u *UserHandler) SelfGet(c echo.Context) error {
	u.log.Debug("UserHandler: GetSelf")
	id, err := helpers.GetUserID(c, u.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Failed to get your details")
	}

	return u.getByID(c, id)
}

func (u *UserHandler) GetUserByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}

	return u.getByID(c, id)
}

func (u *UserHandler) ChangePassword(c echo.Context) error {
	u.log.Debug("UserHandler: ChangePassword")
	userID, err := helpers.GetUserID(c, u.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
	}

	var request structs.UserPassword
	u.bindAndValidate(c, &request)

	hashedPassword, err := helpers.PasswordHash(request.Password)
	if err != nil {
		u.log.Error("Error while hashing password", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while hashing password")
	}

	user := structs.User{
		ID:       userID,
		Password: hashedPassword,
	}

	return u.update(c, user)
}

func (u *UserHandler) SelfUpdate(c echo.Context) error {
	paramID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}

	id, err := helpers.GetUserID(c, u.cfg.JWT.Secret)
	if id != paramID {
		return helpers.Response(c, http.StatusForbidden, nil, "You can't update other user")
	}

	var request structs.UserUpdate
	u.bindAndValidate(c, &request)

	if request.Name == "" && request.Email == "" {
		return helpers.Response(c, http.StatusBadRequest, nil, "Nothing to update")
	}

	return u.update(c, request.ToTable(id))
}

func (u *UserHandler) bindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(&req); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}

	if err := c.Validate(&req); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}
	return nil
}
