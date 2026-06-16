package config

import "os"

type AppConfig struct {
	RedisURL string
	NodeID   string
	Port     string
	GRPCPort string
}

func Load() AppConfig {
	return AppConfig{
		RedisURL: getEnv("REDIS_URL", "localhost:6379"),
		NodeID:   getEnv("NODE_ID", "node-1"),
		Port:     getEnv("PORT", "8080"),
		GRPCPort: getEnv("GRPC_PORT", "50051"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
