package store

import (
	"fmt"
	"sort"
	"strings"

	"firmwarehub/internal/domain"
	"go.etcd.io/bbolt"
)

type QueryOptions struct {
	Text          string
	Products      []string
	Platforms     []string
	Statuses      []domain.RecordStatus
	MinScore      int
	MaxScore      int
	IncludeDrafts bool
	Page          int
	PageSize      int
}

type QueryPage struct {
	Items    []domain.Record
	Page     int
	PageSize int
	Total    int
	HasNext  bool
}

func (s *Store) Query(options QueryOptions) (QueryPage, error) {
	records, err := s.ListRecords()
	if err != nil {
		return QueryPage{}, err
	}
	matched := make([]domain.Record, 0, len(records))
	for _, record := range records {
		if !matchesQuery(record, options) {
			continue
		}
		matched = append(matched, record)
	}
	matched = domain.SortRecordsByScore(matched)
	page := options.Page
	if page < 1 {
		page = 1
	}
	pageSize := options.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(matched) {
		return QueryPage{Items: []domain.Record{}, Page: page, PageSize: pageSize, Total: len(matched), HasNext: false}, nil
	}
	end := start + pageSize
	if end > len(matched) {
		end = len(matched)
	}
	return QueryPage{Items: matched[start:end], Page: page, PageSize: pageSize, Total: len(matched), HasNext: end < len(matched)}, nil
}

func matchesQuery(record domain.Record, options QueryOptions) bool {
	if !options.IncludeDrafts && !record.IsVisible() && len(options.Statuses) == 0 {
		return false
	}
	if options.Text != "" && !strings.Contains(strings.ToLower(record.Summary()), strings.ToLower(options.Text)) {
		return false
	}
	if len(options.Products) > 0 && !containsString(options.Products, record.Product) {
		return false
	}
	if len(options.Platforms) > 0 && !containsString(options.Platforms, record.Platform) {
		return false
	}
	if len(options.Statuses) > 0 && !containsStatus(options.Statuses, record.Status) {
		return false
	}
	if record.Score < options.MinScore {
		return false
	}
	if options.MaxScore > 0 && record.Score > options.MaxScore {
		return false
	}
	return true
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsStatus(values []domain.RecordStatus, target domain.RecordStatus) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (s *Store) FindByChecksum(checksum string) ([]domain.Record, error) {
	records, err := s.ListRecords()
	if err != nil {
		return nil, err
	}
	matched := make([]domain.Record, 0)
	for _, record := range records {
		if record.Checksum == checksum {
			matched = append(matched, record)
		}
	}
	return matched, nil
}

func (s *Store) AuditCount(recordID string) (int, error) {
	events, err := s.ListAudit(recordID)
	if err != nil {
		return 0, err
	}
	return len(events), nil
}

func (s *Store) WorkflowCount(recordID string) (int, error) {
	workflows, err := s.ListWorkflows(recordID)
	if err != nil {
		return 0, err
	}
	return len(workflows), nil
}

func (s *Store) AttachmentCount(recordID string) (int, error) {
	attachments, err := s.ListAttachments(recordID)
	if err != nil {
		return 0, err
	}
	return len(attachments), nil
}

func (s *Store) SaveWorkflowTransition(workflow domain.Workflow, stage, at string) (domain.Workflow, error) {
	if workflow.ID == "" {
		return domain.Workflow{}, fmt.Errorf("workflow id is required")
	}
	if strings.TrimSpace(stage) == "" {
		return domain.Workflow{}, fmt.Errorf("workflow stage is required")
	}
	workflow.Stage = stage
	workflow.UpdatedAt = at
	if err := s.SaveWorkflow(workflow); err != nil {
		return domain.Workflow{}, err
	}
	return workflow, nil
}

func (s *Store) DeleteWorkflow(id string) error {
	return s.withWrite(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketWorkflows)
		if bucket.Get([]byte(id)) == nil {
			return domain.ErrNotFound
		}
		return bucket.Delete([]byte(id))
	})
}

func (s *Store) RecordIDs() ([]string, error) {
	records, err := s.ListRecords()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	sort.Strings(ids)
	return ids, nil
}

func (s *Store) SummarizeRecords() (map[domain.RecordStatus]int, error) {
	records, err := s.ListRecords()
	if err != nil {
		return nil, err
	}
	counts := make(map[domain.RecordStatus]int)
	for _, record := range records {
		counts[record.Status]++
	}
	return counts, nil
}
