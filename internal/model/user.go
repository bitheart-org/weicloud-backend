package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
)

const (
	UserStatusActive   = "active"
	UserStatusDisabled = "disabled"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:256;not null" json:"-"`
	DisplayName  string    `gorm:"size:128" json:"display_name"`
	Email        string    `gorm:"size:128" json:"email"`
	Role         string    `gorm:"size:16;not null;default:user" json:"role"`
	Status       string    `gorm:"size:16;not null;default:active" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
