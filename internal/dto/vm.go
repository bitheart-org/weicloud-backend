package dto

import "time"

type InstancePayload struct {
	ID             string    `json:"id"`
	IncusInstance  string    `json:"incus_instance"`
	HostID         string    `json:"host_id"`
	OwnerID        *string   `json:"owner_id"`
	Name           string    `json:"name"`
	Image          string    `json:"image"`
	CPUCores       int       `json:"cpu_cores"`
	MemoryBytes    int64     `json:"memory_bytes"`
	DiskRootBytes  int64     `json:"disk_root_bytes"`
	NetworkIngress string    `json:"network_ingress"`
	NetworkEgress  string    `json:"network_egress"`
	Status         string    `json:"status"`
	VNCEnabled     bool      `json:"vnc_enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateVMRequest struct {
	Name           string `json:"name" binding:"required,min=2,max=128"`
	Image          string `json:"image" binding:"required,min=1,max=256"`
	HostID         string `json:"host_id" binding:"required"`
	OwnerID        string `json:"owner_id"`
	CPUCores       int    `json:"cpu_cores" binding:"required,min=1,max=128"`
	MemoryBytes    int64  `json:"memory_bytes" binding:"required,min=268435456"`
	DiskRootBytes  int64  `json:"disk_root_bytes" binding:"required,min=1073741824"`
	NetworkIngress string `json:"network_ingress"`
	NetworkEgress  string `json:"network_egress"`
}

type UpdateVMConfigRequest struct {
	CPUCores    int   `json:"cpu_cores" binding:"required,min=1,max=128"`
	MemoryBytes int64 `json:"memory_bytes" binding:"required,min=268435456"`
}

type ResizeVMDiskRequest struct {
	DiskRootBytes int64 `json:"disk_root_bytes" binding:"required,min=1073741824"`
}

type UpdateVMNetworkRequest struct {
	Ingress string `json:"ingress"`
	Egress  string `json:"egress"`
}

type AssignVMRequest struct {
	OwnerID string `json:"owner_id" binding:"required"`
}

type MigrateVMRequest struct {
	TargetHostID string `json:"target_host_id" binding:"required"`
}

type ResetRootPasswordResponse struct {
	NewPassword string `json:"new_password"`
}
