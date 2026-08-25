package store

import (
	"encoding/json"
	"fmt"
	"sort"

	"firmwarehub/internal/domain"
	"go.etcd.io/bbolt"
)

type ExportBundle struct {
	Records     []domain.Record     `json:"records"`
	Audits      []domain.AuditEvent `json:"audits"`
	Workflows   []domain.Workflow   `json:"workflows"`
	Attachments []domain.Attachment `json:"attachments"`
}

func (s *Store) ExportBundle(recordID string) (ExportBundle, error) {
	records, err := s.ListRecords()
	if err != nil {
		return ExportBundle{}, err
	}
	if recordID != "" {
		filtered := make([]domain.Record, 0, 1)
		for _, record := range records {
			if record.ID == recordID {
				filtered = append(filtered, record)
			}
		}
		records = filtered
	}
	audits, err := s.ListAudit(recordID)
	if err != nil {
		return ExportBundle{}, err
	}
	workflows, err := s.ListWorkflows(recordID)
	if err != nil {
		return ExportBundle{}, err
	}
	attachments, err := s.ListAttachments(recordID)
	if err != nil {
		return ExportBundle{}, err
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	sort.Slice(audits, func(i, j int) bool { return audits[i].ID < audits[j].ID })
	sort.Slice(workflows, func(i, j int) bool { return workflows[i].ID < workflows[j].ID })
	sort.Slice(attachments, func(i, j int) bool { return attachments[i].ID < attachments[j].ID })
	return ExportBundle{Records: records, Audits: audits, Workflows: workflows, Attachments: attachments}, nil
}

func EncodeBundle(bundle ExportBundle) ([]byte, error) {
	payload, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode export bundle: %w", err)
	}
	return payload, nil
}

func DecodeBundle(payload []byte) (ExportBundle, error) {
	var bundle ExportBundle
	if len(payload) == 0 {
		return bundle, fmt.Errorf("decode export bundle: empty payload")
	}
	if err := json.Unmarshal(payload, &bundle); err != nil {
		return bundle, fmt.Errorf("decode export bundle: %w", err)
	}
	if bundle.Records == nil {
		bundle.Records = []domain.Record{}
	}
	if bundle.Audits == nil {
		bundle.Audits = []domain.AuditEvent{}
	}
	if bundle.Workflows == nil {
		bundle.Workflows = []domain.Workflow{}
	}
	if bundle.Attachments == nil {
		bundle.Attachments = []domain.Attachment{}
	}
	return bundle, nil
}

func (s *Store) ImportBundle(bundle ExportBundle) error {
	return s.withWrite(func(tx interfaceTx) error {
		for _, record := range bundle.Records {
			if record.ID == "" {
				return fmt.Errorf("bundle contains record without id")
			}
			if err := putJSON(tx, bucketRecords, []byte(record.ID), record); err != nil {
				return err
			}
		}
		for _, event := range bundle.Audits {
			if err := putJSON(tx, bucketAudit, []byte(event.ID), event); err != nil {
				return err
			}
		}
		for _, workflow := range bundle.Workflows {
			if err := putJSON(tx, bucketWorkflows, []byte(workflow.ID), workflow); err != nil {
				return err
			}
		}
		for _, attachment := range bundle.Attachments {
			if err := putJSON(tx, bucketAttachments, []byte(attachment.ID), attachment); err != nil {
				return err
			}
		}
		return nil
	})
}

type interfaceTx = *bbolt.Tx
