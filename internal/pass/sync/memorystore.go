package sync

import (
	"sync"
	"time"
)

// MemoryStore keeps in memory pairing between extension and mobile.
//
// TODO: check ttlcache pkg, right now entries are not invalidated.
type MemoryStore struct {
	mu            sync.Mutex
	extensionsMap map[string]Item
}

type Item struct {
	FCMToken  string
	Expires   time.Time
	Confirmed bool
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		extensionsMap: make(map[string]Item),
	}
}

func (s *MemoryStore) RequestSync(fcmToken string) {
	s.setItem(fcmToken, Item{FCMToken: fcmToken})
}

func (s *MemoryStore) ConfirmSync(fcmToken string) bool {
	v, ok := s.getItem(fcmToken)
	if !ok {
		return false
	}
	v.Confirmed = true
	s.setItem(fcmToken, v)
	return true
}

func (s *MemoryStore) IsSyncConfirmed(fcmToken string) bool {
	v, ok := s.getItem(fcmToken)
	if !ok {
		return false
	}
	return v.Confirmed
}

func (s *MemoryStore) setItem(key string, item Item) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.extensionsMap[key] = item
}

func (s *MemoryStore) getItem(key string) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.extensionsMap[key]
	return v, ok
}
