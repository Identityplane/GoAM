package session

import (
	"goiam/internal/auth/flows"
	"sync"
)

var sessions = struct {
	sync.RWMutex
	store map[string]*flows.FlowState
}{
	store: make(map[string]*flows.FlowState),
}

func Save(sessionID string, state *flows.FlowState) {
	sessions.Lock()
	defer sessions.Unlock()
	sessions.store[sessionID] = state
}

func Load(sessionID string) *flows.FlowState {
	sessions.RLock()
	defer sessions.RUnlock()
	return sessions.store[sessionID]
}
