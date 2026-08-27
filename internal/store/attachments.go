package store

import (
	"fmt"

	"firmwarehub/internal/domain"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveAttachment(attachment domain.Attachment) error {
	if attachment.ID == "" || attachment.RecordID == "" || attachment.Digest == "" {
		return fmt.Errorf("attachment identity is required")
	}
	if attachment.Size < 0 {
		return fmt.Errorf("attachment size cannot be negative")
	}
	return s.withWrite(func(tx *bbolt.Tx) error {
		return putJSON(tx, bucketAttachments, []byte(attachment.ID), attachment)
	})
}

func (s *Store) ListAttachments(recordID string) ([]domain.Attachment, error) {
	items := make([]domain.Attachment, 0)
	err := s.withRead(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketAttachments).ForEach(func(_, value []byte) error {
			var attachment domain.Attachment
			if err := unmarshal(value, &attachment); err != nil {
				return err
			}
			if recordID == "" || attachment.RecordID == recordID {
				items = append(items, attachment)
			}
			return nil
		})
	})
	return items, err
}
