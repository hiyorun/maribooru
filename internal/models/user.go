package models

import (
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
	err := u.db.Model(&structs.Admin{}).Delete(&admin).Error
	return admin, err
}

func (u *UserModel) Create(payload structs.User) (structs.User, error) {
	user := payload
	err := u.db.Create(&user).Clauses(clause.Returning{}).Error
	return user, err
}

func (u *UserModel) GetAll(bounds structs.PagedRequest) (structs.UserSlice, error) {
	users := []structs.User{}
	err := u.db.
		Model(&structs.User{}).
		Preload("Admin").
		Preload("Permission").
		Limit(bounds.Limit).
		Offset(bounds.Offset).
		Where("name ilike %?%", bounds.Keywords).
		Order(bounds.Sort).
		Find(&users).
		Error
	return users, err
}

func (u *UserModel) GetAllAdmin(bounds structs.PagedRequest) (structs.UserSlice, error) {
	users := []structs.User{}
	err := u.db.
		Model(&structs.User{}).
		InnerJoins("Admin").
		Limit(bounds.Limit).
		Offset(bounds.Offset).
		Where("name ilike %?%", bounds.Keywords).
		Order(bounds.Sort).
		Find(&users).
		Error
	return users, err
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
	res := u.db.Model(&structs.User{}).Delete(&structs.User{}, id)
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return res.Error
}
