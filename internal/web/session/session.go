package session

import (
	"sync"

	"github.com/gianlucafrei/GoAM/internal/model"
)

var sessions = struct {
	sync.RWMutex
	store map[string]*model.AuthenticationSession
}{
	store: make(map[string]*model.AuthenticationSession),
}

func Save(sessionID string, state *model.AuthenticationSession) {
	sessions.Lock()
	defer sessions.Unlock()
	sessions.store[sessionID] = state
}

func Load(sessionID string) *model.AuthenticationSession {
	sessions.RLock()
	defer sessions.RUnlock()
	return sessions.store[sessionID]
}
