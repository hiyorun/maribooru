package models

import (
	"maribooru/internal/structs"

	"gorm.io/gorm"
)

type (
	SettingsModel struct {
		db *gorm.DB
	}
)

func NewSettingsModel(db *gorm.DB) *SettingsModel {
	return &SettingsModel{
		db: db,
	}
}

func (s *SettingsModel) Get() (structs.AppSettingsSlice, error) {
	settings := []structs.AppSettings{}
	err := s.db.Find(&settings).Error
	return settings, err
}

func (s *SettingsModel) GetByKey(key string) (structs.AppSettings, error) {
	settings := structs.AppSettings{}
	err := s.db.Where("key = ?", key).First(&settings).Error
	return settings, err
}

func (s *SettingsModel) Update(settings structs.AppSettings) error {
	res := s.db.Model(&structs.AppSettings{}).Where("key = ?", settings.Key).Updates(&settings)
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
