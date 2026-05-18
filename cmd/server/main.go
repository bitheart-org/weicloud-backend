package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/config"
	"weicloud-backend/internal/database"
	"weicloud-backend/internal/handler"
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

	authHandler := handler.NewAuthHandler(authService, userService)
	adminUserHandler := handler.NewAdminUserHandler(userService)

	router := gin.Default()
	handler.RegisterRoutes(router, handler.RouterDeps{
		AuthHandler:      authHandler,
		AdminUserHandler: adminUserHandler,
		AuthMiddleware:   middleware.JWTAuth(authService),
	})

	if err := router.Run(cfg.ServerAddr); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
