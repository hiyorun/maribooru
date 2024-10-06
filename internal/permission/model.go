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
		Permission PermissionLevel
		CreatedAt  time.Time
		UpdatedAt  time.Time
		DeletedAt  gorm.DeletedAt
	}
	PermissionModel struct {
		db *gorm.DB
	}
)

func NewPermissionModel(db *gorm.DB) *PermissionModel {
	return &PermissionModel{
		db: db,
	}
}

func (p *PermissionModel) GetByUserID(id uuid.UUID) (Permission, error) {
	permission := Permission{}
	err := p.db.Model(&Permission{}).Where("user_id = ?", id).First(&permission).Error
	return permission, err
}

func (p *PermissionModel) SetPermission(permission Permission) (Permission, error) {
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
