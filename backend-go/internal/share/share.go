// Package share manages location shares (SoloShare and GroupShare).
package share

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/hauk/hauk-go/internal/config"
	"github.com/hauk/hauk-go/internal/linkgen"
	"github.com/hauk/hauk-go/internal/session"
	"github.com/hauk/hauk-go/internal/store"
)

const (
	prefixLocData = "-locdata-"
	prefixGroupID = "-groupid-"

	shareTypeAlone  = 0
	shareTypeGroup  = 1

	groupPinMin = 100000
	groupPinMax = 999999
)

// Share is the base type for all shares.
type Share struct {
	id      string
	data    shareData
	store   store.Store
	cfg     *config.Config
	expired bool
}

type shareData struct {
	Type    int                     `json:"type"`
	Expire  int64                   `json:"expire"`
	// SoloShare specific fields
	Host     string `json:"host,omitempty"`
	Adoptable bool `json:"adoptable,omitempty"`
	// GroupShare specific fields
	Hosts    map[string]string `json:"hosts,omitempty"`
	GroupPin int               `json:"group_pin,omitempty"`
}

// NewSoloShare creates a new solo share.
func NewSoloShare(store store.Store, cfg *config.Config) (*Share, error) {
	id, err := linkgen.GenerateID(linkgen.Style(cfg.LinkStyle), store)
	if err != nil {
		return nil, fmt.Errorf("generating share ID: %w", err)
	}
	return &Share{
		id: id,
		data: shareData{
			Type:    shareTypeAlone,
			Expire:  0,
			Adoptable: false,
		},
		store: store,
		cfg:   cfg,
	}, nil
}

// NewGroupShare creates a new group share.
func NewGroupShare(store store.Store, cfg *config.Config) (*Share, error) {
	id, err := linkgen.GenerateID(linkgen.Style(cfg.LinkStyle), store)
	if err != nil {
		return nil, fmt.Errorf("generating share ID: %w", err)
	}
	pin, err := generateGroupPIN(store)
	if err != nil {
		return nil, fmt.Errorf("generating group PIN: %w", err)
	}
	return &Share{
		id: id,
		data: shareData{
			Type:    shareTypeGroup,
			Expire:  0,
			Hosts:   make(map[string]string),
			GroupPin: pin,
		},
		store: store,
		cfg:   cfg,
	}, nil
}

// FromID retrieves an existing share by ID.
func FromID(st store.Store, cfg *config.Config, id string) (*Share, error) {
	fullKey := prefixLocData + id
	data, found := st.Get(fullKey)
	if !found {
		return &Share{
			id: id,
			data: shareData{},
			store: st,
			cfg:   cfg,
			expired: true,
		}, nil
	}

	s := &Share{
		id:    id,
		data:  shareData{},
		store: st,
		cfg:   cfg,
	}

	// Handle both raw struct and map[string]interface{} (from JSON-encoded stores).
	switch v := data.(type) {
	case shareData:
		s.data = v
	case map[string]interface{}:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshaling share data: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &s.data); err != nil {
			return nil, fmt.Errorf("unmarshaling share data: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid share data type for ID %s: %T", id, data)
	}

	if s.HasExpired() {
		s.expired = true
	}

	return s, nil
}

// FromGroupPIN retrieves a share by its group PIN.
func FromGroupPIN(store store.Store, cfg *config.Config, pin int) (*Share, error) {
	fullKey := prefixGroupID + fmt.Sprintf("%d", pin)
	data, found := store.Get(fullKey)
	if !found {
		return &Share{
			data:  shareData{},
			store: store,
			cfg:   cfg,
			expired: true,
		}, nil
	}

	shareID, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("invalid group PIN data")
	}

	return FromID(store, cfg, shareID)
}

func (s *Share) ID() string {
	return s.id
}

func (s *Share) Exists() bool {
	return !s.expired
}

// Save persists the share to the store.
func (s *Share) Save() error {
	if s.data.Expire == 0 {
		return fmt.Errorf("share cannot be indefinite")
	}

	if s.HasExpired() {
		// Share has expired, delete it.
		if err := s.store.Delete(prefixLocData + s.id); err != nil {
			return fmt.Errorf("deleting expired share: %w", err)
		}
		// Also clean up group PIN mapping if applicable.
		if s.data.Type == shareTypeGroup {
			_ = s.store.Delete(prefixGroupID + fmt.Sprintf("%d", s.data.GroupPin))
		}
		return nil
	}

	fullKey := prefixLocData + s.id
	expire := s.expirationSeconds()
	if err := s.store.Set(fullKey, s.data, expire); err != nil {
		return fmt.Errorf("saving share: %w", err)
	}

	// Group shares also save the group PIN mapping.
	if s.data.Type == shareTypeGroup {
		pinKey := prefixGroupID + fmt.Sprintf("%d", s.data.GroupPin)
		if err := s.store.Set(pinKey, s.id, expire); err != nil {
			return fmt.Errorf("saving group PIN mapping: %w", err)
		}
	}

	return nil
}

// End deletes this share from the store.
func (s *Share) End() error {
	if err := s.store.Delete(prefixLocData + s.id); err != nil {
		return fmt.Errorf("deleting share: %w", err)
	}
	if s.data.Type == shareTypeGroup {
		if err := s.store.Delete(prefixGroupID + fmt.Sprintf("%d", s.data.GroupPin)); err != nil {
			return fmt.Errorf("deleting group PIN mapping: %w", err)
		}
	}
	return nil
}

