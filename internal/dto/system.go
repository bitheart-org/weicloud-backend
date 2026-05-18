package dto

import "time"

type DashboardPayload struct {
	HostsTotal       int64 `json:"hosts_total"`
	HostsOnline      int64 `json:"hosts_online"`
	VMsTotal         int64 `json:"vms_total"`
	VMsRunning       int64 `json:"vms_running"`
	VMsStopped       int64 `json:"vms_stopped"`
	CPUCoresTotal    int64 `json:"cpu_cores_total"`
	CPUCoresUsed     int64 `json:"cpu_cores_used"`
	MemoryBytesTotal int64 `json:"memory_bytes_total"`
	MemoryBytesUsed  int64 `json:"memory_bytes_used"`
}

type OperationLogPayload struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	InstanceID *string   `json:"instance_id"`
	Action     string    `json:"action"`
	Detail     string    `json:"detail"`
	CreatedAt  time.Time `json:"created_at"`
}

type ListOperationLogsResponse struct {
	Items    []OperationLogPayload `json:"items"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

type VMResourcePayload struct {
	CPUNanoseconds int64 `json:"cpu_nanoseconds"`
	MemoryBytes    int64 `json:"memory_bytes"`
}
