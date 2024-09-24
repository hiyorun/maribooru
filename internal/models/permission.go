package models

import (
	"maribooru/internal/structs"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PermissionModel struct {
	db *gorm.DB
}

func NewPermissionModel(db *gorm.DB) *PermissionModel {
	return &PermissionModel{
		db: db,
	}
}

func (p *PermissionModel) GetByUserID(id uuid.UUID) (structs.Permission, error) {
	permission := structs.Permission{}
	err := p.db.Model(&structs.Permission{}).Where("user_id = ?", id).First(&permission).Error
	return permission, err
}

func (p *PermissionModel) IsAdmin(id uuid.UUID) (bool, error) {
	admin := structs.Admin{}
	err := p.db.Model(&structs.Admin{}).Where("user_id = ?", id).First(&admin).Error
	if err != nil {
		return false, err
	}
	return admin.ID != uuid.Nil, nil
}

func (p *PermissionModel) SetPermission(permission structs.Permission) (structs.Permission, error) {
	res := p.db.Model(&structs.Permission{}).Where("user_id = ?", permission.UserID).Updates(permission)
	if res.RowsAffected == 0 {
		res := permission
		if err := p.db.Model(&structs.Permission{}).Create(&res).Clauses(clause.Returning{}).Error; err != nil {
			return structs.Permission{}, err
		}
		return res, nil
	}
	return p.GetByUserID(permission.UserID)
}
