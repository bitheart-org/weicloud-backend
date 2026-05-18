package dto

import "time"

type HostPayload struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Address     string     `json:"address"`
	CPUCores    int        `json:"cpu_cores"`
	MemoryBytes int64      `json:"memory_bytes"`
	Status      string     `json:"status"`
	LastSeenAt  *time.Time `json:"last_seen_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateHostRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=128"`
	Address     string `json:"address" binding:"required,url,max=256"`
	Certificate string `json:"certificate" binding:"required,min=1"`
	Key         string `json:"key" binding:"required,min=1"`
}

type UpdateHostRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=128"`
	Address     string `json:"address" binding:"omitempty,url,max=256"`
	Certificate string `json:"certificate" binding:"omitempty,min=1"`
	Key         string `json:"key" binding:"omitempty,min=1"`
	Status      string `json:"status" binding:"omitempty,oneof=online offline maintenance"`
}