// ViewLink returns the public URL for viewing this share.
func (s *Share) ViewLink() string {
	return s.cfg.PublicURL + s.id
}

// SetExpirationTime sets the expiration time.
func (s *Share) SetExpirationTime(expire time.Time) *Share {
	s.data.Expire = expire.Unix()
	return s
}

// ExpirationTime returns the expiration time.
func (s *Share) ExpirationTime() time.Time {
	return time.Unix(s.data.Expire, 0)
}

// HasExpired returns whether the share has expired.
// A share with Expire=0 (unset) is considered not expired yet.
func (s *Share) HasExpired() bool {
	return s.data.Expire > 0 && s.data.Expire <= time.Now().Unix()
}

// GetExpirationTime returns the expiration Unix timestamp.
func (s *Share) GetExpirationTime() int64 {
	return s.data.Expire
}

// Type returns the share type (0=solo, 1=group).
func (s *Share) Type() int {
	return s.data.Type
}

// GetGroupPin returns the group PIN for group shares.
func (s *Share) GetGroupPin() int {
	return s.data.GroupPin
}

func (s *Share) expirationSeconds() int {
	now := time.Now().Unix()
	if s.data.Expire <= now {
		return 1
	}
	return int(s.data.Expire - now)
}

// --- SoloShare methods ---

// SetAdoptable sets whether this share can be adopted.
func (s *Share) SetAdoptable(adoptable bool) *Share {
	s.data.Adoptable = adoptable
	return s
}

// IsAdoptable returns whether this share can be adopted.
func (s *Share) IsAdoptable() bool {
	return s.data.Adoptable && s.data.Type == shareTypeAlone
}

// SetHost sets the host session for this solo share.
func (s *Share) SetHost(host *session.Session) *Share {
	s.data.Host = host.ID()
	return s
}

// GetHost returns the host session.
func (s *Share) GetHost() *session.Session {
	if s.data.Host == "" {
		return nil
	}
	sess, err := session.FromID(s.store, s.cfg, s.data.Host)
	if err != nil {
		return nil
	}
	return sess
}

// --- GroupShare methods ---

// AddHost adds a new host to this group share.
func (s *Share) AddHost(nick string, host *session.Session) *Share {
	s.data.Hosts[nick] = host.ID()
	s.setAutoExpiration()
	return s
}

// RemoveHost removes a host by session ID.
func (s *Share) RemoveHost(sessionID string) *Share {
	for nick, id := range s.data.Hosts {
		if id == sessionID {
			delete(s.data.Hosts, nick)
		}
	}
	return s
}

// Clean removes non-existent hosts and ends the share if empty.
func (s *Share) Clean() error {
	toRemove := []string{}
	for nick, hostID := range s.data.Hosts {
		host, err := session.FromID(s.store, s.cfg, hostID)
		if err != nil || !host.Exists() || host.HasExpired() {
			toRemove = append(toRemove, nick)
		}
	}

	for _, nick := range toRemove {
		delete(s.data.Hosts, nick)
	}

	if len(s.data.Hosts) == 0 {
		return s.End()
	}

	s.setAutoExpiration()
	return s.Save()
}

// AutoInterval returns the minimum interval across all hosts.
func (s *Share) AutoInterval() float64 {
	minInterval := -1.0
	for _, hostID := range s.data.Hosts {
		host, err := session.FromID(s.store, s.cfg, hostID)
		if err != nil || !host.Exists() {
			continue
		}
		interval := host.Interval()
		if minInterval < 0 || interval < minInterval {
			minInterval = interval
		}
	}
	if minInterval < 0 {
		return 0
	}
	return minInterval
}

// GetAllPoints returns all points from all hosts.
func (s *Share) GetAllPoints(sinceTime *float64) map[string][]map[string]float64 {
	points := make(map[string][]map[string]float64)
	for nick, hostID := range s.data.Hosts {
		host, err := session.FromID(s.store, s.cfg, hostID)
		if err != nil || !host.Exists() {
			continue
		}
		points[nick] = host.GetPoints(sinceTime)
	}
	return points
}

// AutoExpirationTime returns the latest expiration time across all hosts.
func (s *Share) AutoExpirationTime() int64 {
	maxExpire := int64(0)
	for _, hostID := range s.data.Hosts {
		host, err := session.FromID(s.store, s.cfg, hostID)
		if err != nil || !host.Exists() {
			continue
		}
		exp := host.GetExpirationTime()
		if exp > maxExpire {
			maxExpire = exp
		}
	}
	return maxExpire
}

// SetAutoExpirationTime sets the share expiration to the latest host expiration.
func (s *Share) SetAutoExpirationTime() *Share {
	s.data.Expire = s.AutoExpirationTime()
	return s
}

func (s *Share) setAutoExpiration() {
	s.data.Expire = s.AutoExpirationTime()
}

// --- Helpers ---

func generateGroupPIN(st store.Store) (int, error) {
	for {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(groupPinMax-groupPinMin+1)))
		if err != nil {
			return 0, err
		}
		pin := int(n.Int64()) + groupPinMin
		if _, found := st.Get(prefixGroupID + fmt.Sprintf("%d", pin)); !found {
			return pin, nil
		}
	}
}
