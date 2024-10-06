package structs

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	PermissionLevel int
	Permission      struct {
		UserID     uuid.UUID `gorm:"type:uuid;primary_key"`
		Permission PermissionLevel
		CreatedAt  time.Time
		UpdatedAt  time.Time
		DeletedAt  gorm.DeletedAt
	}

	PermissionRequest struct {
		UserID     uuid.UUID       `json:"user_id" validate:"required"`
		Permission PermissionLevel `json:"permission_level" validate:"required"`
	}

	PermissionResponse struct {
		UserID     uuid.UUID       `json:"user_id"`
		Permission PermissionLevel `json:"permission_level"`
		UpdatedAt  time.Time       `json:"updated_at"`
	}
)

const (
	Read PermissionLevel = 1 << iota
	Write
	Approve
	Moderate
)

func (p *PermissionRequest) ToTable() Permission {
	return Permission{
		Permission: p.Permission,
		UserID:     p.UserID,
	}
}

func (p *Permission) ToResponse() PermissionResponse {
	return PermissionResponse{
		UserID:     p.UserID,
		Permission: p.Permission,
		UpdatedAt:  p.UpdatedAt,
	}
}
