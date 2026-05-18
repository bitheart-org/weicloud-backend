package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/errcode"
	"weicloud-backend/internal/middleware"
	"weicloud-backend/internal/service"
)

type AdminVMHandler struct {
	vmService *service.VMService
}

func NewAdminVMHandler(vmService *service.VMService) *AdminVMHandler {
	return &AdminVMHandler{vmService: vmService}
}

func (h *AdminVMHandler) List(c *gin.Context) {
	items, err := h.vmService.ListAll()
	if err != nil {
		fail(c, http.StatusInternalServerError, errcode.AdminVMQueryFailed, "query vms failed")
		return
	}

	result := make([]dto.InstancePayload, 0, len(items))
	for _, vm := range items {
		result = append(result, service.ToInstancePayload(vm))
	}
	ok(c, gin.H{"items": result})
}

func (h *AdminVMHandler) Create(c *gin.Context) {
	var req dto.CreateVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminVMCreateInvalidInput, "invalid request")
		return
	}

	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, errcode.AdminVMCreateUnauthorized, "unauthorized")
		return
	}

	vm, err := h.vmService.Create(c.Request.Context(), req, claims.UserID)
	if err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminVMCreateFailed, "create vm failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) Delete(c *gin.Context) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, errcode.AdminVMDeleteUnauthorized, "unauthorized")
		return
	}
	vmID := c.Param("id")
	if err := h.vmService.Delete(c.Request.Context(), vmID, claims.UserID); err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminVMNotFoundForDelete, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, errcode.AdminVMDeleteFailed, "delete vm failed")
		return
	}
	ok(c, gin.H{"id": vmID})
}

func (h *AdminVMHandler) UpdateConfig(c *gin.Context) {
	var req dto.UpdateVMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminVMUpdateConfigInvalidInput, "invalid request")
		return
	}
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, errcode.AdminVMUpdateConfigUnauthorized, "unauthorized")
		return
	}
	vm, err := h.vmService.UpdateConfig(c.Request.Context(), c.Param("id"), req, claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminVMNotFoundForUpdateConfig, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, errcode.AdminVMUpdateConfigFailed, "update vm config failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) ResizeDisk(c *gin.Context) {
	var req dto.ResizeVMDiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminVMResizeInvalidInput, "invalid request")
		return
	}
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, errcode.AdminVMResizeUnauthorized, "unauthorized")
		return
	}
	vm, err := h.vmService.ResizeDisk(c.Request.Context(), c.Param("id"), req, claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminVMNotFoundForResize, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, errcode.AdminVMResizeFailed, "resize vm disk failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) UpdateNetwork(c *gin.Context) {
	var req dto.UpdateVMNetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminVMUpdateNetworkInvalidInput, "invalid request")
		return
	}
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, errcode.AdminVMUpdateNetworkUnauthorized, "unauthorized")
		return
	}
	vm, err := h.vmService.UpdateNetwork(c.Request.Context(), c.Param("id"), req, claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminVMNotFoundForUpdateNetwork, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, errcode.AdminVMUpdateNetworkFailed, "update vm network failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) Assign(c *gin.Context) {
	var req dto.AssignVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminVMAssignInvalidInput, "invalid request")
		return
	}
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, errcode.AdminVMAssignUnauthorized, "unauthorized")
		return
	}
	vm, err := h.vmService.AssignOwner(c.Param("id"), req.OwnerID, claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminVMNotFoundForAssign, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, errcode.AdminVMAssignFailed, "assign vm failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) Migrate(c *gin.Context) {
	var req dto.MigrateVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminVMMigrateInvalidInput, "invalid request")
		return
	}
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, errcode.AdminVMMigrateUnauthorized, "unauthorized")
		return
	}
	vm, err := h.vmService.Migrate(c.Param("id"), req, claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminVMNotFoundForMigrate, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, errcode.AdminVMMigrateFailed, "migrate vm failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) Action(c *gin.Context, action string) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, errcode.AdminVMActionUnauthorized, "unauthorized")
		return
	}
	vm, err := h.vmService.AdminAction(c.Request.Context(), c.Param("id"), claims.UserID, action)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminVMNotFoundForAction, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, errcode.AdminVMActionFailed, action+" vm failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) Start(c *gin.Context)  { h.Action(c, "start") }
func (h *AdminVMHandler) Stop(c *gin.Context)   { h.Action(c, "stop") }
func (h *AdminVMHandler) Reboot(c *gin.Context) { h.Action(c, "reboot") }

func (h *AdminVMHandler) ListImages(c *gin.Context) {
	images, err := h.vmService.ListImages(c.Request.Context())
	if err != nil {
		fail(c, http.StatusBadGateway, errcode.AdminVMListImagesFailed, "list images failed")
		return
	}
	ok(c, gin.H{"items": images})
}
