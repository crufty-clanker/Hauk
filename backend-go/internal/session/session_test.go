package session

import (
	"testing"
	"time"

	"github.com/hauk/hauk-go/internal/config"
	"github.com/hauk/hauk-go/internal/store"
)

// InMemoryStore is a simple in-memory store for testing.
type InMemoryStore struct {
	data map[string]interface{}
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{data: make(map[string]interface{})}
}

func (s *InMemoryStore) Get(key string) (interface{}, bool) {
	v, ok := s.data[key]
	return v, ok
}

func (s *InMemoryStore) Set(key string, data interface{}, expire int) error {
	s.data[key] = data
	return nil
}

func (s *InMemoryStore) Delete(key string) error {
	delete(s.data, key)
	return nil
}

func TestNewSession(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()

	sess, err := New(st, cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if sess.ID() == "" {
		t.Fatal("expected non-empty session ID")
	}

	// New session should not be expired yet.
	if sess.HasExpired() {
		t.Fatal("new session should not be expired")
	}

	// Save and retrieve.
	sess.SetExpirationTime(time.Now().Add(time.Hour)).SetInterval(10)
	if err := sess.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	saved, err := FromID(st, cfg, sess.ID())
	if err != nil {
		t.Fatalf("FromID failed: %v", err)
	}

	if saved.ID() != sess.ID() {
		t.Fatalf("expected ID %s, got %s", sess.ID(), saved.ID())
	}

	if saved.Interval() != 10 {
		t.Fatalf("expected interval 10, got %f", saved.Interval())
	}

	if !saved.Exists() {
		t.Fatal("expected saved session to exist")
	}
}

func TestSessionExpiration(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()

	sess, _ := New(st, cfg)
	sess.SetExpirationTime(time.Now().Add(-time.Hour)).SetInterval(10)
	sess.Save()

	if !sess.HasExpired() {
		t.Fatal("expected session to be expired")
	}

	if sess.Exists() {
		t.Fatal("expired session should not exist")
	}
}

func TestSessionTargets(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()

	sess, _ := New(st, cfg)
	sess.AddTarget("share1").AddTarget("share2")

	targets := sess.Targets()
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	if targets[0] != "share1" || targets[1] != "share2" {
		t.Fatalf("unexpected targets: %v", targets)
	}

	sess.RemoveTarget("share1")
	targets = sess.Targets()
	if len(targets) != 1 {
		t.Fatalf("expected 1 target after removal, got %d", len(targets))
	}
}

func TestSessionPoints(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()
	cfg.MaxCachedPts = 5

	sess, _ := New(st, cfg)

	// Add more points than max.
	for i := 0; i < 10; i++ {
		sess.AddPoint(map[string]float64{"lat": float64(i), "lon": float64(i), "t2": float64(i)})
	}

	// Should only keep the last MaxCachedPts points.
	points := sess.GetPoints(nil)
	if len(points) != cfg.MaxCachedPts {
		t.Fatalf("expected %d points, got %d", cfg.MaxCachedPts, len(points))
	}

	// Test filtering by since time.
	// Points have t2 values 5,6,7,8,9 (last 5 of 0-9). Filtering with sinceTime=5
	// returns only points where t2 > 5 (i.e., 6,7,8,9 = 4 points).
	sinceTime := float64(5)
	filtered := sess.GetPoints(&sinceTime)
	if len(filtered) != 4 {
		t.Fatalf("expected 4 points after filtering (t2 > 5), got %d", len(filtered))
	}
}

func TestSessionEncrypted(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()

	sess, _ := New(st, cfg)

	if sess.IsEncrypted() {
		t.Fatal("expected not encrypted initially")
	}

	sess.SetEncrypted(true, "test-salt")
	if !sess.IsEncrypted() {
		t.Fatal("expected encrypted after SetEncrypted(true)")
	}
	if sess.GetEncryptionSalt() != "test-salt" {
		t.Fatal("expected salt 'test-salt'")
	}

	sess.SetEncrypted(false, "")
	if sess.IsEncrypted() {
		t.Fatal("expected not encrypted after SetEncrypted(false)")
	}
}

func TestSessionEnd(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()

	sess, _ := New(st, cfg)
	sess.SetExpirationTime(time.Now().Add(time.Hour)).SetInterval(10)
	sess.AddTarget("share1")
	sess.Save()

	// End the session.
	err := sess.End(st)
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}

	// Session should no longer exist.
	saved, _ := FromID(st, cfg, sess.ID())
	if saved.Exists() {
		t.Fatal("expected session to be gone after End")
	}
}

// Verify InMemoryStore implements store.Store.
var _ store.Store = (*InMemoryStore)(nil)
