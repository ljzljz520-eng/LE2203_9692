package store

import (
	"fmt"
	"sort"

	"firmwarehub/internal/domain"
	"go.etcd.io/bbolt"
)

func (s *Store) AppendAudit(event domain.AuditEvent) error {
	if event.ID == "" || event.RecordID == "" {
		return fmt.Errorf("audit event identity is required")
	}
	return s.withWrite(func(tx *bbolt.Tx) error {
		return putJSON(tx, bucketAudit, []byte(event.ID), event)
	})
}

func (s *Store) ListAudit(recordID string) ([]domain.AuditEvent, error) {
	events := make([]domain.AuditEvent, 0)
	err := s.withRead(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketAudit).ForEach(func(_, value []byte) error {
			var event domain.AuditEvent
			if err := unmarshal(value, &event); err != nil {
				return err
			}
			if recordID == "" || event.RecordID == recordID {
				events = append(events, event)
			}
			return nil
		})
	})
	sort.Slice(events, func(i, j int) bool {
		if events[i].At == events[j].At {
			return events[i].ID < events[j].ID
		}
		return events[i].At < events[j].At
	})
	return events, err
}
