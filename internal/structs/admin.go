package structs

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Admin struct {
		ID        uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
		UserID    uuid.UUID `gorm:"type:uuid"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt
	}
)
