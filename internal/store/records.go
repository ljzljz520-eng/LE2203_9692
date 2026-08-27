package store

import (
	"bytes"
	"sort"

	"firmwarehub/internal/domain"
	"go.etcd.io/bbolt"
)

func (s *Store) PutRecord(record domain.Record) error {
	return s.withWrite(func(tx *bbolt.Tx) error {
		return putJSON(tx, bucketRecords, []byte(record.ID), record)
	})
}

func (s *Store) GetRecord(id string) (domain.Record, error) {
	var record domain.Record
	err := s.withRead(func(tx *bbolt.Tx) error {
		return getJSON(tx, bucketRecords, []byte(id), &record)
	})
	return record, err
}

func (s *Store) DeleteRecord(id string) error {
	return s.withWrite(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketRecords)
		if bucket.Get([]byte(id)) == nil {
			return domain.ErrNotFound
		}
		return bucket.Delete([]byte(id))
	})
}

func (s *Store) ListRecords() ([]domain.Record, error) {
	result := make([]domain.Record, 0)
	err := s.withRead(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketRecords).ForEach(func(_, value []byte) error {
			var record domain.Record
			if err := unmarshal(value, &record); err != nil {
				return err
			}
			result = append(result, record)
			return nil
		})
	})
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, err
}

func (s *Store) FindRecords(product, platform, status string, minScore int) ([]domain.Record, error) {
	records, err := s.ListRecords()
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.Record, 0, len(records))
	for _, record := range records {
		if product != "" && !bytes.Equal([]byte(record.Product), []byte(product)) {
			continue
		}
		if platform != "" && record.Platform != platform {
			continue
		}
		if status != "" && string(record.Status) != status {
			continue
		}
		if record.Score < minScore {
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered, nil
}

func (s *Store) CountRecords() (int, error) {
	count := 0
	err := s.withRead(func(tx *bbolt.Tx) error {
		count = tx.Bucket(bucketRecords).Stats().KeyN
		return nil
	})
	return count, err
}
