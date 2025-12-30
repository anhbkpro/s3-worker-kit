package config

import (
	"os"
	"strconv"
)

type Config struct {
	AWSRegion   string
	AWSEndpoint string
	PoolSize    int
}

func Load() Config {
	poolSize, _ := strconv.Atoi(getEnv("WORKER_POOL_SIZE", "5"))

	return Config{
		AWSRegion:   getEnv("AWS_REGION", "us-east-1"),
		AWSEndpoint: os.Getenv("AWS_ENDPOINT"),
		PoolSize:    poolSize,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
