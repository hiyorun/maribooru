package models

import (
	"errors"
	"fmt"
	"maribooru/internal/structs"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserModel struct {
	db *gorm.DB
}

func NewUserModel(db *gorm.DB) *UserModel {
	return &UserModel{
		db: db,
	}
}

func (u *UserModel) AssignAdmin(uuid uuid.UUID) (structs.Admin, error) {
	admin := structs.Admin{
		UserID: uuid,
	}
	err := u.db.Model(&structs.Admin{}).Create(&admin).Clauses(clause.Returning{}).Error
	return admin, err
}

func (u *UserModel) RemoveAdmin(uuid uuid.UUID) (structs.Admin, error) {
	admin := structs.Admin{
		UserID: uuid,
	}
	err := u.db.Model(&structs.Admin{}).Where("user_id = ?", uuid).Delete(&admin).Error
	return admin, err
}

func (u *UserModel) Create(payload structs.User) (structs.User, error) {
	user := payload
	err := u.db.Create(&user).Clauses(clause.Returning{}).Error
	return user, err
}

func (u *UserModel) GetAll(bounds structs.PagedRequest) (structs.UserSlice, int64, error) {
	users := []structs.User{}
	err := u.db.
		Model(&structs.User{}).
		Preload("Admin").
		Preload("Permission").
		Limit(bounds.Limit).
		Offset(bounds.Offset).
		Where("name ilike ?", fmt.Sprintf("%%%s%%", bounds.Keywords)).
		Order(bounds.Sort).
		Find(&users).
		Error

	total := int64(0)
	err = u.db.
		Model(&structs.User{}).
		Where("name ilike ?", fmt.Sprintf("%%%s%%", bounds.Keywords)).
		Count(&total).Error

	return users, total, err
}

func (u *UserModel) GetAllAdmin(bounds structs.PagedRequest) (structs.UserSlice, int64, error) {
	users := []structs.User{}
	err := u.db.
		Model(&structs.User{}).
		InnerJoins("Admin").
		Limit(bounds.Limit).
		Offset(bounds.Offset).
		Where("name ilike ?", fmt.Sprintf("%%%s%%", bounds.Keywords)).
		Order(bounds.Sort).
		Find(&users).
		Error

	total := int64(0)
	err = u.db.
		Model(&structs.User{}).
		InnerJoins("Admin").
		Where("name ilike ?", fmt.Sprintf("%%%s%%", bounds.Keywords)).
		Count(&total).Error

	return users, total, err
}

func (u *UserModel) GetByID(id uuid.UUID) (structs.User, error) {
	user := structs.User{}
	err := u.db.Model(&structs.User{}).Preload("Admin").Preload("Permission").First(&user, id).Error
	return user, err
}

func (u *UserModel) GetByName(name string) (structs.User, error) {
	user := structs.User{}
	err := u.db.Model(&structs.User{}).
		Where("name = ?", name).
		First(&user).Error
	return user, err
}

func (u *UserModel) GetByNameOrEmail(mailname string) (structs.User, error) {
	user := structs.User{}
	err := u.db.Model(&structs.User{}).
		Where("name = ?", mailname).
		Or(u.db.Where("email = ?", mailname)).
		First(&user).Error
	return user, err
}

func (u *UserModel) Update(user structs.User) (structs.User, error) {
	res := u.db.Model(&structs.User{}).Where(user.ID).Updates(&user)
	if res.RowsAffected == 0 {
		return structs.User{}, gorm.ErrRecordNotFound
	}
	return u.GetByID(user.ID)
}

func (u *UserModel) Delete(id uuid.UUID) error {
	user, err := u.GetByID(id)
	if err != nil {
		return err
	}

	tx := u.db.Begin()

	if user.Admin != (structs.Admin{}) {
		res := tx.Model(&structs.Admin{}).Delete(&structs.Admin{}, user.Admin.ID)
		if res.RowsAffected == 0 {
			tx.Rollback()
			return errors.New("Unable to delete admin position")
		}
	}

	if user.Permission != (structs.Permission{}) {
		res := tx.Model(&structs.Permission{}).Delete(&structs.Permission{}, user.ID)
		if res.RowsAffected == 0 {
			tx.Rollback()
			return errors.New("Unable to delete user permissions")
		}
	}

	res := tx.Model(&structs.User{}).Delete(&structs.User{}, id)
	if res.RowsAffected == 0 {
		tx.Rollback()
		return gorm.ErrRecordNotFound
	}

	tx.Commit()
	return res.Error
}
