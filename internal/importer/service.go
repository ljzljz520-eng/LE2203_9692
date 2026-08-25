package importer

import (
	"fmt"

	"firmwarehub/internal/catalog"
	"firmwarehub/internal/domain"
	"firmwarehub/internal/store"
)

type Processor struct {
	catalog *catalog.Service
	store   *store.Store
}

func NewProcessor(c *catalog.Service, s *store.Store) *Processor {
	return &Processor{catalog: c, store: s}
}

func (p *Processor) Import(batch Batch) (domain.ImportResult, error) {
	result := domain.ImportResult{BatchID: batch.ID, Imported: make([]domain.Record, 0, len(batch.Rows)), Rejected: make([]string, 0), Warnings: make([]string, 0)}
	for index, row := range batch.Rows {
		record, err := p.processRow(batch, row, index)
		if err != nil {
			result.Rejected = append(result.Rejected, fmt.Sprintf("%s: %v", row.ID, err))
			result.Warnings = append(result.Warnings, fmt.Sprintf("row %d rejected", index+1))
			continue
		}
		result.Imported = append(result.Imported, record)
		result.TotalScore += record.Score
	}
	return result, nil
}

func (p *Processor) processRow(batch Batch, row domain.ImportRow, index int) (domain.Record, error) {
	if err := domain.ValidateImportRow(row); err != nil {
		return domain.Record{}, p.wrapRowError(batch, index, err)
	}
	if row.Checksum == "" {
		return domain.Record{}, p.wrapRowError(batch, index, fmt.Errorf("%w: checksum is required", domain.ErrValidation))
	}
	record := domain.StartRecord(row.ID, row.Product, row.Version, row.Checksum, row.DownloadURL, row.Platform, row.Owner, batch.StartedAt, row.Score)
	if !domain.IsSupportedPlatform(record.Platform) {
		return domain.Record{}, p.wrapRowError(batch, index, fmt.Errorf("%w: unsupported platform", domain.ErrValidation))
	}
	if err := domain.ValidateRecord(record); err != nil {
		return domain.Record{}, p.wrapRowError(batch, index, err)
	}
	if existing, err := p.store.GetRecord(record.ID); err == nil {
		if existing.Status == domain.StatusArchived {
			return domain.Record{}, p.wrapRowError(batch, index, fmt.Errorf("%w: archived record", domain.ErrInvalidState))
		}
		return domain.Record{}, p.wrapRowError(batch, index, domain.ErrAlreadyExists)
	}
	if err := p.store.PutRecord(record); err != nil {
		return domain.Record{}, p.wrapRowError(batch, index, err)
	}
	if err := p.store.AppendAudit(domain.AuditEvent{ID: fmt.Sprintf("%s-import-%02d", batch.ID, index), RecordID: record.ID, Action: "imported", Actor: row.Owner, Message: batch.Source, At: batch.StartedAt, Revision: record.Revision}); err != nil {
		return domain.Record{}, p.wrapRowError(batch, index, err)
	}
	return record, nil
}

func (p *Processor) wrapRowError(batch Batch, index int, err error) error {
	return fmt.Errorf("batch %s row %d: %w", batch.ID, index+1, err)
}

func (p *Processor) Preview(batch Batch) (map[string]int, error) {
	counts := map[string]int{"pass": 0, "warn": 0, "fail": 0}
	for _, row := range batch.Rows {
		if err := domain.ValidateImportRow(row); err != nil {
			continue
		}
		switch domain.StateForScore(row.Score) {
		case domain.VerificationPass:
			counts["pass"]++
		case domain.VerificationWarn:
			counts["warn"]++
		default:
			counts["fail"]++
		}
	}
	return counts, nil
}
