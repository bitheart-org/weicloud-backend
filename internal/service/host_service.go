package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/incus"
	"weicloud-backend/internal/model"
)

type HostService struct {
	db      *gorm.DB
	manager *incus.Manager
}

func NewHostService(db *gorm.DB, manager *incus.Manager) *HostService {
	return &HostService{
		db:      db,
		manager: manager,
	}
}

func (s *HostService) LoadRegisteredHosts() error {
	var hosts []model.Host
	if err := s.db.Find(&hosts).Error; err != nil {
		return fmt.Errorf("load hosts: %w", err)
	}

	for _, host := range hosts {
		if err := s.manager.RegisterHost(host); err != nil {
			return fmt.Errorf("register host %s: %w", host.ID, err)
		}
	}
	return nil
}

func (s *HostService) List() ([]model.Host, error) {
	var hosts []model.Host
	if err := s.db.Order("created_at DESC").Find(&hosts).Error; err != nil {
		return nil, fmt.Errorf("list hosts: %w", err)
	}
	return hosts, nil
}

func (s *HostService) GetByID(id string) (model.Host, error) {
	var host model.Host
	if err := s.db.Where("id = ?", id).Take(&host).Error; err != nil {
		return model.Host{}, err
	}
	return host, nil
}

func (s *HostService) Create(ctx context.Context, req dto.CreateHostRequest) (model.Host, error) {
	host := model.Host{
		Name:        req.Name,
		Address:     req.Address,
		Certificate: req.Certificate,
		Key:         req.Key,
		Status:      model.HostStatusOffline,
	}

	if err := s.db.Create(&host).Error; err != nil {
		return model.Host{}, fmt.Errorf("create host: %w", err)
	}

	snapshot, err := s.manager.SyncHost(ctx, host)
	if err == nil {
		if applyErr := s.applySnapshot(host.ID.String(), snapshot); applyErr != nil {
			return model.Host{}, applyErr
		}
	}

	return s.GetByID(host.ID.String())
}

func (s *HostService) Update(ctx context.Context, hostID string, req dto.UpdateHostRequest) (model.Host, error) {
	host, err := s.GetByID(hostID)
	if err != nil {
		return model.Host{}, err
	}

	if req.Name != "" {
		host.Name = req.Name
	}
	if req.Address != "" {
		host.Address = req.Address
	}
	if req.Certificate != "" {
		host.Certificate = req.Certificate
	}
	if req.Key != "" {
		host.Key = req.Key
	}
	if req.Status != "" {
		host.Status = req.Status
	}

	if err := s.db.Save(&host).Error; err != nil {
		return model.Host{}, fmt.Errorf("update host: %w", err)
	}

	if req.Address != "" || req.Certificate != "" || req.Key != "" {
		s.manager.RemoveHost(hostID)
	}

	snapshot, syncErr := s.manager.SyncHost(ctx, host)
	if syncErr == nil {
		if applyErr := s.applySnapshot(hostID, snapshot); applyErr != nil {
			return model.Host{}, applyErr
		}
	}

	return s.GetByID(hostID)
}

func (s *HostService) Delete(hostID string) error {
	result := s.db.Delete(&model.Host{}, "id = ?", hostID)
	if result.Error != nil {
		return fmt.Errorf("delete host: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	s.manager.RemoveHost(hostID)
	return nil
}

func (s *HostService) Sync(ctx context.Context, hostID string) (model.Host, error) {
	host, err := s.GetByID(hostID)
	if err != nil {
		return model.Host{}, err
	}

	snapshot, err := s.manager.SyncHost(ctx, host)
	if err != nil {
		host.Status = model.HostStatusOffline
		if saveErr := s.db.Model(&model.Host{}).Where("id = ?", hostID).Update("status", model.HostStatusOffline).Error; saveErr != nil {
			return model.Host{}, fmt.Errorf("sync host failed and update status failed: %w", saveErr)
		}
		return host, err
	}

	if err := s.applySnapshot(hostID, snapshot); err != nil {
		return model.Host{}, err
	}

	return s.GetByID(hostID)
}

func (s *HostService) applySnapshot(hostID string, snapshot incus.ResourceSnapshot) error {
	updates := map[string]any{
		"cpu_cores":    snapshot.CPUCores,
		"memory_bytes": snapshot.MemoryBytes,
		"status":       snapshot.Status,
		"last_seen_at": snapshot.LastSeenAt,
	}
	if err := s.db.Model(&model.Host{}).Where("id = ?", hostID).Updates(updates).Error; err != nil {
		return fmt.Errorf("apply host snapshot: %w", err)
	}
	return nil
}

func ToHostPayload(host model.Host) dto.HostPayload {
	return dto.HostPayload{
		ID:          host.ID.String(),
		Name:        host.Name,
		Address:     host.Address,
		CPUCores:    host.CPUCores,
		MemoryBytes: host.MemoryBytes,
		Status:      host.Status,
		LastSeenAt:  host.LastSeenAt,
		CreatedAt:   host.CreatedAt,
		UpdatedAt:   host.UpdatedAt,
	}
}
