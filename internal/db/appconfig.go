package db

import (
	"maribooru/internal/config"
	"maribooru/internal/setting"

	"gorm.io/gorm"
)

func FetchSettings(cfg *config.Config, db *gorm.DB) {
	adminSettings := setting.AppSetting{}
	if err := db.First(&adminSettings).Where("key = ?", "ADMIN_CREATED").Error; err != nil {
		adminSettings := setting.AppSetting{Key: "ADMIN_CREATED", ValueBool: false}
		db.Create(&adminSettings)
	}

	cfg.AppConfig.AdminCreated = adminSettings.ValueBool
}
