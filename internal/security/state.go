package security

import "sync"

type SecurityState struct {
	Tainted bool
	mu      sync.RWMutex
}

var GlobalState = &SecurityState{}

// MarkTainted sets the tainted flag to true.
func (s *SecurityState) MarkTainted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tainted = true
}

// IsTainted returns the current status of the tainted flag.
func (s *SecurityState) IsTainted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Tainted
}
