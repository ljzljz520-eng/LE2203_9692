package store

import (
	"fmt"

	"firmwarehub/internal/domain"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveWorkflow(workflow domain.Workflow) error {
	if workflow.ID == "" || workflow.RecordID == "" {
		return fmt.Errorf("workflow identity is required")
	}
	return s.withWrite(func(tx *bbolt.Tx) error {
		return putJSON(tx, bucketWorkflows, []byte(workflow.ID), workflow)
	})
}

func (s *Store) GetWorkflow(id string) (domain.Workflow, error) {
	var workflow domain.Workflow
	err := s.withRead(func(tx *bbolt.Tx) error {
		return getJSON(tx, bucketWorkflows, []byte(id), &workflow)
	})
	return workflow, err
}

func (s *Store) ListWorkflows(recordID string) ([]domain.Workflow, error) {
	items := make([]domain.Workflow, 0)
	err := s.withRead(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketWorkflows).ForEach(func(_, value []byte) error {
			var workflow domain.Workflow
			if err := unmarshal(value, &workflow); err != nil {
				return err
			}
			if recordID == "" || workflow.RecordID == recordID {
				items = append(items, workflow)
			}
			return nil
		})
	})
	return items, err
}
