package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/dto"
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
		fail(c, http.StatusInternalServerError, 50020, "query users failed")
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
		fail(c, http.StatusBadRequest, 40020, "invalid request")
		return
	}

	user, err := h.userService.Create(req)
	if err != nil {
		fail(c, http.StatusBadRequest, 40021, "create user failed")
		return
	}

	ok(c, service.ToUserPayload(user))
}

func (h *AdminUserHandler) Update(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40022, "invalid request")
		return
	}

	userID := c.Param("id")
	user, err := h.userService.Update(userID, req)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40420, "user not found")
			return
		}
		fail(c, http.StatusBadRequest, 40023, "update user failed")
		return
	}

	ok(c, service.ToUserPayload(user))
}

func (h *AdminUserHandler) Disable(c *gin.Context) {
	userID := c.Param("id")
	if err := h.userService.Disable(userID); err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40421, "user not found")
			return
		}
		fail(c, http.StatusInternalServerError, 50021, "disable user failed")
		return
	}

	ok(c, gin.H{"id": userID})
}

func (h *AdminUserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40024, "invalid request")
		return
	}

	userID := c.Param("id")
	if err := h.userService.ResetPassword(userID, req.Password); err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40422, "user not found")
			return
		}
		fail(c, http.StatusInternalServerError, 50022, "reset password failed")
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
