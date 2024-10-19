package tag

import (
	"fmt"
	"maribooru/internal/common"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	Tag struct {
		ID         uuid.UUID   `gorm:"primary_key;type:uuid"`
		Slug       string      `gorm:"type:varchar(255);not null;uniqueIndex:idx_slug_category"`
		Name       string      `gorm:"type:varchar(255)"`
		CategoryID uuid.UUID   `gorm:"type:uuid;uniqueIndex:idx_slug_category"`
		Category   TagCategory `gorm:"foreignKey:CategoryID"`

		common.AuditFields
	}

	TagSlice []Tag

	TagModel struct {
		db *gorm.DB
	}
)

func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	t.ID = uuid.New()
	return nil
}

func NewTagModel(db *gorm.DB) *TagModel {
	return &TagModel{
		db: db,
	}
}

func (t *TagModel) Create(tag Tag) (Tag, error) {
	err := t.db.Create(&tag).
		Clauses(clause.Returning{}).
		Error
	if err != nil {
		return Tag{}, err
	}

	return t.GetByID(tag.ID)
}

func (t *TagModel) baseSelect() *gorm.DB {
	return t.db.
		Model(&Tag{}).
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Preload("DeletedBy").
		Preload("Category")
}

func (t *TagModel) GetByName(name string) (Tag, error) {
	tag := Tag{}
	err := t.baseSelect().
		Where("slug = ?", name).
		First(&tag).
		Error
	return tag, err
}

func (t *TagModel) GetByID(id uuid.UUID) (Tag, error) {
	tag := Tag{}
	err := t.baseSelect().
		Where("id = ?", id).
		First(&tag).
		Error
	return tag, err
}

func (t *TagModel) GetAll(params TagParams) (TagSlice, int64, error) {
	tags := TagSlice{}
	var total int64

	tx := t.baseSelect().
		Where("slug ilike ?", fmt.Sprintf("%%%s%%", params.Keywords)).
		Offset(params.Offset).
		Limit(params.Limit)

	if params.CategoryID != uuid.Nil {
		tx.Where("category_id = ?", params.CategoryID)
	}

	err := tx.Find(&tags).Error
	if err != nil {
		return nil, 0, err
	}

	err = tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

func (t *TagModel) Update(tag Tag) (Tag, error) {
	res := t.db.Model(&Tag{}).Where("id = ?", tag.ID).Updates(tag)
	if res.RowsAffected == 0 {
		return Tag{}, gorm.ErrRecordNotFound
	}
	return t.GetByID(tag.ID)
}

func (t *TagModel) Delete(id, userID uuid.UUID) error {
	tag, err := t.GetByID(id)
	if err != nil {
		return err
	}
	tag.DeletedByID = userID

	_, err = t.Update(tag)
	if err != nil {
		return err
	}

	res := t.db.Delete(&Tag{}, id)
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
