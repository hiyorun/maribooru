package structs

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	User struct {
		ID         uuid.UUID `gorm:"primary_key;type:uuid"`
		Name       string    `gorm:"unique;not null"`
		Email      string    `gorm:"unique;default:null"`
		Password   string    `gorm:"not null"`
		CreatedAt  time.Time
		UpdatedAt  time.Time
		DeletedAt  gorm.DeletedAt
		Admin      Admin
		Permission Permission
	}

	UserSlice []User

	UserUpdate struct {
		Name  string `json:"name" validate:"omitempty"`
		Email string `json:"email" validate:"omitempty,email"`
	}

	UserPassword struct {
		Password       string `json:"password" validate:"required,min=8"`
		HashedPassword string `json:"-"`
	}

	UserResponse struct {
		ID         uuid.UUID       `json:"id"`
		Name       string          `json:"name"`
		Email      string          `json:"email,omitempty"`
		CreatedAt  time.Time       `json:"created_at"`
		UpdatedAt  time.Time       `json:"updated_at"`
		Admin      bool            `json:"admin"`
		Permission PermissionLevel `json:"permission"`
	}

	SignUp struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"omitempty,email"`
		UserPassword
	}

	SignIn struct {
		NameOrEmail string `json:"name_or_email" validate:"required"`
		UserPassword
	}

	AuthResponse struct {
		ID    uuid.UUID `json:"id"`
		Name  string    `json:"name"`
		Email string    `json:"email"`
	}
)

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}

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
