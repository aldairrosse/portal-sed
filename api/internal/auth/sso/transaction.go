package sso

import (
	"sync"
	"time"
)

type OIDCTransaction struct {
	State         string
	Nonce         string
	CodeVerifier  string
	ReturnTo      string
	RequestedACR  string
	CreatedAt     time.Time
}

func (t *OIDCTransaction) IsStepUp() bool {
	return t.RequestedACR == "mobo-2fa"
}

type TransactionStore struct {
	mu       sync.Mutex
	entries  map[string]*OIDCTransaction
	ttl      time.Duration
}

func NewTransactionStore(ttl time.Duration) *TransactionStore {
	s := &TransactionStore{
		entries: make(map[string]*OIDCTransaction),
		ttl:     ttl,
	}
	go s.cleanupLoop()
	return s
}

func (s *TransactionStore) Store(state string, tx *OIDCTransaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx.CreatedAt = time.Now().UTC()
	s.entries[state] = tx
}

func (s *TransactionStore) Consume(state string) (*OIDCTransaction, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, ok := s.entries[state]
	if !ok {
		return nil, false
	}
	if time.Now().UTC().Sub(tx.CreatedAt) > s.ttl {
		delete(s.entries, state)
		return nil, false
	}
	delete(s.entries, state)
	return tx, true
}

func (s *TransactionStore) cleanupLoop() {
	ticker := time.NewTicker(s.ttl / 2)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now().UTC()
		for state, tx := range s.entries {
			if now.Sub(tx.CreatedAt) > s.ttl {
				delete(s.entries, state)
			}
		}
		s.mu.Unlock()
	}
}
