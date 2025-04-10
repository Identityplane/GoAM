package session

import (
	"goiam/internal/auth/graph"
	"sync"
)

var sessions = struct {
	sync.RWMutex
	store map[string]*graph.FlowState
}{
	store: make(map[string]*graph.FlowState),
}

func Save(sessionID string, state *graph.FlowState) {
	sessions.Lock()
	defer sessions.Unlock()
	sessions.store[sessionID] = state
}

func Load(sessionID string) *graph.FlowState {
	sessions.RLock()
	defer sessions.RUnlock()
	return sessions.store[sessionID]
}
