package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/dto"
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
		fail(c, http.StatusInternalServerError, 50040, "query vms failed")
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
		fail(c, http.StatusBadRequest, 40040, "invalid request")
		return
	}

	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40120, "unauthorized")
		return
	}

	vm, err := h.vmService.Create(c.Request.Context(), req, claims.UserID)
	if err != nil {
		fail(c, http.StatusBadRequest, 40041, "create vm failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) Delete(c *gin.Context) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40121, "unauthorized")
		return
	}
	vmID := c.Param("id")
	if err := h.vmService.Delete(c.Request.Context(), vmID, claims.UserID); err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40440, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, 40042, "delete vm failed")
		return
	}
	ok(c, gin.H{"id": vmID})
}

func (h *AdminVMHandler) UpdateConfig(c *gin.Context) {
	var req dto.UpdateVMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40043, "invalid request")
		return
	}
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40122, "unauthorized")
		return
	}
	vm, err := h.vmService.UpdateConfig(c.Request.Context(), c.Param("id"), req, claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40441, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, 40044, "update vm config failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) ResizeDisk(c *gin.Context) {
	var req dto.ResizeVMDiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40045, "invalid request")
		return
	}
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40123, "unauthorized")
		return
	}
	vm, err := h.vmService.ResizeDisk(c.Request.Context(), c.Param("id"), req, claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40442, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, 40046, "resize vm disk failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) UpdateNetwork(c *gin.Context) {
	var req dto.UpdateVMNetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40047, "invalid request")
		return
	}
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40124, "unauthorized")
		return
	}
	vm, err := h.vmService.UpdateNetwork(c.Request.Context(), c.Param("id"), req, claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40443, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, 40048, "update vm network failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) Assign(c *gin.Context) {
	var req dto.AssignVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40049, "invalid request")
		return
	}
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40125, "unauthorized")
		return
	}
	vm, err := h.vmService.AssignOwner(c.Param("id"), req.OwnerID, claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40444, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, 40050, "assign vm failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) Migrate(c *gin.Context) {
	var req dto.MigrateVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40051, "invalid request")
		return
	}
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40126, "unauthorized")
		return
	}
	vm, err := h.vmService.Migrate(c.Param("id"), req, claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40445, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, 40052, "migrate vm failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *AdminVMHandler) Action(c *gin.Context, action string) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40127, "unauthorized")
		return
	}
	vm, err := h.vmService.AdminAction(c.Request.Context(), c.Param("id"), claims.UserID, action)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40446, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, 40053, action+" vm failed")
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
		fail(c, http.StatusBadGateway, 50240, "list images failed")
		return
	}
	ok(c, gin.H{"items": images})
}
