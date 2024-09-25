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
)

func (a *Admin) BeforeCreate(tx *gorm.DB) error {
	a.ID = uuid.New()
	return nil
}
