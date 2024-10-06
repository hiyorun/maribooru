package account

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	User struct {
		ID         uuid.UUID `gorm:"primary_key;type:uuid"`
		Name       string    `gorm:"unique;not null"`
		Email      string    `gorm:"unique;default:null"`
		Password   string    `gorm:"not null"`
		CreatedAt  time.Time
		UpdatedAt  time.Time
		DeletedAt  gorm.DeletedAt
		Admin      Admin
		Permission Permission
	}

	UserSlice []User

	UserModel struct {
		db *gorm.DB
	}
)

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}

func NewUserModel(db *gorm.DB) *UserModel {
	return &UserModel{
		db: db,
	}
}

func (u *UserModel) Create(payload User) (User, error) {
	user := payload
	err := u.db.Create(&user).Clauses(clause.Returning{}).Error
	return user, err
}

func (u *UserModel) GetAll(params UserParams) (UserSlice, int64, error) {
	users := []User{}
	tx := u.db.
		Model(&User{}).
		Preload("Admin").
		Preload("Permission").
		Where("name ilike ?", fmt.Sprintf("%%%s%%", params.Keywords))

	if params.IsAdmin {
		tx = tx.InnerJoins("Admin")
	}

	total := int64(0)
	err := tx.Count(&total).Error

	err = tx.Order(params.Sort).
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&users).
		Error

	return users, total, err
}

func (u *UserModel) GetByID(id uuid.UUID) (User, error) {
	user := User{}
	err := u.db.Model(&User{}).Preload("Admin").Preload("Permission").First(&user, id).Error
	return user, err
}

// For authentication purpose >>
func (u *UserModel) GetByNameOrEmail(mailname string) (User, error) {
	user := User{}
	err := u.db.Model(&User{}).
		Where("name = ?", mailname).
		Or(u.db.Where("email = ?", mailname)).
		First(&user).Error
	return user, err
}

// <<

func (u *UserModel) Update(user User) (User, error) {
	res := u.db.Model(&User{}).Where(user.ID).Updates(&user)
	if res.RowsAffected == 0 {
		return User{}, gorm.ErrRecordNotFound
	}
	return u.GetByID(user.ID)
}

func (u *UserModel) Delete(id uuid.UUID) error {
	user, err := u.GetByID(id)
	if err != nil {
		return err
	}

	tx := u.db.Begin()

	if user.Admin != (Admin{}) {
		res := tx.Model(&Admin{}).Delete(&Admin{}, user.Admin.ID)
		if res.RowsAffected == 0 {
			tx.Rollback()
			return errors.New("Unable to delete admin position")
		}
	}

	if user.Permission != (Permission{}) {
		res := tx.Model(&Permission{}).Delete(&Permission{}, user.ID)
		if res.RowsAffected == 0 {
			tx.Rollback()
			return errors.New("Unable to delete user permissions")
		}
	}

	res := tx.Model(&User{}).Delete(&User{}, id)
	if res.RowsAffected == 0 {
		tx.Rollback()
		return gorm.ErrRecordNotFound
	}

	tx.Commit()
	return res.Error
}
