package account

import (
	"errors"
	"maribooru/internal/config"
	"maribooru/internal/helpers"
	"maribooru/internal/permission"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type (
	UserUpdate struct {
		Name  string `json:"name" validate:"omitempty"`
		Email string `json:"email" validate:"omitempty,email"`
	}

	UserPassword struct {
		OldPassword string `json:"old_password" validate:"required,min=8"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	UserResponse struct {
		ID         uuid.UUID        `json:"id"`
		Name       string           `json:"name"`
		Email      string           `json:"email,omitempty"`
		CreatedAt  time.Time        `json:"created_at"`
		UpdatedAt  time.Time        `json:"updated_at"`
		Admin      bool             `json:"admin"`
		Permission permission.Level `json:"permission"`
	}

	UserParams struct {
		helpers.GenericPagedQuery
		IsAdmin bool `query:"is_admin"`
	}

	SignUp struct {
		Name           string `json:"name" validate:"required"`
		Email          string `json:"email" validate:"omitempty,email"`
		Password       string `json:"password" validate:"required,min=8"`
		HashedPassword string `json:"-"`
	}

	SignIn struct {
		NameOrEmail    string `json:"name_or_email" validate:"required"`
		Password       string `json:"password" validate:"required,min=8"`
		HashedPassword string `json:"-"`
	}

	AuthResponse struct {
		ID    uuid.UUID `json:"id"`
		Name  string    `json:"name"`
		Email string    `json:"email"`
	}

	UserHandler struct {
		db    *gorm.DB
		model *UserModel
		cfg   *config.Config
		log   *zap.Logger
	}
)

func (s *SignUp) ToTable() User {
	user := User{
		Name:     s.Name,
		Password: s.HashedPassword,
	}
	if s.Email != "" {
		user.Email = s.Email
	}
	return user
}

func (u *UserUpdate) ToTable(id uuid.UUID) User {
	user := User{
		ID: id,
	}
	if u.Name != "" {
		user.Name = u.Name
	}
	if u.Email != "" {
		user.Email = u.Email
	}
	return user
}

func (u *User) ToAuthResponse() AuthResponse {
	return AuthResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}
}

func (u *User) ToResponse(includeEmail bool) UserResponse {
	user := UserResponse{
		ID:         u.ID,
		Name:       u.Name,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
		Admin:      u.Admin != (Admin{}),
		Permission: u.Permission.Permission,
	}

	if includeEmail {
		user.Email = u.Email
	}

	return user
}

func (u UserSlice) ToResponse(includeEmail bool) []UserResponse {
	response := make([]UserResponse, 0)
	for _, user := range u {
		response = append(response, user.ToResponse(includeEmail))
	}
	return response
}

func NewUserHandler(db *gorm.DB, cfg *config.Config, log *zap.Logger) *UserHandler {
	return &UserHandler{
		db,
		NewUserModel(db),
		cfg,
		log,
	}
}

// INTERNAL CRUD ------------------------------- //

func (u *UserHandler) create(c echo.Context, user User) error {
	u.log.Debug("UserHandler: Create")

	data, err := u.model.Create(user)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return helpers.Response(c, http.StatusConflict, nil, "User with this name/email already exists")
		}
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

func (u *UserHandler) getAllUser(c echo.Context, includeEmail bool) error {
	u.log.Debug("UserHandler: GetAll")

	params := UserParams{
		GenericPagedQuery: helpers.GenericPagedQuery{
			Limit:    50,
			Offset:   0,
			Keywords: "",
			Sort:     "",
		},
		IsAdmin: false,
	}
	if err := c.Bind(&params); err != nil {
		u.log.Error("Failed to set limit and offset, defaulting to 50 limit and 0 offset", zap.Error(err))
	}

	data, total, err := u.model.GetAll(params)
	if err != nil {
		u.log.Error("Failed to get users", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while getting users")
	}

	paged := helpers.PageData(data.ToResponse(includeEmail), int(total), params.Offset, params.Limit)

	if total == 0 {
		return helpers.Response(c, http.StatusNotFound, paged, "There's no user")
	}

	return helpers.Response(c, http.StatusOK, paged, "")
}

func (u *UserHandler) getByID(c echo.Context, id uuid.UUID, includeEmail bool) error {
	u.log.Debug("UserHandler: GetByID")

	data, err := u.model.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "User not found")
		}
		u.log.Debug("Failed to get user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while getting user")
	}

	return helpers.Response(c, http.StatusOK, data.ToResponse(includeEmail), "")
}

func (u *UserHandler) update(c echo.Context, user User) error {
	u.log.Debug("UserHandler: Update")

	data, err := u.model.Update(user)
	if err != nil {
		u.log.Debug("Failed to update user", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while updating user")
	}

	return helpers.Response(c, http.StatusOK, data.ToResponse(true), "")
}

func (u *UserHandler) delete(c echo.Context, id uuid.UUID) error {
	u.log.Debug("UserHandler: Delete")

	if err := u.model.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "User not found")
		}
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while deleting user")
	}
	return helpers.Response(c, http.StatusOK, nil, "")
}

// END OF INTERNAL CRUD ------------------------ //

func (u *UserHandler) SignUp(c echo.Context) error {
	u.log.Debug("UserHandler: SignUp")
	var request SignUp
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}

	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

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
	userPermission := permission.Permission{}
	if u.cfg.AppConfig.EnforceEmail {
		userPermission.Permission = permission.Read
	} else {
		userPermission.Permission = permission.Write | permission.Read
	}

	user.Permission = userPermission

	return u.create(c, user)
}

func (u *UserHandler) SignIn(c echo.Context) error {
	u.log.Debug("UserHandler: SignIn")
	var request SignIn
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
		u.log.Error("Error while comparing passwords", zap.Error(err))
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
	return u.getAllUser(c, false)
}

func (u *UserHandler) SelfGet(c echo.Context) error {
	u.log.Debug("UserHandler: GetSelf")
	id, err := helpers.GetUserID(c, u.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Failed to get your details")
	}

	return u.getByID(c, id, true)
}

func (u *UserHandler) GetUserByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}

	return u.getByID(c, id, true)
}

func (u *UserHandler) ChangePassword(c echo.Context) error {
	u.log.Debug("UserHandler: ChangePassword")
	id, err := helpers.GetUserID(c, u.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
	}

	var request UserPassword
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}

	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	user, err := u.model.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "User not found")
		}
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to change password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.OldPassword)); err != nil {
		u.log.Error("Error while comparing passwords", zap.Error(err))
		return helpers.Response(c, http.StatusUnauthorized, nil, "Old and new password don't match")
	}

	hashedPassword, err := helpers.PasswordHash(request.NewPassword)
	if err != nil {
		u.log.Error("Error while hashing password", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "There was an error while hashing password")
	}

	user.Password = hashedPassword

	return u.update(c, user)
}

func (u *UserHandler) SelfUpdate(c echo.Context) error {
	id, err := helpers.GetUserID(c, u.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
	}

	var request UserUpdate
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}

	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	if request.Name == "" && request.Email == "" {
		return helpers.Response(c, http.StatusBadRequest, nil, "Nothing to update")
	}

	return u.update(c, request.ToTable(id))
}

func (u *UserHandler) SelfDelete(c echo.Context) error {
	u.log.Debug("UserHandler: SelfDelete")
	id, err := helpers.GetUserID(c, u.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
	}

	return u.delete(c, id)
}
