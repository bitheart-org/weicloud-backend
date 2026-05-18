package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/errcode"
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
		fail(c, http.StatusUnauthorized, errcode.UserVMListUnauthorized, "unauthorized")
		return
	}

	items, err := h.vmService.ListByOwner(claims.UserID)
	if err != nil {
		fail(c, http.StatusInternalServerError, errcode.UserVMListFailed, "query vms failed")
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
		fail(c, http.StatusUnauthorized, errcode.UserVMDetailUnauthorized, "unauthorized")
		return
	}
	vm, err := h.vmService.GetByIDForOwner(c.Param("id"), claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.UserVMNotFoundForDetail, "vm not found")
			return
		}
		fail(c, http.StatusInternalServerError, errcode.UserVMDetailFailed, "query vm failed")
		return
	}
	ok(c, service.ToInstancePayload(vm))
}

func (h *UserVMHandler) Action(c *gin.Context, action string) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, errcode.UserVMActionUnauthorized, "unauthorized")
		return
	}
	vm, err := h.vmService.UserAction(c.Request.Context(), c.Param("id"), claims.UserID, action)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.UserVMNotFoundForAction, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, errcode.UserVMActionFailed, action+" vm failed")
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
		fail(c, http.StatusUnauthorized, errcode.UserVMResourceUnauthorized, "unauthorized")
		return
	}
	payload, err := h.vmService.GetResourceForOwner(c.Request.Context(), c.Param("id"), claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.UserVMNotFoundForResource, "vm not found")
			return
		}
		fail(c, http.StatusBadGateway, errcode.UserVMResourceFailed, "query vm resource failed")
		return
	}
	ok(c, payload)
}

func (h *UserVMHandler) ResetPassword(c *gin.Context) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, errcode.UserVMResetPasswordUnauthorized, "unauthorized")
		return
	}
	password, err := h.vmService.ResetRootPasswordForOwner(c.Request.Context(), c.Param("id"), claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.UserVMNotFoundForResetPassword, "vm not found")
			return
		}
		fail(c, http.StatusBadGateway, errcode.UserVMResetPasswordFailed, "reset vm password failed")
		return
	}
	ok(c, dto.ResetRootPasswordResponse{NewPassword: password})
}

func (h *UserVMHandler) UpdateLoginPassword(c *gin.Context) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, errcode.UserVMResetPasswordUnauthorized, "unauthorized")
		return
	}
	var req dto.UpdateLoginPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, errcode.UserVMUpdatePasswordInvalidInput, "invalid request")
		return
	}
	username, err := h.vmService.UpdateLoginPasswordForOwner(c.Request.Context(), c.Param("id"), claims.UserID, req.Password)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.UserVMNotFoundForResetPassword, "vm not found")
			return
		}
		fail(c, http.StatusBadGateway, errcode.UserVMUpdatePasswordFailed, "update vm password failed")
		return
	}
	ok(c, dto.UpdateLoginPasswordResponse{
		Username:    username,
		NewPassword: req.Password,
	})
}
