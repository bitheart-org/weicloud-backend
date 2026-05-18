package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OperationLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	InstanceID *uuid.UUID `gorm:"type:uuid;index" json:"instance_id"`
	Action     string     `gorm:"size:64;not null" json:"action"`
	Detail     string     `gorm:"type:text" json:"detail"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (o *OperationLog) BeforeCreate(_ *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
