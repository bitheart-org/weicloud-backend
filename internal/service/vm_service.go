package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/incus"
	"weicloud-backend/internal/model"
)

type VMService struct {
	db          *gorm.DB
	hostService *HostService
	incus       *incus.Manager
}

func NewVMService(db *gorm.DB, hostService *HostService, incusManager *incus.Manager) *VMService {
	return &VMService{
		db:          db,
		hostService: hostService,
		incus:       incusManager,
	}
}

func (s *VMService) ListAll() ([]model.Instance, error) {
	var items []model.Instance
	if err := s.db.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	return items, nil
}

func (s *VMService) ListByOwner(ownerID string) ([]model.Instance, error) {
	var items []model.Instance
	if err := s.db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list owner instances: %w", err)
	}
	return items, nil
}

func (s *VMService) GetByID(id string) (model.Instance, error) {
	var item model.Instance
	if err := s.db.Where("id = ?", id).Take(&item).Error; err != nil {
		return model.Instance{}, err
	}
	return item, nil
}

func (s *VMService) GetByIDForOwner(id, ownerID string) (model.Instance, error) {
	var item model.Instance
	if err := s.db.Where("id = ? AND owner_id = ?", id, ownerID).Take(&item).Error; err != nil {
		return model.Instance{}, err
	}
	return item, nil
}

func (s *VMService) Create(ctx context.Context, req dto.CreateVMRequest, actorUserID string) (model.Instance, error) {
	host, err := s.hostService.GetByID(req.HostID)
	if err != nil {
		return model.Instance{}, err
	}

	instance := model.Instance{
		IncusInstance:  "vm-" + uuid.NewString(),
		HostID:         host.ID,
		Name:           req.Name,
		Image:          req.Image,
		CPUCores:       req.CPUCores,
		MemoryBytes:    req.MemoryBytes,
		DiskRootBytes:  req.DiskRootBytes,
		NetworkIngress: req.NetworkIngress,
		NetworkEgress:  req.NetworkEgress,
		Status:         model.InstanceStatusCreating,
	}

	if req.OwnerID != "" {
		ownerUUID, parseErr := uuid.Parse(req.OwnerID)
		if parseErr != nil {
			return model.Instance{}, fmt.Errorf("invalid owner id: %w", parseErr)
		}
		instance.OwnerID = &ownerUUID
	}

	if err := s.db.Create(&instance).Error; err != nil {
		return model.Instance{}, fmt.Errorf("create instance record: %w", err)
	}

	createErr := s.incus.CreateInstance(ctx, host, incus.InstanceCreateRequest{
		InstanceName:   instance.IncusInstance,
		Image:          req.Image,
		CPUCores:       req.CPUCores,
		MemoryBytes:    req.MemoryBytes,
		DiskRootBytes:  req.DiskRootBytes,
		NetworkIngress: req.NetworkIngress,
		NetworkEgress:  req.NetworkEgress,
	})
	if createErr != nil {
		_ = s.db.Model(&model.Instance{}).Where("id = ?", instance.ID).Update("status", model.InstanceStatusError).Error
		return model.Instance{}, fmt.Errorf("create instance on host: %w", createErr)
	}

	if err := s.db.Model(&model.Instance{}).Where("id = ?", instance.ID).Update("status", model.InstanceStatusStopped).Error; err != nil {
		return model.Instance{}, fmt.Errorf("update instance status: %w", err)
	}

	if err := s.logOperation(actorUserID, &instance.ID, "create", "create vm "+instance.Name); err != nil {
		return model.Instance{}, err
	}
	return s.GetByID(instance.ID.String())
}

func (s *VMService) Delete(ctx context.Context, vmID string, actorUserID string) error {
	vm, err := s.GetByID(vmID)
	if err != nil {
		return err
	}
	host, err := s.hostService.GetByID(vm.HostID.String())
	if err != nil {
		return err
	}

	_ = s.incus.SetInstanceState(ctx, host, vm.IncusInstance, "stop", true)
	if err := s.incus.DeleteInstance(ctx, host, vm.IncusInstance); err != nil {
		return err
	}

	if err := s.db.Delete(&model.Instance{}, "id = ?", vmID).Error; err != nil {
		return fmt.Errorf("delete instance record: %w", err)
	}
	return s.logOperation(actorUserID, &vm.ID, "destroy", "destroy vm "+vm.Name)
}

