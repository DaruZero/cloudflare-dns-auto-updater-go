package env

import (
	"os"
	"os/exec"
	"testing"
)

const (
	testKey    = "TEST_ENV"
	testString = "test"
	testSlice  = "test1,test2,test3"
)

func setup(value string) {
	err := os.Setenv("TEST_ENV", value)
	if err != nil {
		panic(err)
	}
}

func teardown() {
	err := os.Unsetenv("TEST_ENV")
	if err != nil {
		panic(err)
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback string
		expected string
		required bool
		wantExit bool
	}{
		{
			name:     "Required",
			key:      testKey,
			value:    testString,
			required: true,
			fallback: "",
			expected: testString,
			wantExit: false,
		},
		{
			name:     "RequiredMissing",
			key:      testKey,
			value:    "",
			required: true,
			fallback: "",
			expected: "",
			wantExit: true,
		},
		{
			name:     "NotRequired",
			key:      testKey,
			value:    "",
			required: false,
			fallback: "",
			expected: "",
			wantExit: false,
		},
		{
			name:     "NotRequiredWithFallback",
			key:      testKey,
			value:    "",
			required: false,
			fallback: testString,
			expected: testString,
			wantExit: false,
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		setup(tt.value)
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantExit {
				if os.Getenv("BE_CRASHER") == "1" {
					GetEnv(tt.key, tt.required, tt.fallback)
					return
				}
				cmd := exec.Command(os.Args[0], "-test.run=TestGetEnv")
				cmd.Env = append(os.Environ(), "BE_CRASHER=1")
				err := cmd.Run()
				if e, ok := err.(*exec.ExitError); ok && !e.Success() {
					return
				}
				t.Fatalf("process ran with err %v, want exit status 1", err)
			}

			value := GetEnv(tt.key, tt.required, tt.fallback)

			if value != tt.expected {
				t.Errorf("Function() = %s; want %s", value, tt.expected)
			}
		})
		teardown()
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback int
		expected int
		required bool
		wantExit bool
	}{
		{
			name:     "Required",
			key:      testKey,
			value:    "1",
			required: true,
			fallback: 0,
			expected: 1,
			wantExit: false,
		},
		{
			name:     "RequiredMissing",
			key:      testKey,
			value:    "",
			required: true,
			fallback: 0,
			expected: 1,
			wantExit: true,
		},
		{
			name:     "NotRequired",
			key:      testKey,
			value:    "1",
			required: false,
			fallback: 0,
			expected: 1,
			wantExit: false,
		},
		{
			name:     "NotRequiredWithFallback",
			key:      testKey,
			value:    "",
			required: false,
			fallback: 1,
			expected: 1,
			wantExit: false,
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		setup(tt.value)
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantExit {
				if os.Getenv("BE_CRASHER") == "1" {
					GetEnvAsInt(tt.key, tt.required, tt.fallback)
					return
				}
				cmd := exec.Command(os.Args[0], "-test.run=TestGetEnv")
				cmd.Env = append(os.Environ(), "BE_CRASHER=1")
				err := cmd.Run()
				if e, ok := err.(*exec.ExitError); ok && !e.Success() {
					return
				}
				t.Fatalf("process ran with err %v, want exit status 1", err)
			}

			value := GetEnvAsInt(tt.key, tt.required, tt.fallback)

			if value != tt.expected {
				t.Errorf("Function() = %d; want %d", value, tt.expected)
			}
		})
		teardown()
	}
}

func TestGetEnvAsStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback []string
		expected []string
		required bool
		wantExit bool
	}{
		{
			name:     "Required",
			key:      testKey,
			value:    testSlice,
			required: true,
			fallback: []string{},
			expected: []string{"test1", "test2", "test3"},
			wantExit: false,
		},
		{
			name:     "RequiredMissing",
			key:      testKey,
			value:    "",
			required: true,
			fallback: []string{},
			expected: []string{},
			wantExit: true,
		},
		{
			name:     "NotRequired",
			key:      testKey,
			value:    testSlice,
			required: false,
			fallback: []string{},
			expected: []string{"test1", "test2", "test3"},
			wantExit: false,
		},
		{
			name:     "NotRequiredWithFallback",
			key:      testKey,
			value:    "",
			required: false,
			fallback: []string{"test1", "test2", "test3"},
			expected: []string{"test1", "test2", "test3"},
			wantExit: false,
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		setup(tt.value)
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantExit {
				if os.Getenv("BE_CRASHER") == "1" {
					GetEnvAsStringSlice(tt.key, tt.required, tt.fallback)
					return
				}
				cmd := exec.Command(os.Args[0], "-test.run=TestGetEnv")
				cmd.Env = append(os.Environ(), "BE_CRASHER=1")
				err := cmd.Run()
				if e, ok := err.(*exec.ExitError); ok && !e.Success() {
					return
				}
				t.Fatalf("process ran with err %v, want exit status 1", err)
			}

			value := GetEnvAsStringSlice(tt.key, tt.required, tt.fallback)

			if len(value) != len(tt.expected) {
				t.Errorf("Function() = %v; want %v", value, tt.expected)
			}
		})
		teardown()
	}
}
