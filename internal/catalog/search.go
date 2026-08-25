package catalog

import (
	"sort"
	"strings"

	"firmwarehub/internal/domain"
)

type SearchQuery struct {
	Text     string
	Product  string
	Platform string
	Status   string
	MinScore int
	Limit    int
}

func (s *Service) Search(query SearchQuery) ([]domain.Record, error) {
	records, err := s.store.FindRecords(query.Product, query.Platform, query.Status, query.MinScore)
	if err != nil {
		return nil, err
	}
	text := strings.ToLower(strings.TrimSpace(query.Text))
	matched := make([]domain.Record, 0, len(records))
	for _, record := range records {
		if text != "" && !containsRecordText(record, text) {
			continue
		}
		if !record.IsVisible() && query.Status == "" {
			continue
		}
		matched = append(matched, record)
	}
	sort.SliceStable(matched, func(i, j int) bool {
		if matched[i].Score == matched[j].Score {
			return matched[i].UpdatedAt > matched[j].UpdatedAt
		}
		return matched[i].Score > matched[j].Score
	})
	if query.Limit > 0 && len(matched) > query.Limit {
		matched = matched[:query.Limit]
	}
	return matched, nil
}

func containsRecordText(record domain.Record, text string) bool {
	fields := []string{record.ID, record.Product, record.Version, record.Checksum, record.Platform, record.Owner, record.Notes}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), text) {
			return true
		}
	}
	return false
}

func (s *Service) History(id string) ([]domain.AuditEvent, error) {
	return s.store.ListAudit(id)
}
