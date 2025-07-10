package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Get retrieves the value of the environment variable named by the key.
// If the variable is present in the environment, the value (which may be empty) is returned.
// Otherwise, the defaultVal is returned.
func Get(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// GetHostName retrieves the host name of the machine.
// If an error occurs during retrieval, "localhost" is returned.
func GetHostName() string {
	nodeName, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return nodeName
}

// GetInt retrieves the value of the environment variable named by the key and converts it to an integer.
// If the variable is not present in the environment or the value cannot be converted to an integer,
// the defaultVal is returned.
func GetInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if iVal, err := strconv.Atoi(value); err == nil {
			return iVal
		}
	}
	return defaultVal
}

// GetBool retrieves the value of the environment variable named by the key and converts it to a boolean.
// If the value is "1" or "true" (case insensitive), it returns true. Otherwise, it returns false.
// If the variable is not present in the environment, the defaultVal is returned.
func GetBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if value == "1" || strings.ToLower(value) == "true" {
			return true
		}
		return false
	}
	return defaultVal
}

// GetSlice retrieves the value of the environment variable named by the key, splits it using the specified separator,
// and returns it as a slice of strings. If the variable is not present in the environment, the defaultVal is returned.
func GetSlice(name string, defaultVal []string, sep string) []string {
	valStr := Get(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}

// GetMust retrieves the value of the environment variable named by the key.
// If the variable is not present in the environment, it panics with a custom error.
func GetMust(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Errorf("mandatory environment variable is not set: %v", key))
	}
	return value
}
