package service

import (
	"fmt"

	"gorm.io/gorm"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/model"
)

type SystemService struct {
	db *gorm.DB
}

func NewSystemService(db *gorm.DB) *SystemService {
	return &SystemService{db: db}
}

func (s *SystemService) Dashboard() (dto.DashboardPayload, error) {
	var (
		hostsTotal       int64
		hostsOnline      int64
		vmsTotal         int64
		vmsRunning       int64
		vmsStopped       int64
		cpuCoresTotal    int64
		cpuCoresUsed     int64
		memoryBytesTotal int64
		memoryBytesUsed  int64
	)

	if err := s.db.Model(&model.Host{}).Count(&hostsTotal).Error; err != nil {
		return dto.DashboardPayload{}, fmt.Errorf("count hosts: %w", err)
	}
	if err := s.db.Model(&model.Host{}).Where("status = ?", model.HostStatusOnline).Count(&hostsOnline).Error; err != nil {
		return dto.DashboardPayload{}, fmt.Errorf("count online hosts: %w", err)
	}
	if err := s.db.Model(&model.Instance{}).Count(&vmsTotal).Error; err != nil {
		return dto.DashboardPayload{}, fmt.Errorf("count vms: %w", err)
	}
	if err := s.db.Model(&model.Instance{}).Where("status = ?", model.InstanceStatusRunning).Count(&vmsRunning).Error; err != nil {
		return dto.DashboardPayload{}, fmt.Errorf("count running vms: %w", err)
	}
	if err := s.db.Model(&model.Instance{}).Where("status = ?", model.InstanceStatusStopped).Count(&vmsStopped).Error; err != nil {
		return dto.DashboardPayload{}, fmt.Errorf("count stopped vms: %w", err)
	}

	if err := s.db.Model(&model.Host{}).Select("COALESCE(SUM(cpu_cores), 0)").Scan(&cpuCoresTotal).Error; err != nil {
		return dto.DashboardPayload{}, fmt.Errorf("sum host cpu: %w", err)
	}
	if err := s.db.Model(&model.Instance{}).Select("COALESCE(SUM(cpu_cores), 0)").Scan(&cpuCoresUsed).Error; err != nil {
		return dto.DashboardPayload{}, fmt.Errorf("sum vm cpu: %w", err)
	}
	if err := s.db.Model(&model.Host{}).Select("COALESCE(SUM(memory_bytes), 0)").Scan(&memoryBytesTotal).Error; err != nil {
		return dto.DashboardPayload{}, fmt.Errorf("sum host memory: %w", err)
	}
	if err := s.db.Model(&model.Instance{}).Select("COALESCE(SUM(memory_bytes), 0)").Scan(&memoryBytesUsed).Error; err != nil {
		return dto.DashboardPayload{}, fmt.Errorf("sum vm memory: %w", err)
	}

	return dto.DashboardPayload{
		HostsTotal:       hostsTotal,
		HostsOnline:      hostsOnline,
		VMsTotal:         vmsTotal,
		VMsRunning:       vmsRunning,
		VMsStopped:       vmsStopped,
		CPUCoresTotal:    cpuCoresTotal,
		CPUCoresUsed:     cpuCoresUsed,
		MemoryBytesTotal: memoryBytesTotal,
		MemoryBytesUsed:  memoryBytesUsed,
	}, nil
}

func (s *SystemService) ListOperationLogs(page, pageSize int) (dto.ListOperationLogsResponse, error) {
	var (
		total int64
		logs  []model.OperationLog
	)

	if err := s.db.Model(&model.OperationLog{}).Count(&total).Error; err != nil {
		return dto.ListOperationLogsResponse{}, fmt.Errorf("count logs: %w", err)
	}
	if err := s.db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return dto.ListOperationLogsResponse{}, fmt.Errorf("list logs: %w", err)
	}

	items := make([]dto.OperationLogPayload, 0, len(logs))
	for _, item := range logs {
		var instanceID *string
		if item.InstanceID != nil {
			id := item.InstanceID.String()
			instanceID = &id
		}
		items = append(items, dto.OperationLogPayload{
			ID:         item.ID.String(),
			UserID:     item.UserID.String(),
			InstanceID: instanceID,
			Action:     item.Action,
			Detail:     item.Detail,
			CreatedAt:  item.CreatedAt,
		})
	}

	return dto.ListOperationLogsResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
