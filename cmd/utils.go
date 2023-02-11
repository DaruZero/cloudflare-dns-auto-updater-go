package main

import (
	"go.uber.org/zap"
	"log"
	"os"
	"strconv"
	"strings"
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

	i, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("Environment variable %s must be an integer", name)
	}

	return i
}

func SanitizeString(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	return s
}
