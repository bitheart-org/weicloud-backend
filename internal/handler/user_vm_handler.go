package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/middleware"
	"weicloud-backend/internal/service"
)

type UserVMHandler struct {
	vmService *service.VMService
}

func NewUserVMHandler(vmService *service.VMService) *UserVMHandler {
	return &UserVMHandler{vmService: vmService}
}

func (h *UserVMHandler) List(c *gin.Context) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40130, "unauthorized")
		return
	}

	items, err := h.vmService.ListByOwner(claims.UserID)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50050, "query vms failed")
		return
	}
	result := make([]dto.InstancePayload, 0, len(items))
	for _, vm := range items {
		result = append(result, service.ToInstancePayload(vm))
	}
	ok(c, gin.H{"items": result})
}

func (h *UserVMHandler) Detail(c *gin.Context) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40131, "unauthorized")
		return
	}
	vm, err := h.vmService.GetByIDForOwner(c.Param("id"), claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40450, "vm not found")
			return
		}
		fail(c, http.StatusInternalServerError, 50051, "query vm failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *UserVMHandler) Action(c *gin.Context, action string) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40132, "unauthorized")
		return
	}
	vm, err := h.vmService.UserAction(c.Request.Context(), c.Param("id"), claims.UserID, action)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40451, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, 40060, action+" vm failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *UserVMHandler) Start(c *gin.Context)  { h.Action(c, "start") }
func (h *UserVMHandler) Stop(c *gin.Context)   { h.Action(c, "stop") }
func (h *UserVMHandler) Reboot(c *gin.Context) { h.Action(c, "reboot") }

func (h *UserVMHandler) Resource(c *gin.Context) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40133, "unauthorized")
		return
	}
	payload, err := h.vmService.GetResourceForOwner(c.Request.Context(), c.Param("id"), claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40452, "vm not found")
			return
		}
		fail(c, http.StatusBadGateway, 50250, "query vm resource failed")
		return
	}
	ok(c, payload)
}

func (h *UserVMHandler) ResetPassword(c *gin.Context) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40134, "unauthorized")
		return
	}
	password, err := h.vmService.ResetRootPasswordForOwner(c.Request.Context(), c.Param("id"), claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40453, "vm not found")
			return
		}
		fail(c, http.StatusBadGateway, 50251, "reset vm password failed")
		return
	}
	ok(c, dto.ResetRootPasswordResponse{NewPassword: password})
}
