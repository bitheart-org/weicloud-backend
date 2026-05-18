package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	userService *service.UserService
}

func NewAuthHandler(authService *service.AuthService, userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}

	user, token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredential):
			fail(c, http.StatusUnauthorized, 40110, "invalid username or password")
		case errors.Is(err, service.ErrUserDisabled):
			fail(c, http.StatusForbidden, 40310, "user is disabled")
		default:
			fail(c, http.StatusInternalServerError, 50010, "login failed")
		}
		return
	}

	ok(c, dto.LoginResponse{
		Token: token,
		User:  service.ToUserPayload(user),
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	claims, exists := c.Get("current_user_claims")
	if !exists {
		fail(c, http.StatusUnauthorized, 40111, "unauthorized")
		return
	}

	userClaims, castOK := claims.(*service.UserClaims)
	if !castOK {
		fail(c, http.StatusUnauthorized, 40112, "unauthorized")
		return
	}

	user, err := h.userService.GetByID(userClaims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusUnauthorized, 40113, "user not found")
			return
		}
		fail(c, http.StatusInternalServerError, 50011, "query user failed")
		return
	}

	ok(c, service.ToUserPayload(user))
}
