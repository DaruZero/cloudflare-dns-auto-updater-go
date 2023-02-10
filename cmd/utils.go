package main

import (
	"go.uber.org/zap"
	"log"
	"os"
	"strconv"
)

// GetEnv returns the value of the environment variable
func GetEnv(name string, required bool, fallback string) string {
	zap.S().Debugf("Loading environment variable %s", name)
	value := os.Getenv(name)

	if required && value == "" {
		zap.S().Fatalf("Environment variable %s is required", name)
	}

	if value == "" {
		value = fallback
	}

	zap.S().Debugf("Environment variable %s loaded: %s", name, value)

	return value
}

// GetEnvAsInt returns the value of the environment variable as an integer
func GetEnvAsInt(name string, required bool, fallback int) int {
	zap.S().Debugf("Loading environment variable %s", name)
	value := os.Getenv(name)

	if required && value == "" {
		log.Fatalf("Environment variable %s is required", name)
	}

	if value == "" {
		return fallback
	}

	zap.S().Debugf("Environment variable %s loaded: %s", name, value)

	i, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("Environment variable %s must be an integer", name)
	}

	return i
}
