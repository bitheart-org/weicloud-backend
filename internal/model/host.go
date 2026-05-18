package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	HostStatusOnline      = "online"
	HostStatusOffline     = "offline"
	HostStatusMaintenance = "maintenance"
)

type Host struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string     `gorm:"size:128;uniqueIndex;not null" json:"name"`
	Address     string     `gorm:"size:256;not null" json:"address"`
	Certificate string     `gorm:"type:text;not null" json:"-"`
	Key         string     `gorm:"type:text;not null" json:"-"`
	CPUCores    int        `gorm:"not null;default:0" json:"cpu_cores"`
	MemoryBytes int64      `gorm:"not null;default:0" json:"memory_bytes"`
	Status      string     `gorm:"size:16;not null;default:offline" json:"status"`
	LastSeenAt  *time.Time `json:"last_seen_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (h *Host) BeforeCreate(_ *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}
