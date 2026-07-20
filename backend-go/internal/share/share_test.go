package share

import (
	"testing"
	"time"

	"github.com/hauk/hauk-go/internal/config"
	"github.com/hauk/hauk-go/internal/session"
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

func TestNewSoloShare(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()

	share, err := NewSoloShare(st, cfg)
	if err != nil {
		t.Fatalf("NewSoloShare failed: %v", err)
	}

	if share.ID() == "" {
		t.Fatal("expected non-empty share ID")
	}

	if share.Type() != shareTypeAlone {
		t.Fatalf("expected type %d, got %d", shareTypeAlone, share.Type())
	}
}

func TestNewGroupShare(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()

	share, err := NewGroupShare(st, cfg)
	if err != nil {
		t.Fatalf("NewGroupShare failed: %v", err)
	}

	if share.ID() == "" {
		t.Fatal("expected non-empty share ID")
	}

	if share.Type() != shareTypeGroup {
		t.Fatalf("expected type %d, got %d", shareTypeGroup, share.Type())
	}

	pin := share.GetGroupPin()
	if pin < groupPinMin || pin > groupPinMax {
		t.Fatalf("expected group PIN between %d and %d, got %d", groupPinMin, groupPinMax, pin)
	}
}

func TestSoloShareSaveAndRetrieve(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()

	share, _ := NewSoloShare(st, cfg)
	share.SetAdoptable(true)
	share.SetExpirationTime(time.Now().Add(time.Hour))
	if err := share.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := FromID(st, cfg, share.ID())
	if err != nil {
		t.Fatalf("FromID failed: %v", err)
	}

	if !retrieved.Exists() {
		t.Fatal("expected share to exist")
	}

	if retrieved.Type() != shareTypeAlone {
		t.Fatalf("expected type %d, got %d", shareTypeAlone, retrieved.Type())
	}
}

func TestGroupShareHosts(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()

	groupShare, _ := NewGroupShare(st, cfg)

	// Create a mock host session.
	host, _ := session.New(st, cfg)
	host.SetExpirationTime(time.Now().Add(time.Hour)).SetInterval(30)
	host.Save()

	groupShare.AddHost("alice", host)

	// Verify the host was added.
	autoExp := groupShare.AutoExpirationTime()
	if autoExp <= 0 {
		t.Fatal("expected auto expiration time > 0")
	}

	autoInt := groupShare.AutoInterval()
	if autoInt != 30 {
		t.Fatalf("expected auto interval 30, got %f", autoInt)
	}
}

func TestGroupShareRemoveHost(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()

	groupShare, _ := NewGroupShare(st, cfg)
	host, _ := session.New(st, cfg)
	host.SetExpirationTime(time.Now().Add(time.Hour)).SetInterval(30)
	host.Save()
	groupShare.AddHost("alice", host)

	groupShare.RemoveHost(host.ID())
	autoExp := groupShare.AutoExpirationTime()
	if autoExp != 0 {
		t.Fatalf("expected auto expiration 0 after removing all hosts, got %d", autoExp)
	}
}

func TestShareViewLink(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PublicURL = "https://example.com/"

	st := NewInMemoryStore()
	share, _ := NewSoloShare(st, cfg)

	expected := "https://example.com/" + share.ID()
	if share.ViewLink() != expected {
		t.Fatalf("expected view link %q, got %q", expected, share.ViewLink())
	}
}

func TestShareExpiration(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()

	share, _ := NewSoloShare(st, cfg)
	share.SetExpirationTime(time.Now().Add(-time.Hour))

	if !share.HasExpired() {
		t.Fatal("expected share to be expired")
	}
}

func TestShareAdoptable(t *testing.T) {
	st := NewInMemoryStore()
	cfg := config.DefaultConfig()

	share, _ := NewSoloShare(st, cfg)
	share.SetAdoptable(true)

	if !share.IsAdoptable() {
		t.Fatal("expected share to be adoptable")
	}

	share.SetAdoptable(false)
	if share.IsAdoptable() {
		t.Fatal("expected share to not be adoptable")
	}
}

// Verify InMemoryStore implements store.Store.
var _ store.Store = (*InMemoryStore)(nil)
