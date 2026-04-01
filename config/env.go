package config

import (
	"os"
	"strconv"
)

// Environment represents the application environment
type Environment string

const (
	Local      Environment = "local"
	Staging    Environment = "staging"
	Production Environment = "production"
)

func (e Environment) String() string {
	return string(e)
}

// GetEnv retrieves an environment variable with a fallback value
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// GetEnvBool retrieves a boolean environment variable
func GetEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return fallback
}

// GetEnvInt retrieves an integer environment variable
func GetEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
