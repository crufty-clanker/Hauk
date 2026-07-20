package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	// Verify defaults.
	if cfg.StorageBackend != "redis" {
		t.Fatalf("expected storage_backend 'redis', got %q", cfg.StorageBackend)
	}
	if cfg.AuthMethod != "password" {
		t.Fatalf("expected auth_method 'password', got %q", cfg.AuthMethod)
	}
	if cfg.MaxDuration != 86400 {
		t.Fatalf("expected max_duration 86400, got %d", cfg.MaxDuration)
	}
	if cfg.MinInterval != 1 {
		t.Fatalf("expected min_interval 1, got %d", cfg.MinInterval)
	}
	if cfg.LinkStyle != 0 {
		t.Fatalf("expected link_style 0, got %d", cfg.LinkStyle)
	}
	if cfg.ServerPort != 8080 {
		t.Fatalf("expected server_port 8080, got %d", cfg.ServerPort)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables.
	os.Setenv("HAUK_STORAGE_BACKEND", "redis")
	os.Setenv("HAUK_REDIS_HOST", "redis.example.com")
	os.Setenv("HAUK_REDIS_PORT", "6380")
	os.Setenv("HAUK_AUTH_METHOD", "htpasswd")
	os.Setenv("HAUK_SERVER_PORT", "9090")
	defer func() {
		os.Unsetenv("HAUK_STORAGE_BACKEND")
		os.Unsetenv("HAUK_REDIS_HOST")
		os.Unsetenv("HAUK_REDIS_PORT")
		os.Unsetenv("HAUK_AUTH_METHOD")
		os.Unsetenv("HAUK_SERVER_PORT")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.StorageBackend != "redis" {
		t.Fatalf("expected storage_backend 'redis', got %q", cfg.StorageBackend)
	}
	if cfg.RedisHost != "redis.example.com" {
		t.Fatalf("expected redis_host 'redis.example.com', got %q", cfg.RedisHost)
	}
	if cfg.RedisPort != 6380 {
		t.Fatalf("expected redis_port 6380, got %d", cfg.RedisPort)
	}
	if cfg.AuthMethod != "htpasswd" {
		t.Fatalf("expected auth_method 'htpasswd', got %q", cfg.AuthMethod)
	}
	if cfg.ServerPort != 9090 {
		t.Fatalf("expected server_port 9090, got %d", cfg.ServerPort)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := `
storage_backend: "redis"
redis_host: "localhost"
redis_port: 6379
auth_method: "password"
server_port: 8888
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	// Temporarily modify ConfigPaths to point to our test file.
	originalPaths := ConfigPaths
	ConfigPaths = []string{configPath}
	defer func() { ConfigPaths = originalPaths }()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.ServerPort != 8888 {
		t.Fatalf("expected server_port 8888, got %d", cfg.ServerPort)
	}
}

func TestVelocityMultiplier(t *testing.T) {
	tests := []struct {
		unit   string
		expect float64
	}{
		{"kmh", 3.6},
		{"mph", 3.6 * 0.6213712},
		{"mps", 1.0},
		{"unknown", 3.6}, // Default.
	}

	for _, tt := range tests {
		t.Run(tt.unit, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.VelocityUnit = tt.unit
			got := cfg.VelocityMultiplier()
			if got != tt.expect {
				t.Fatalf("VelocityMultiplier() = %f, want %f", got, tt.expect)
			}
		})
	}
}
