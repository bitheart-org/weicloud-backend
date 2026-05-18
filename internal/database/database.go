package database

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"weicloud-backend/internal/config"
	"weicloud-backend/internal/model"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.User{}, &model.Host{}); err != nil {
		return fmt.Errorf("auto migrate models: %w", err)
	}
	return nil
}

func SeedAdmin(db *gorm.DB, cfg config.Config) error {
	var existing model.User
	err := db.Where("username = ?", cfg.AdminUsername).Take(&existing).Error
	if err == nil {
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("query admin user: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	admin := model.User{
		Username:     cfg.AdminUsername,
		PasswordHash: string(passwordHash),
		DisplayName:  cfg.AdminDisplayName,
		Role:         model.UserRoleAdmin,
		Status:       model.UserStatusActive,
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}
	return nil
}
