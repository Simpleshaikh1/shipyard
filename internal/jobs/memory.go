package jobs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"sync"
	"time"
)

// MemoryStore is a Store kept in a map. The mutex matters: net/http
// serves every request on its own goroutine, so the map is shared.
type MemoryStore struct {
	mu   sync.RWMutex
	jobs map[string]Job
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{jobs: make(map[string]Job)}
}

func (s *MemoryStore) Create(_ context.Context, j Job) (Job, error) {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	now := time.Now().UTC()
	j.ID = hex.EncodeToString(b)
	j.Status = StatusPending
	j.CreatedAt, j.UpdatedAt = now, now

	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[j.ID] = j
	return j, nil
}

func (s *MemoryStore) Get(_ context.Context, id string) (Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	if !ok {
		return Job{}, ErrNotFound
	}
	return j, nil
}

func (s *MemoryStore) List(_ context.Context) ([]Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		out = append(out, j)
	}
	sort.Slice(out, func(a, b int) bool { return out[a].CreatedAt.Before(out[b].CreatedAt) })
	return out, nil
}
