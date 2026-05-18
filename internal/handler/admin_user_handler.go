package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/errcode"
	"weicloud-backend/internal/service"
)

type AdminUserHandler struct {
	userService *service.UserService
}

func NewAdminUserHandler(userService *service.UserService) *AdminUserHandler {
	return &AdminUserHandler{
		userService: userService,
	}
}

func (h *AdminUserHandler) List(c *gin.Context) {
	page := parsePositiveInt(c.Query("page"), 1)
	pageSize := parsePositiveInt(c.Query("page_size"), 20)
	if pageSize > 100 {
		pageSize = 100
	}

	query := c.Query("query")
	users, total, err := h.userService.List(page, pageSize, query)
	if err != nil {
		fail(c, http.StatusInternalServerError, errcode.AdminUserQueryFailed, "query users failed")
		return
	}

	items := make([]dto.UserPayload, 0, len(users))
	for _, user := range users {
		items = append(items, service.ToUserPayload(user))
	}

	ok(c, dto.ListUsersResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *AdminUserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminUserCreateInvalidInput, "invalid request")
		return
	}

	user, err := h.userService.Create(req)
	if err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminUserCreateFailed, "create user failed")
		return
	}

	ok(c, service.ToUserPayload(user))
}

func (h *AdminUserHandler) Update(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminUserUpdateInvalidInput, "invalid request")
		return
	}

	userID := c.Param("id")
	user, err := h.userService.Update(userID, req)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminUserNotFoundForUpdate, "user not found")
			return
		}
		fail(c, http.StatusBadRequest, errcode.AdminUserUpdateFailed, "update user failed")
		return
	}

	ok(c, service.ToUserPayload(user))
}

func (h *AdminUserHandler) Disable(c *gin.Context) {
	userID := c.Param("id")
	if err := h.userService.Disable(userID); err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminUserNotFoundForDisable, "user not found")
			return
		}
		fail(c, http.StatusInternalServerError, errcode.AdminUserDisableFailed, "disable user failed")
		return
	}

	ok(c, gin.H{"id": userID})
}

func (h *AdminUserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminUserResetInvalidInput, "invalid request")
		return
	}

	userID := c.Param("id")
	if err := h.userService.ResetPassword(userID, req.Password); err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminUserNotFoundForReset, "user not found")
			return
		}
		fail(c, http.StatusInternalServerError, errcode.AdminUserResetPasswordFailed, "reset password failed")
		return
	}

	ok(c, gin.H{"id": userID})
}

func parsePositiveInt(raw string, defaultValue int) int {
	if raw == "" {
		return defaultValue
	}

	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return defaultValue
	}
	return v
}
