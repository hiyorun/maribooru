package tag

import (
	"fmt"
	"maribooru/internal/common"
	"maribooru/internal/helpers"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	TagCategory struct {
		ID   uuid.UUID `gorm:"primary_key;type:uuid"`
		Slug string    `gorm:"type:varchar(255);not null;unique"`
		Name string    `gorm:"type:varchar(255)"`
		common.AuditFields
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
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Preload("DeletedBy").
		Model(&TagCategory{}).
		Where("slug ilike ?", fmt.Sprintf("%%%s%%", params.Keywords))

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
	err := m.db.Model(&TagCategory{}).
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Preload("DeletedBy").
		First(&category, id).
		Error

	return category, err
}

func (m *CategoryModel) Update(category TagCategory) (TagCategory, error) {
	res := m.db.Model(&TagCategory{}).Where("id = ?", category.ID).Updates(&category)
	if res.RowsAffected == 0 {
		return TagCategory{}, gorm.ErrRecordNotFound
	}
	return m.GetByID(category.ID)
}

func (m *CategoryModel) Delete(id, userID uuid.UUID) error {
	category, err := m.GetByID(id)
	if err != nil {
		return err
	}
	category.DeletedByID = userID

	_, err = m.Update(category)
	if err != nil {
		return err
	}

	res := m.db.Model(&TagCategory{}).Delete(&TagCategory{}, id)
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
