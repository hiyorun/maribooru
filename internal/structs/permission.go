package structs

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	PermissionLevel int
	Permission      struct {
		gorm.Model
		UserID     uuid.UUID `gorm:"type:uuid"`
		Permission PermissionLevel
	}
)

const (
	Read PermissionLevel = 1 << iota
	Write
	Approve
	Moderate
)
