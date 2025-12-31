package config

import (
	"os"
	"strconv"
)

type Config struct {
	AWSRegion   string
	AWSEndpoint string
	PoolSize    int
	Otel        OtelConfig
}

type OtelConfig struct {
	Enabled  bool   `json:"enabled"`
	Exporter string `json:"exporter"`
	Endpoint string `json:"endpoint"`
	Insecure bool   `json:"insecure"`
}

func Load() Config {
	poolSize, _ := strconv.Atoi(getEnv("WORKER_POOL_SIZE", "5"))

	return Config{
		AWSRegion:   getEnv("AWS_REGION", "us-east-1"),
		AWSEndpoint: os.Getenv("AWS_ENDPOINT"),
		PoolSize:    poolSize,
		Otel: OtelConfig{
			Enabled:  getEnv("OTEL_ENABLED", "true") == "true",
			Exporter: getEnv("OTEL_EXPORTER", "otlp"),
			Endpoint: getEnv("OTEL_ENDPOINT", "localhost:4318"),
			Insecure: getEnv("OTEL_INSECURE", "true") == "true",
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
