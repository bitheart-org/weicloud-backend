package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerAddr         string
	DBHost             string
	DBPort             int
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	JWTSecret          string
	JWTExpireHours     int
	AdminUsername      string
	AdminPassword      string
	AdminDisplayName   string
	VMDefaultUsername  string
	FRPSServerAddr     string
	FRPSServerPort     int
	FRPSToken          string
	FRPRemotePortStart int
}

func Load() Config {
	return Config{
		ServerAddr:         getEnv("SERVER_ADDR", ":8080"),
		DBHost:             getEnv("DB_HOST", "127.0.0.1"),
		DBPort:             getEnvAsInt("DB_PORT", 5432),
		DBUser:             getEnv("DB_USER", "weicloud"),
		DBPassword:         getEnv("DB_PASSWORD", "weicloud"),
		DBName:             getEnv("DB_NAME", "weicloud"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		JWTSecret:          getEnv("JWT_SECRET", "change-me"),
		JWTExpireHours:     getEnvAsInt("JWT_EXPIRE_HOURS", 24),
		AdminUsername:      getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:      getEnv("ADMIN_PASSWORD", "Admin@123456"),
		AdminDisplayName:   getEnv("ADMIN_DISPLAY_NAME", "Administrator"),
		VMDefaultUsername:  getEnv("VM_DEFAULT_USERNAME", "weicloud"),
		FRPSServerAddr:     getEnv("FRPS_SERVER_ADDR", ""),
		FRPSServerPort:     getEnvAsInt("FRPS_SERVER_PORT", 30000),
		FRPSToken:          getEnv("FRPS_TOKEN", ""),
		FRPRemotePortStart: getEnvAsInt("FRP_REMOTE_PORT_START", 30900),
	}
}

func (c Config) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	return value
}
