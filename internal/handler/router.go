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
	AdminVMHandler   *AdminVMHandler
	UserVMHandler    *UserVMHandler
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

		adminGroup.GET("/vms", deps.AdminVMHandler.List)
		adminGroup.POST("/vms", deps.AdminVMHandler.Create)
		adminGroup.DELETE("/vms/:id", deps.AdminVMHandler.Delete)
		adminGroup.PUT("/vms/:id/config", deps.AdminVMHandler.UpdateConfig)
		adminGroup.POST("/vms/:id/disk", deps.AdminVMHandler.ResizeDisk)
		adminGroup.PUT("/vms/:id/network", deps.AdminVMHandler.UpdateNetwork)
		adminGroup.PUT("/vms/:id/assign", deps.AdminVMHandler.Assign)
		adminGroup.POST("/vms/:id/migrate", deps.AdminVMHandler.Migrate)
		adminGroup.POST("/vms/:id/start", deps.AdminVMHandler.Start)
		adminGroup.POST("/vms/:id/stop", deps.AdminVMHandler.Stop)
		adminGroup.POST("/vms/:id/reboot", deps.AdminVMHandler.Reboot)
		adminGroup.GET("/images", deps.AdminVMHandler.ListImages)
	}

	userGroup := api.Group("/user")
	userGroup.Use(deps.AuthMiddleware, middleware.RequireRoles(model.UserRoleUser, model.UserRoleAdmin))
	{
		userGroup.GET("/vms", deps.UserVMHandler.List)
		userGroup.GET("/vms/:id", deps.UserVMHandler.Detail)
		userGroup.POST("/vms/:id/start", deps.UserVMHandler.Start)
		userGroup.POST("/vms/:id/stop", deps.UserVMHandler.Stop)
		userGroup.POST("/vms/:id/reboot", deps.UserVMHandler.Reboot)
	}
}
