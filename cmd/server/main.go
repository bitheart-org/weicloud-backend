package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/config"
	"weicloud-backend/internal/database"
	"weicloud-backend/internal/handler"
	"weicloud-backend/internal/incus"
	"weicloud-backend/internal/middleware"
	"weicloud-backend/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Printf("database connect failed: %v", err)
		os.Exit(1)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Printf("database migration failed: %v", err)
		os.Exit(1)
	}

	if err := database.SeedAdmin(db, cfg); err != nil {
		log.Printf("seed admin user failed: %v", err)
		os.Exit(1)
	}

	authService := service.NewAuthService(db, cfg.JWTSecret, cfg.JWTExpireHours)
	userService := service.NewUserService(db)
	incusManager := incus.NewManager()
	hostService := service.NewHostService(db, incusManager)
	vmService := service.NewVMService(db, hostService, incusManager)
	systemService := service.NewSystemService(db)
	vncService := service.NewVNCService()

	if err := hostService.LoadRegisteredHosts(); err != nil {
		log.Printf("load registered hosts failed: %v", err)
	}

	authHandler := handler.NewAuthHandler(authService, userService)
	adminUserHandler := handler.NewAdminUserHandler(userService)
	adminHostHandler := handler.NewAdminHostHandler(hostService)
	adminVMHandler := handler.NewAdminVMHandler(vmService)
	adminSystemHandler := handler.NewAdminSystemHandler(systemService, vmService)
	userVMHandler := handler.NewUserVMHandler(vmService)
	userVNCHandler := handler.NewUserVNCHandler(vmService, vncService)
	vncWSHandler := handler.NewVNCWSHandler(vmService, vncService)

	router := gin.Default()
	router.Use(middleware.SecurityHeaders(), middleware.RateLimit(300, time.Minute))
	handler.RegisterRoutes(router, handler.RouterDeps{
		AuthHandler:        authHandler,
		AdminUserHandler:   adminUserHandler,
		AdminHostHandler:   adminHostHandler,
		AdminVMHandler:     adminVMHandler,
		AdminSystemHandler: adminSystemHandler,
		UserVMHandler:      userVMHandler,
		UserVNCHandler:     userVNCHandler,
		VNCWSHandler:       vncWSHandler,
		AuthMiddleware:     middleware.JWTAuth(authService),
	})

	if err := router.Run(cfg.ServerAddr); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
