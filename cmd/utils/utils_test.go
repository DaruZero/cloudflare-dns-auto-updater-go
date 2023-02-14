package utils

import (
	"os"
	"testing"
)

const testString = "test"

// TestGetEnvRequired tests the GetEnv function with a required environment variable
func TestGetEnvRequired(t *testing.T) {
	// Set the environment variable
	err := os.Setenv("TEST_ENV", testString)
	if err != nil {
		t.Errorf("Error setting environment variable: %s", err)
	}

	// Get the environment variable
	value := GetEnv("TEST_ENV", true, "")

	// Check the value
	if value != testString {
		t.Errorf("GetEnv() = %s; want test", value)
	}
}

// TestGetEnvNotRequired tests the GetEnv function with a not required environment variable
func TestGetEnvNotRequired(t *testing.T) {
	// Set the environment variable
	err := os.Setenv("TEST_ENV", testString)
	if err != nil {
		t.Errorf("Error setting environment variable: %s", err)
	}

	// Get the environment variable
	value := GetEnv("TEST_ENV", false, "")

	// Check the value
	if value != testString {
		t.Errorf("GetEnv() = %s; want test", value)
	}
}

// TestGetEnvNotRequired tests the GetEnv function with a not required environment variable
func TestGetEnvNotRequiredFallback(t *testing.T) {
	// Get the environment variable
	value := GetEnv("TEST_ENV", false, testString)

	// Check the value
	if value != testString {
		t.Errorf("GetEnv() = %s; want test", value)
	}
}

// TestGetEnvAsIntRequired tests the GetEnvAsInt function with a required environment variable
func TestGetEnvAsIntRequired(t *testing.T) {
	// Set the environment variable
	err := os.Setenv("TEST_ENV", "1")
	if err != nil {
		t.Errorf("Error setting environment variable: %s", err)
	}

	// Get the environment variable
	value := GetEnvAsInt("TEST_ENV", true, 0)

	// Check the value
	if value != 1 {
		t.Errorf("GetEnvAsInt() = %d; want 1", value)
	}
}

// TestGetEnvAsIntNotRequired tests the GetEnvAsInt function with a not required environment variable
func TestGetEnvAsIntNotRequired(t *testing.T) {
	// Set the environment variable
	err := os.Setenv("TEST_ENV", "1")
	if err != nil {
		t.Errorf("Error setting environment variable: %s", err)
	}

	// Get the environment variable
	value := GetEnvAsInt("TEST_ENV", false, 0)

	// Check the value
	if value != 1 {
		t.Errorf("GetEnvAsInt() = %d; want 1", value)
	}
}

// TestGetEnvAsIntNotRequired tests the GetEnvAsInt function with a not required environment variable
func TestGetEnvAsIntNotRequiredFallback(t *testing.T) {
	// Get the environment variable
	value := GetEnvAsInt("TEST_ENV", false, 1)

	// Check the value
	if value != 1 {
		t.Errorf("GetEnvAsInt() = %d; want 1", value)
	}
}
