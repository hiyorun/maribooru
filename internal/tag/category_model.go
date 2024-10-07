package tag

import (
	"fmt"
	"maribooru/internal/helpers"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	TagCategory struct {
		ID               uuid.UUID `gorm:"primary_key;type:uuid"`
		PhoneticCategory string    `gorm:"type:varchar(255);not null;unique"`
		ReadableCategory string    `gorm:"type:varchar(255)"`
		CreatedAt        time.Time
		CreatedBy        uuid.UUID
		UpdatedAt        time.Time
		UpdatedBy        uuid.UUID
		DeletedAt        gorm.DeletedAt
		DeletedBy        uuid.UUID
	}

	TagCategorySlice []TagCategory

	CategoryModel struct {
		db *gorm.DB
	}
)

func (c *TagCategory) BeforeCreate(tx *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}

func NewCategoryModel(db *gorm.DB) *CategoryModel {
	return &CategoryModel{
		db: db,
	}
}

func (m *CategoryModel) Create(category TagCategory) (TagCategory, error) {
	err := m.db.Create(&category).Clauses(clause.Returning{}).Error
	return category, err
}

func (m *CategoryModel) GetAll(params helpers.GenericPagedQuery) (TagCategorySlice, int64, error) {
	categories := TagCategorySlice{}

	tx := m.db.
		Model(&TagCategory{}).
		Where("phonetic_category ilike ?", fmt.Sprintf("%%%s%%", params.Keywords))

	total := int64(0)
	err := tx.Count(&total).Error

	err = tx.Order(params.Sort).
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&categories).
		Error

	return categories, total, err
}

func (m *CategoryModel) GetByID(id uuid.UUID) (TagCategory, error) {
	category := TagCategory{}
	err := m.db.Model(&TagCategory{}).First(&category, id).Error
	return category, err
}

func (m *CategoryModel) Update(category TagCategory) (TagCategory, error) {
	res := m.db.Model(&TagCategory{}).Where("id = ?", category.ID).Updates(&category)
	if res.RowsAffected == 0 {
		return TagCategory{}, gorm.ErrRecordNotFound
	}
	return m.GetByID(category.ID)
}

func (m *CategoryModel) Delete(id uuid.UUID) error {
	res := m.db.Model(&TagCategory{}).Delete(&TagCategory{}, id)
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
