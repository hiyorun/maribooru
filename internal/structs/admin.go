package structs

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Admin struct {
		ID        uuid.UUID `gorm:"primary_key;type:uuid"`
		UserID    uuid.UUID `gorm:"type:uuid;unique"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt
	}
	AdminResponse struct {
		AdminID   uuid.UUID `json:"admin_id"`
		UserID    uuid.UUID `json:"user_id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)

func (a *Admin) BeforeCreate(tx *gorm.DB) error {
	a.ID = uuid.New()
	return nil
}

func (a *Admin) ToResponse() AdminResponse {
	return AdminResponse{
		AdminID:   a.ID,
		UserID:    a.UserID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}
