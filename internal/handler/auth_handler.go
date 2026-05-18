package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/errcode"
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
		fail(c, http.StatusBadRequest, errcode.AuthInvalidRequest, "invalid request")
		return
	}

	user, token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredential):
			fail(c, http.StatusUnauthorized, errcode.AuthInvalidCredential, "invalid username or password")
		case errors.Is(err, service.ErrUserDisabled):
			fail(c, http.StatusForbidden, errcode.AuthUserDisabled, "user is disabled")
		default:
			fail(c, http.StatusInternalServerError, errcode.AuthLoginFailed, "login failed")
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
		fail(c, http.StatusUnauthorized, errcode.AuthUnauthorized, "unauthorized")
		return
	}

	userClaims, castOK := claims.(*service.UserClaims)
	if !castOK {
		fail(c, http.StatusUnauthorized, errcode.AuthUnauthorizedContext, "unauthorized")
		return
	}

	user, err := h.userService.GetByID(userClaims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusUnauthorized, errcode.AuthUserNotFound, "user not found")
			return
		}
		fail(c, http.StatusInternalServerError, errcode.AuthQueryUserFailed, "query user failed")
		return
	}

	ok(c, service.ToUserPayload(user))
}
