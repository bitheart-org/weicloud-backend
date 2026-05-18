package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/middleware"
	"weicloud-backend/internal/model"
)

type RouterDeps struct {
	AuthHandler      *AuthHandler
	AdminUserHandler *AdminUserHandler
	AdminHostHandler *AdminHostHandler
	AuthMiddleware   gin.HandlerFunc
}

func RegisterRoutes(r *gin.Engine, deps RouterDeps) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	api := r.Group("/api")

	authGroup := api.Group("/auth")
	authGroup.POST("/login", deps.AuthHandler.Login)
	authGroup.GET("/me", deps.AuthMiddleware, deps.AuthHandler.Me)

	adminGroup := api.Group("/admin")
	adminGroup.Use(deps.AuthMiddleware, middleware.RequireRoles(model.UserRoleAdmin))
	{
		adminGroup.GET("/users", deps.AdminUserHandler.List)
		adminGroup.POST("/users", deps.AdminUserHandler.Create)
		adminGroup.PUT("/users/:id", deps.AdminUserHandler.Update)
		adminGroup.DELETE("/users/:id", deps.AdminUserHandler.Disable)
		adminGroup.PUT("/users/:id/password", deps.AdminUserHandler.ResetPassword)

		adminGroup.GET("/hosts", deps.AdminHostHandler.List)
		adminGroup.POST("/hosts", deps.AdminHostHandler.Create)
		adminGroup.PUT("/hosts/:id", deps.AdminHostHandler.Update)
		adminGroup.DELETE("/hosts/:id", deps.AdminHostHandler.Delete)
		adminGroup.POST("/hosts/:id/sync", deps.AdminHostHandler.Sync)
	}
}
