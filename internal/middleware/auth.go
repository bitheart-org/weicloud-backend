package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/errcode"
	"weicloud-backend/internal/service"
)

const currentUserClaimsKey = "current_user_claims"

func JWTAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    errcode.AuthMissingAuthorizationHeader,
				"message": "missing authorization header",
				"data":    nil,
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    errcode.AuthInvalidAuthorizationHeader,
				"message": "invalid authorization header",
				"data":    nil,
			})
			return
		}

		claims, err := authService.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    errcode.AuthInvalidToken,
				"message": "invalid token",
				"data":    nil,
			})
			return
		}

		c.Set(currentUserClaimsKey, claims)
		c.Next()
	}
}

func CurrentUserClaims(c *gin.Context) (*service.UserClaims, bool) {
	v, ok := c.Get(currentUserClaimsKey)
	if !ok {
		return nil, false
	}

	claims, ok := v.(*service.UserClaims)
	return claims, ok
}
