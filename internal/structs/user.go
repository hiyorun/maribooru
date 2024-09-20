package structs

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	User struct {
		ID        uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
		Name      string    `gorm:"unique;not null"`
		Email     string    `gorm:"unique;default:null"`
		Password  string    `gorm:"not null"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt
		Admin     Admin
	}

	SignUp struct {
		Name           string `json:"name" validate:"required"`
		Email          string `json:"email" validate:"omitempty,email"`
		Password       string `json:"password" validate:"required,min=8"`
		HashedPassword string `json:"-"`
	}

	SignIn struct {
		NameOrEmail string `json:"name_or_email" validate:"required"`
		Password    string `json:"password" validate:"required"`
	}

	AuthResponse struct {
		ID    uuid.UUID `json:"id"`
		Name  string    `json:"name"`
		Email string    `json:"email"`
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

func (u *User) ToAuthResponse() AuthResponse {
	return AuthResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}
}