func (s *VMService) UpdateConfig(ctx context.Context, vmID string, req dto.UpdateVMConfigRequest, actorUserID string) (model.Instance, error) {
	vm, err := s.GetByID(vmID)
	if err != nil {
		return model.Instance{}, err
	}
	host, err := s.hostService.GetByID(vm.HostID.String())
	if err != nil {
		return model.Instance{}, err
	}

	if err := s.incus.UpdateInstanceConfig(ctx, host, vm.IncusInstance, req.CPUCores, req.MemoryBytes); err != nil {
		return model.Instance{}, err
	}

	vm.CPUCores = req.CPUCores
	vm.MemoryBytes = req.MemoryBytes
	if err := s.db.Save(&vm).Error; err != nil {
		return model.Instance{}, fmt.Errorf("save instance config: %w", err)
	}

	if err := s.logOperation(actorUserID, &vm.ID, "update_config", "update vm cpu/memory "+vm.Name); err != nil {
		return model.Instance{}, err
	}
	return vm, nil
}

func (s *VMService) ResizeDisk(ctx context.Context, vmID string, req dto.ResizeVMDiskRequest, actorUserID string) (model.Instance, error) {
	vm, err := s.GetByID(vmID)
	if err != nil {
		return model.Instance{}, err
	}
	host, err := s.hostService.GetByID(vm.HostID.String())
	if err != nil {
		return model.Instance{}, err
	}

	if err := s.incus.ResizeRootDisk(ctx, host, vm.IncusInstance, req.DiskRootBytes); err != nil {
		return model.Instance{}, err
	}
	vm.DiskRootBytes = req.DiskRootBytes
	if err := s.db.Save(&vm).Error; err != nil {
		return model.Instance{}, fmt.Errorf("save instance disk: %w", err)
	}

	if err := s.logOperation(actorUserID, &vm.ID, "resize_disk", "resize vm disk "+vm.Name); err != nil {
		return model.Instance{}, err
	}
	return vm, nil
}

func (s *VMService) UpdateNetwork(ctx context.Context, vmID string, req dto.UpdateVMNetworkRequest, actorUserID string) (model.Instance, error) {
	vm, err := s.GetByID(vmID)
	if err != nil {
		return model.Instance{}, err
	}
	host, err := s.hostService.GetByID(vm.HostID.String())
	if err != nil {
		return model.Instance{}, err
	}

	if err := s.incus.UpdateNetworkLimits(ctx, host, vm.IncusInstance, req.Ingress, req.Egress); err != nil {
		return model.Instance{}, err
	}

	if req.Ingress != "" {
		vm.NetworkIngress = req.Ingress
	}
	if req.Egress != "" {
		vm.NetworkEgress = req.Egress
	}
	if err := s.db.Save(&vm).Error; err != nil {
		return model.Instance{}, fmt.Errorf("save network limits: %w", err)
	}

	if err := s.logOperation(actorUserID, &vm.ID, "update_network", "update vm network limit "+vm.Name); err != nil {
		return model.Instance{}, err
	}
	return vm, nil
}

func (s *VMService) AssignOwner(vmID string, ownerID string, actorUserID string) (model.Instance, error) {
	vm, err := s.GetByID(vmID)
	if err != nil {
		return model.Instance{}, err
	}

	ownerUUID, err := uuid.Parse(ownerID)
	if err != nil {
		return model.Instance{}, fmt.Errorf("invalid owner id: %w", err)
	}
	vm.OwnerID = &ownerUUID
	if err := s.db.Save(&vm).Error; err != nil {
		return model.Instance{}, fmt.Errorf("assign owner: %w", err)
	}

	if err := s.logOperation(actorUserID, &vm.ID, "assign", "assign vm owner "+vm.Name); err != nil {
		return model.Instance{}, err
	}
	return vm, nil
}

func (s *VMService) Migrate(vmID string, req dto.MigrateVMRequest, actorUserID string) (model.Instance, error) {
	vm, err := s.GetByID(vmID)
	if err != nil {
		return model.Instance{}, err
	}

	targetHost, err := s.hostService.GetByID(req.TargetHostID)
	if err != nil {
		return model.Instance{}, err
	}

	vm.HostID = targetHost.ID
	if err := s.db.Save(&vm).Error; err != nil {
		return model.Instance{}, fmt.Errorf("update vm target host: %w", err)
	}

	if err := s.logOperation(actorUserID, &vm.ID, "migrate", "mark vm target host "+vm.Name); err != nil {
		return model.Instance{}, err
	}
	return vm, nil
}

