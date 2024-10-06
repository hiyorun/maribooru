package account

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	Admin struct {
		ID        uuid.UUID `gorm:"primary_key;type:uuid"`
		UserID    uuid.UUID `gorm:"type:uuid;unique"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt
	}

	AdminModel struct {
		db *gorm.DB
	}
)

func (a *Admin) BeforeCreate(tx *gorm.DB) error {
	a.ID = uuid.New()
	return nil
}

func NewAdminModel(db *gorm.DB) *AdminModel {
	return &AdminModel{
		db: db,
	}
}

func (a *AdminModel) AssignAdmin(uuid uuid.UUID) (Admin, error) {
	admin := Admin{
		UserID: uuid,
	}
	err := a.db.Model(&Admin{}).Create(&admin).Clauses(clause.Returning{}).Error
	return admin, err
}

func (a *AdminModel) RemoveAdmin(uuid uuid.UUID) (Admin, error) {
	admin := Admin{
		UserID: uuid,
	}
	err := a.db.Model(&Admin{}).Where("user_id = ?", uuid).Delete(&admin).Error
	return admin, err
}
