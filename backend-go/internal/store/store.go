// Package store defines the interface for the Hauk storage backend.
package store

// Store is the interface for key-value storage operations.
// All keys are prefixed automatically by the implementation.
type Store interface {
	// Get retrieves a value by key. Returns the value and true if found,
	// or nil and false if not found.
	Get(key string) (interface{}, bool)

	// Set stores a value with the given key, expiring after expire seconds.
	// expire of 0 means no expiration.
	Set(key string, data interface{}, expire int) error

	// Delete removes the value at the given key.
	Delete(key string) error
}
