package db

import (
	"fmt"
	"maribooru/internal/config"
	"maribooru/internal/structs"

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

	db, err := gorm.Open(postgres.Open(dc), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", zap.Error(err))
	}

	db.AutoMigrate(structs.User{}, structs.Admin{}, structs.AppSettings{})

	if err := db.First(&structs.AppSettings{}).Where("key = ?", "ADMIN_CREATED").Error; err != nil {
		adminSettings := structs.AppSettings{Key: "ADMIN_CREATED", ValueBool: false}
		db.Create(&adminSettings)
	}

	return db, err
}
