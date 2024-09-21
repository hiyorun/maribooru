package db

import (
	"maribooru/internal/config"
	"maribooru/internal/structs"

	"gorm.io/gorm"
)

func FetchSettings(cfg *config.Config, db *gorm.DB) {
	adminSettings := structs.AppSettings{}
	if err := db.First(&adminSettings).Where("key = ?", "ADMIN_CREATED").Error; err != nil {
		adminSettings := structs.AppSettings{Key: "ADMIN_CREATED", ValueBool: false}
		db.Create(&adminSettings)
	}

	cfg.AppConfig.AdminCreated = adminSettings.ValueBool
}
