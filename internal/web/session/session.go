package session

import (
	"goiam/internal/model"
	"sync"
)

var sessions = struct {
	sync.RWMutex
	store map[string]*model.FlowState
}{
	store: make(map[string]*model.FlowState),
}

func Save(sessionID string, state *model.FlowState) {
	sessions.Lock()
	defer sessions.Unlock()
	sessions.store[sessionID] = state
}

func Load(sessionID string) *model.FlowState {
	sessions.RLock()
	defer sessions.RUnlock()
	return sessions.store[sessionID]
}
