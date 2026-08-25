package store

import (
	"fmt"
	"sync"

	"firmwarehub/internal/domain"
	"go.etcd.io/bbolt"
)

type Store struct {
	db   *bbolt.DB
	mu   sync.RWMutex
	path string
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	db, err := bbolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	s := &Store{db: db, path: path}
	if err := s.db.Update(initializeBuckets); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) Path() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.path
}

func (s *Store) withRead(fn func(*bbolt.Tx) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return fmt.Errorf("store is closed")
	}
	return s.db.View(fn)
}

func (s *Store) withWrite(fn func(*bbolt.Tx) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return fmt.Errorf("store is closed")
	}
	return s.db.Update(fn)
}

func putJSON(tx *bbolt.Tx, bucket, key []byte, value any) error {
	b := tx.Bucket(bucket)
	if b == nil {
		return fmt.Errorf("bucket %s does not exist", bucket)
	}
	payload, err := marshal(value)
	if err != nil {
		return err
	}
	return b.Put(key, payload)
}

func getJSON(tx *bbolt.Tx, bucket, key []byte, target any) error {
	b := tx.Bucket(bucket)
	if b == nil {
		return fmt.Errorf("bucket %s does not exist", bucket)
	}
	payload := b.Get(key)
	if payload == nil {
		return domain.ErrNotFound
	}
	return unmarshal(payload, target)
}
