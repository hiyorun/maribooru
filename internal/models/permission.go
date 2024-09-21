package models

import (
	"maribooru/internal/structs"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

func (p *PermissionModel) SetPermission(id uuid.UUID, permission structs.Permission) error {
	res := p.db.Model(&structs.Permission{}).Where("user_id = ?", id).Updates(&permission)
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
