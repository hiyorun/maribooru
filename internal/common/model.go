package common

import (
	"maribooru/internal/account"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	AuditFields struct {
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt

		CreatedByID uuid.UUID    `gorm:"type:uuid"`
		CreatedBy   account.User `gorm:"foreignKey:CreatedByID"`

		UpdatedByID uuid.UUID    `gorm:"type:uuid;default:null"`
		UpdatedBy   account.User `gorm:"foreignKey:UpdatedByID"`

		DeletedByID uuid.UUID    `gorm:"type:uuid;default:null"`
		DeletedBy   account.User `gorm:"foreignKey:DeletedByID"`
	}
)
