package permission

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	Permission struct {
		UserID     uuid.UUID `gorm:"type:uuid;primary_key"`
		Permission Level
		CreatedAt  time.Time
		UpdatedAt  time.Time
		DeletedAt  gorm.DeletedAt
	}
	Model struct {
		db *gorm.DB
	}
)

func NewModel(db *gorm.DB) *Model {
	return &Model{
		db: db,
	}
}

func (p *Model) GetByUserID(id uuid.UUID) (Permission, error) {
	permission := Permission{}
	err := p.db.Model(&Permission{}).Where("user_id = ?", id).First(&permission).Error
	return permission, err
}

func (p *Model) SetPermission(permission Permission) (Permission, error) {
	res := p.db.Model(&Permission{}).Where("user_id = ?", permission.UserID).Updates(permission)
	if res.RowsAffected == 0 {
		res := permission
		if err := p.db.Model(&Permission{}).Create(&res).Clauses(clause.Returning{}).Error; err != nil {
			return Permission{}, err
		}
		return res, nil
	}
	return p.GetByUserID(permission.UserID)
}
