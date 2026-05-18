package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/errcode"
)

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		claims, ok := CurrentUserClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    errcode.AuthMissingUserContext,
				"message": "missing user context",
				"data":    nil,
			})
			return
		}

		if _, exists := allowed[claims.Role]; !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    errcode.AuthPermissionDenied,
				"message": "permission denied",
				"data":    nil,
			})
			return
		}

		c.Next()
	}
}
