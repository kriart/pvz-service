package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	JWT    JWTConfig
}

type ServerConfig struct {
	HTTPPort    int
	GRPCPort    int
	MetricsPort int
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

type JWTConfig struct {
	Secret string
}

func Load() Config {
	cfg := Config{
		Server: ServerConfig{
			HTTPPort:    8080,
			GRPCPort:    3000,
			MetricsPort: 9090,
		},
		DB: DBConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Name:     "pvz",
		},
		JWT: JWTConfig{
			Secret: "secret",
		},
	}
	if portStr := os.Getenv("HTTP_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.HTTPPort = p
		}
	}
	if portStr := os.Getenv("GRPC_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.GRPCPort = p
		}
	}
	if portStr := os.Getenv("METRICS_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.MetricsPort = p
		}
	}
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.DB.Host = host
	}
	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.DB.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.DB.User = user
	}
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		cfg.DB.Password = pass
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.DB.Name = name
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWT.Secret = secret
	}
	return cfg
}
