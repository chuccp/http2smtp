package util

import (
	"os"
	"strings"
)

func GetEnvOrDefault(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	return value
}
func GetEnv(key string) string {
	return os.Getenv(key)
}
func GetEnvIntOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	v := StringToInt(value)
	if v == 0 {
		return defaultValue
	}
	return v
}
func GetEnvInt(key string) int {
	return StringToInt(os.Getenv(key))
}
func GetEnvBool(key string) bool {
	return StringToBool(os.Getenv(key))
}
func GetEnvBoolOrDefault(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return StringToBool(value)
}
