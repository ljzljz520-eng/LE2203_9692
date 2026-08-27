package importer

import (
	"fmt"

	"firmwarehub/internal/domain"
)

type Batch struct {
	ID        string
	Rows      []domain.ImportRow
	Source    string
	StartedAt string
}

func NewBatch(id, source, startedAt string, rows []domain.ImportRow) (Batch, error) {
	if id == "" || source == "" {
		return Batch{}, fmt.Errorf("%w: batch identity is required", domain.ErrValidation)
	}
	if len(rows) == 0 {
		return Batch{}, fmt.Errorf("%w: batch has no rows", domain.ErrValidation)
	}
	copyRows := append([]domain.ImportRow(nil), rows...)
	return Batch{ID: id, Rows: copyRows, Source: source, StartedAt: startedAt}, nil
}

func (b Batch) Size() int {
	return len(b.Rows)
}
