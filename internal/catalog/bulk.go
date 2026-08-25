package catalog

import (
	"fmt"
	"sort"

	"firmwarehub/internal/domain"
)

type BulkChange struct {
	ID               string
	Checksum         string
	DownloadURL      string
	Notes            string
	Score            int
	ExpectedRevision int
}

type BulkResult struct {
	Changed  []domain.Record
	Rejected map[string]string
	Warnings []string
}

type Snapshot struct {
	At           string
	Records      []domain.Record
	ByStatus     map[domain.RecordStatus]int
	ByPlatform   map[string]int
	AverageScore int
}

func (s *Service) PreviewChanges(changes []BulkChange) BulkResult {
	result := BulkResult{Changed: make([]domain.Record, 0), Rejected: make(map[string]string), Warnings: make([]string, 0)}
	for _, change := range changes {
		record, err := s.Get(change.ID)
		if err != nil {
			result.Rejected[change.ID] = err.Error()
			continue
		}
		if change.ExpectedRevision > 0 && change.ExpectedRevision != record.Revision {
			result.Rejected[change.ID] = domain.ErrConflict.Error()
			continue
		}
		updated, err := domain.ApplyChange(record, change.Checksum, change.DownloadURL, change.Notes, record.UpdatedAt, change.Score)
		if err != nil {
			result.Rejected[change.ID] = err.Error()
			continue
		}
		result.Changed = append(result.Changed, updated)
	}
	return result
}

func (s *Service) ApplyChanges(changes []BulkChange, actor, at string) BulkResult {
	result := BulkResult{Changed: make([]domain.Record, 0), Rejected: make(map[string]string), Warnings: make([]string, 0)}
	for _, change := range changes {
		updated, err := s.Change(change.ID, ChangeRequest{Checksum: change.Checksum, DownloadURL: change.DownloadURL, Notes: change.Notes, Score: change.Score, ExpectedRev: change.ExpectedRevision, At: at, Actor: actor})
		if err != nil {
			result.Rejected[change.ID] = err.Error()
			continue
		}
		result.Changed = append(result.Changed, updated)
	}
	if len(result.Rejected) > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("%d changes were rejected", len(result.Rejected)))
	}
	return result
}

func (s *Service) BuildSnapshot(at string) (Snapshot, error) {
	records, err := s.store.ListRecords()
	if err != nil {
		return Snapshot{}, err
	}
	snapshot := Snapshot{At: at, Records: append([]domain.Record(nil), records...), ByStatus: make(map[domain.RecordStatus]int), ByPlatform: make(map[string]int)}
	for _, record := range records {
		snapshot.ByStatus[record.Status]++
		snapshot.ByPlatform[record.Platform]++
		snapshot.AverageScore += record.Score
	}
	if len(records) > 0 {
		snapshot.AverageScore /= len(records)
	}
	sort.Slice(snapshot.Records, func(i, j int) bool {
		if snapshot.Records[i].Product == snapshot.Records[j].Product {
			return snapshot.Records[i].Version < snapshot.Records[j].Version
		}
		return snapshot.Records[i].Product < snapshot.Records[j].Product
	})
	return snapshot, nil
}

func (s *Service) ValidateUniqueChecksum(recordID, checksum string) error {
	records, err := s.store.FindByChecksum(checksum)
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.ID != recordID && !domain.IsTerminal(record.Status) {
			return fmt.Errorf("%w: checksum is used by %s", domain.ErrValidation, record.ID)
		}
	}
	return nil
}

func (s *Service) Publish(id, actor, at string) (domain.Record, error) {
	record, err := s.Get(id)
	if err != nil {
		return domain.Record{}, err
	}
	if record.Status != domain.StatusApproved {
		return domain.Record{}, fmt.Errorf("%w: only approved records can publish", domain.ErrInvalidState)
	}
	if record.Score < 80 {
		return domain.Record{}, fmt.Errorf("%w: published score must be at least 80", domain.ErrValidation)
	}
	return record, s.audit(record, "published", actor, "record published", at)
}

func (s *Service) VisibleSnapshot(at string) (Snapshot, error) {
	snapshot, err := s.BuildSnapshot(at)
	if err != nil {
		return Snapshot{}, err
	}
	visible := make([]domain.Record, 0, len(snapshot.Records))
	for _, record := range snapshot.Records {
		if record.IsVisible() {
			visible = append(visible, record)
		}
	}
	snapshot.Records = visible
	snapshot.ByStatus = make(map[domain.RecordStatus]int)
	snapshot.ByPlatform = make(map[string]int)
	snapshot.AverageScore = 0
	for _, record := range visible {
		snapshot.ByStatus[record.Status]++
		snapshot.ByPlatform[record.Platform]++
		snapshot.AverageScore += record.Score
	}
	if len(visible) > 0 {
		snapshot.AverageScore /= len(visible)
	}
	return snapshot, nil
}
