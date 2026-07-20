package redis

import (
	"context"
	"testing"
	"time"

	"github.com/hauk/hauk-go/internal/config"
)

// TestStoreInterface verifies the Store interface is satisfied.
func TestStoreInterface(t *testing.T) {
	var _ interface {
		Get(key string) (interface{}, bool)
		Set(key string, data interface{}, expire int) error
		Delete(key string) error
	} = (*Store)(nil)
}

// TestNewStoreWithMock validates the store can be created (requires Redis).
// This test requires a running Redis instance.
func TestNewStoreWithMock(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RedisHost = "localhost"
	cfg.RedisPort = 6379

	ctx := context.Background()
	store, err := NewStore(ctx, cfg)
	if err != nil {
		t.Skipf("Redis not available, skipping: %v", err)
	}
	defer store.Close()

	// Test Set and Get.
	key := "test-key"
	value := map[string]interface{}{"hello": "world"}
	if err := store.Set(key, value, 60); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	data, found := store.Get(key)
	if !found {
		t.Fatal("expected key to be found")
	}

	// Verify the data is correct.
	saved, ok := data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", data)
	}
	if saved["hello"] != "world" {
		t.Fatalf("expected 'world', got %v", saved["hello"])
	}

	// Test Delete.
	if err := store.Delete(key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, found = store.Get(key)
	if found {
		t.Fatal("expected key to be deleted")
	}
}

// TestStoreExpiry validates that expiration works correctly.
func TestStoreExpiry(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RedisHost = "localhost"
	cfg.RedisPort = 6379

	ctx := context.Background()
	store, err := NewStore(ctx, cfg)
	if err != nil {
		t.Skipf("Redis not available, skipping: %v", err)
	}
	defer store.Close()

	// Test with 1-second expiration.
	key := "expiry-test"
	value := "test-value"
	if err := store.Set(key, value, 1); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Should be found immediately.
	_, found := store.Get(key)
	if !found {
		t.Fatal("expected key to be found immediately after set")
	}

	// Wait for expiration.
	time.Sleep(2 * time.Second)

	_, found = store.Get(key)
	if found {
		t.Fatal("expected key to be expired")
	}
}

// TestStorePrefix validates that keys are properly prefixed.
func TestStorePrefix(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RedisHost = "localhost"
	cfg.RedisPort = 6379
	cfg.RedisPrefix = "testprefix"

	ctx := context.Background()
	store, err := NewStore(ctx, cfg)
	if err != nil {
		t.Skipf("Redis not available, skipping: %v", err)
	}
	defer store.Close()

	// Use a unique key.
	key := "unique-prefix-test"
	value := "prefix-test"

	// Clean up any existing key.
	store.Delete(key)

	if err := store.Set(key, value, 0); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	data, found := store.Get(key)
	if !found {
		t.Fatal("expected key to be found")
	}

	saved, ok := data.(string)
	if !ok || saved != value {
		t.Fatalf("expected '%s', got %v (%T)", value, data, data)
	}

	// Clean up.
	store.Delete(key)
}
