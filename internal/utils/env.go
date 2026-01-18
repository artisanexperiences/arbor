package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func ReadEnvFile(worktreePath, filename string) map[string]string {
	result := make(map[string]string)

	envPath := filepath.Join(worktreePath, filename)
	data, err := os.ReadFile(envPath)
	if err != nil {
		return result
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}

	return result
}

func EnvExists(env map[string]string, key string) bool {
	_, exists := env[key]
	return exists
}

func EnvNotExists(env map[string]string, key string) bool {
	return !EnvExists(env, key)
}
