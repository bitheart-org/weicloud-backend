package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	InstanceStatusRunning  = "running"
	InstanceStatusStopped  = "stopped"
	InstanceStatusCreating = "creating"
	InstanceStatusError    = "error"
)

type Instance struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	IncusInstance  string     `gorm:"size:128;not null;uniqueIndex" json:"incus_instance"`
	HostID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"host_id"`
	OwnerID        *uuid.UUID `gorm:"type:uuid;index" json:"owner_id"`
	Name           string     `gorm:"size:128;not null" json:"name"`
	Image          string     `gorm:"size:256;not null" json:"image"`
	CPUCores       int        `gorm:"not null;default:2" json:"cpu_cores"`
	MemoryBytes    int64      `gorm:"not null;default:2147483648" json:"memory_bytes"`
	DiskRootBytes  int64      `gorm:"not null;default:10737418240" json:"disk_root_bytes"`
	NetworkIngress string     `gorm:"size:16" json:"network_ingress"`
	NetworkEgress  string     `gorm:"size:16" json:"network_egress"`
	LoginUsername  string     `gorm:"size:64;not null;default:weicloud" json:"login_username"`
	SSHRemotePort  int        `gorm:"not null;default:0" json:"ssh_remote_port"`
	Status         string     `gorm:"size:16;not null;default:stopped" json:"status"`
	VNCEnabled     bool       `gorm:"not null;default:false" json:"vnc_enabled"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (i *Instance) BeforeCreate(_ *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
