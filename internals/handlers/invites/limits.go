package invites

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

type Limits struct {
	mu      sync.Mutex
	entries map[uuid.UUID]limitEntry
}

type limitEntry struct {
	limiter *rate.Limiter
	last    time.Time
}

func NewLimits() *Limits { return &Limits{entries: make(map[uuid.UUID]limitEntry)} }

func (l *Limits) allow(user uuid.UUID, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, exists := l.entries[user]
	if !exists && len(l.entries) >= 4096 {
		for key, old := range l.entries {
			if now.Sub(old.last) >= 10*time.Minute {
				delete(l.entries, key)
			}
		}
		if len(l.entries) >= 4096 {
			return false
		}
	}
	if !exists {
		entry.limiter = rate.NewLimiter(rate.Every(2*time.Second), 6)
	}
	entry.last = now
	l.entries[user] = entry
	return entry.limiter.AllowN(now, 1)
}