func (s *VMService) UserAction(ctx context.Context, vmID string, ownerID string, action string) (model.Instance, error) {
	vm, err := s.GetByIDForOwner(vmID, ownerID)
	if err != nil {
		return model.Instance{}, err
	}
	return s.applyAction(ctx, vm, ownerID, action)
}

func (s *VMService) AdminAction(ctx context.Context, vmID string, actorUserID string, action string) (model.Instance, error) {
	vm, err := s.GetByID(vmID)
	if err != nil {
		return model.Instance{}, err
	}
	return s.applyAction(ctx, vm, actorUserID, action)
}

func (s *VMService) applyAction(ctx context.Context, vm model.Instance, actorUserID string, action string) (model.Instance, error) {
	host, err := s.hostService.GetByID(vm.HostID.String())
	if err != nil {
		return model.Instance{}, err
	}

	switch action {
	case "start":
		err = s.incus.SetInstanceState(ctx, host, vm.IncusInstance, "start", false)
		vm.Status = model.InstanceStatusRunning
	case "stop":
		err = s.incus.SetInstanceState(ctx, host, vm.IncusInstance, "stop", true)
		vm.Status = model.InstanceStatusStopped
	case "reboot":
		err = s.incus.SetInstanceState(ctx, host, vm.IncusInstance, "restart", true)
		vm.Status = model.InstanceStatusRunning
	default:
		return model.Instance{}, fmt.Errorf("unsupported action: %s", action)
	}
	if err != nil {
		return model.Instance{}, err
	}

	if err := s.db.Save(&vm).Error; err != nil {
		return model.Instance{}, fmt.Errorf("save vm action status: %w", err)
	}
	if err := s.logOperation(actorUserID, &vm.ID, action, action+" vm "+vm.Name); err != nil {
		return model.Instance{}, err
	}
	return vm, nil
}

func (s *VMService) ListImages(ctx context.Context) ([]incus.ImageInfo, error) {
	return s.incus.ListImages(ctx)
}

func (s *VMService) GetResourceByID(ctx context.Context, vmID string) (dto.VMResourcePayload, error) {
	vm, err := s.GetByID(vmID)
	if err != nil {
		return dto.VMResourcePayload{}, err
	}
	return s.getResource(ctx, vm)
}

func (s *VMService) GetResourceForOwner(ctx context.Context, vmID string, ownerID string) (dto.VMResourcePayload, error) {
	vm, err := s.GetByIDForOwner(vmID, ownerID)
	if err != nil {
		return dto.VMResourcePayload{}, err
	}
	return s.getResource(ctx, vm)
}

func (s *VMService) getResource(ctx context.Context, vm model.Instance) (dto.VMResourcePayload, error) {
	host, err := s.hostService.GetByID(vm.HostID.String())
	if err != nil {
		return dto.VMResourcePayload{}, err
	}
	metrics, err := s.incus.GetInstanceMetrics(ctx, host, vm.IncusInstance)
	if err != nil {
		return dto.VMResourcePayload{}, err
	}
	return dto.VMResourcePayload{
		CPUNanoseconds: metrics.CPUNanoseconds,
		MemoryBytes:    metrics.MemoryBytes,
	}, nil
}

func ToInstancePayload(vm model.Instance) dto.InstancePayload {
	var ownerID *string
	if vm.OwnerID != nil {
		id := vm.OwnerID.String()
		ownerID = &id
	}
	return dto.InstancePayload{
		ID:             vm.ID.String(),
		IncusInstance:  vm.IncusInstance,
		HostID:         vm.HostID.String(),
		OwnerID:        ownerID,
		Name:           vm.Name,
		Image:          vm.Image,
		CPUCores:       vm.CPUCores,
		MemoryBytes:    vm.MemoryBytes,
		DiskRootBytes:  vm.DiskRootBytes,
		NetworkIngress: vm.NetworkIngress,
		NetworkEgress:  vm.NetworkEgress,
		Status:         vm.Status,
		VNCEnabled:     vm.VNCEnabled,
		CreatedAt:      vm.CreatedAt,
		UpdatedAt:      vm.UpdatedAt,
	}
}

func (s *VMService) logOperation(actorUserID string, instanceID *uuid.UUID, action, detail string) error {
	actorUUID, err := uuid.Parse(actorUserID)
	if err != nil {
		return fmt.Errorf("invalid actor user id: %w", err)
	}
	entry := model.OperationLog{
		UserID:     actorUUID,
		InstanceID: instanceID,
		Action:     action,
		Detail:     detail,
	}
	if err := s.db.Create(&entry).Error; err != nil {
		return fmt.Errorf("write operation log: %w", err)
	}
	return nil
}
