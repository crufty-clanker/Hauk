// Package session manages sharing sessions (Client in PHP).
// Each session represents one user and contains all location data for that user.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/hauk/hauk-go/internal/config"
	"github.com/hauk/hauk-go/internal/store"
)

const (
	sessionIDSize = 32
	prefixSession = "-session-"
)

// Session represents a sharing session. Each session represents one user
// and contains all location data for that user.
type Session struct {
	id      string
	data    sessionData
	store   store.Store
	cfg     *config.Config
	expired bool
}

type sessionData struct {
	Expire    int64                  `json:"expire"`
	Interval  float64                `json:"interval"`
	Targets   []string               `json:"targets"`
	Points    []map[string]float64   `json:"points"`
	Encrypted int                    `json:"encrypted"`
	Salt      string                 `json:"salt"`
}

// New creates a new session with random ID.
func New(st store.Store, cfg *config.Config) (*Session, error) {
	id, err := generateSessionID(st)
	if err != nil {
		return nil, fmt.Errorf("generating session ID: %w", err)
	}
	return &Session{
		id:    id,
		data:  sessionData{},
		store: st,
		cfg:   cfg,
	}, nil
}

// FromID retrieves an existing session by ID.
func FromID(st store.Store, cfg *config.Config, sid string) (*Session, error) {
	fullKey := prefixSession + sid
	data, found := st.Get(fullKey)
	if !found {
		return &Session{
			id:      sid,
			data:    sessionData{},
			store:   st,
			cfg:     cfg,
			expired: true,
		}, nil
	}

	s := &Session{
		id:    sid,
		data:  sessionData{},
		store: st,
		cfg:   cfg,
	}

	// Handle both raw struct and map[string]interface{} (from JSON-encoded stores).
	switch v := data.(type) {
	case sessionData:
		s.data = v
	case map[string]interface{}:
		if v, ok := v["expire"].(float64); ok {
			s.data.Expire = int64(v)
		}
		if v, ok := v["interval"].(float64); ok {
			s.data.Interval = v
		}
		if v, ok := v["targets"].([]interface{}); ok {
			for _, t := range v {
				if id, ok := t.(string); ok {
					s.data.Targets = append(s.data.Targets, id)
				}
			}
		}
		if v, ok := v["points"].([]interface{}); ok {
			for _, p := range v {
				if pm, ok := p.(map[string]interface{}); ok {
					pt := make(map[string]float64)
					for k, val := range pm {
						if f, ok := val.(float64); ok {
							pt[k] = f
						}
					}
					s.data.Points = append(s.data.Points, pt)
				}
			}
		}
		if v, ok := v["encrypted"].(float64); ok {
			s.data.Encrypted = int(v)
		}
		if v, ok := v["salt"].(string); ok {
			s.data.Salt = v
		}
	default:
		return nil, fmt.Errorf("invalid session data type for ID %s: %T", sid, data)
	}

	// Check if session has expired.
	if s.HasExpired() {
		s.expired = true
	}

	return s, nil
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Exists() bool {
	return !s.expired && s.data.Expire > 0 && !s.HasExpired()
}

// Save persists the session to the store.
func (s *Session) Save() error {
	if s.data.Interval == 0 {
		return fmt.Errorf("session interval is undefined")
	}
	if s.data.Expire == 0 {
		return fmt.Errorf("session cannot be indefinite")
	}

	if s.HasExpired() {
		// Session has already expired, delete it.
		return s.store.Delete(prefixSession + s.id)
	}

	fullKey := prefixSession + s.id
	expire := s.expirationSeconds()
	return s.store.Set(fullKey, s.data, expire)
}

// End deletes the session and all associated shares.
func (s *Session) End(st store.Store) error {
	if !s.Exists() {
		return nil
	}

	// Delete the session.
	if err := st.Delete(prefixSession + s.id); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	// Remove from all target shares.
	for _, targetID := range s.data.Targets {
		targetKey := prefixSession + targetID
		if data, found := st.Get(targetKey); found {
			switch v := data.(type) {
			case map[string]interface{}:
				if hosts, ok := v["hosts"].(map[string]interface{}); ok {
					for nick, hostID := range hosts {
						if hostIDStr, ok := hostID.(string); ok && hostIDStr == s.id {
							delete(hosts, nick)
						}
					}
				}
			}
		}
	}

	return nil
}

// SetExpirationTime sets the expiration time for this session.
func (s *Session) SetExpirationTime(expire time.Time) *Session {
	s.data.Expire = expire.Unix()
	return s
}

// ExpirationTime returns the current expiration time.
func (s *Session) ExpirationTime() time.Time {
	return time.Unix(s.data.Expire, 0)
}

// HasExpired returns whether the session has expired.
// A session with Expire=0 (unset) is considered not expired yet.
func (s *Session) HasExpired() bool {
	return s.data.Expire > 0 && s.data.Expire <= time.Now().Unix()
}

// SetInterval sets the sharing interval in seconds.
func (s *Session) SetInterval(interval float64) *Session {
	s.data.Interval = interval
	return s
}

// Interval returns the current sharing interval in seconds.
func (s *Session) Interval() float64 {
	return s.data.Interval
}

// SetEncrypted marks the session as end-to-end encrypted.
func (s *Session) SetEncrypted(encrypted bool, salt string) *Session {
	if encrypted {
		s.data.Encrypted = 1
	} else {
		s.data.Encrypted = 0
	}
	s.data.Salt = salt
	return s
}

// IsEncrypted returns whether this session is end-to-end encrypted.
func (s *Session) IsEncrypted() bool {
	return s.data.Encrypted > 0
}

// GetEncryptionSalt returns the salt used for end-to-end encryption.
func (s *Session) GetEncryptionSalt() string {
	return s.data.Salt
}

// AddTarget adds a share to this session's target list.
func (s *Session) AddTarget(shareID string) *Session {
	s.data.Targets = append(s.data.Targets, shareID)
	return s
}

// Targets returns the list of share IDs this session targets.
func (s *Session) Targets() []string {
	return s.data.Targets
}

// RemoveTarget removes a share from this session's target list.
func (s *Session) RemoveTarget(shareID string) *Session {
	for i, id := range s.data.Targets {
		if id == shareID {
			s.data.Targets = append(s.data.Targets[:i], s.data.Targets[i+1:]...)
			break
		}
	}
	return s
}

// AddPoint adds a location point to this session.
func (s *Session) AddPoint(point map[string]float64) *Session {
	s.data.Points = append(s.data.Points, point)
	// Keep only the max number of points.
	max := s.cfg.MaxCachedPts
	if len(s.data.Points) > max {
		s.data.Points = s.data.Points[len(s.data.Points)-max:]
	}
	return s
}

// GetPoints returns all points, optionally filtered by since time.
func (s *Session) GetPoints(sinceTime *float64) []map[string]float64 {
	if sinceTime == nil {
		return s.data.Points
	}

	var result []map[string]float64
	// For encrypted sessions, timestamp is at index 3; otherwise index 2.
	timeIdx := 2
	if s.data.Encrypted > 0 {
		timeIdx = 3
	}

	for _, point := range s.data.Points {
		if t, ok := point[fmt.Sprintf("t%d", timeIdx)]; ok && t > *sinceTime {
			result = append(result, point)
		}
	}

	return result
}

// GetExpirationTime returns the expiration Unix timestamp.
func (s *Session) GetExpirationTime() int64 {
	return s.data.Expire
}

func (s *Session) expirationSeconds() int {
	now := time.Now().Unix()
	if s.data.Expire <= now {
		return 1
	}
	return int(s.data.Expire - now)
}

// generateSessionID generates a random session ID.
func generateSessionID(st store.Store) (string, error) {
	for {
		b := make([]byte, sessionIDSize)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		id := hex.EncodeToString(b)
		if _, found := st.Get(prefixSession + id); !found {
			return id, nil
		}
	}
}
