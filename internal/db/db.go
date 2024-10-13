package db

import (
	"fmt"
	"maribooru/internal/account"
	"maribooru/internal/config"
	"maribooru/internal/permission"
	"maribooru/internal/setting"
	"maribooru/internal/tag"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	dc := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Database.Host,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSL,
	)

	db, err := gorm.Open(postgres.Open(dc), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", zap.Error(err))
	}

	db.AutoMigrate(
		account.User{},
		account.Admin{},
		setting.AppSetting{},
		permission.Permission{},
		tag.TagCategory{},
		tag.Tag{},
	)

	FetchSettings(cfg, db)

	return db, err
}
