package importer

import (
	"fmt"
	"sort"

	"firmwarehub/internal/domain"
)

type RowDiff struct {
	ID     string
	Fields []string
	Before domain.ImportRow
	After  domain.ImportRow
}

type BatchDiff struct {
	Added     []domain.ImportRow
	Removed   []domain.ImportRow
	Changed   []RowDiff
	Unchanged []string
}

func DiffBatches(previous, current Batch) BatchDiff {
	before := make(map[string]domain.ImportRow, len(previous.Rows))
	for _, row := range previous.Rows {
		before[row.ID] = row
	}
	after := make(map[string]domain.ImportRow, len(current.Rows))
	for _, row := range current.Rows {
		after[row.ID] = row
	}
	diff := BatchDiff{Added: make([]domain.ImportRow, 0), Removed: make([]domain.ImportRow, 0), Changed: make([]RowDiff, 0), Unchanged: make([]string, 0)}
	for id, row := range after {
		old, ok := before[id]
		if !ok {
			diff.Added = append(diff.Added, row)
			continue
		}
		fields := changedFields(old, row)
		if len(fields) == 0 {
			diff.Unchanged = append(diff.Unchanged, id)
		} else {
			diff.Changed = append(diff.Changed, RowDiff{ID: id, Fields: fields, Before: old, After: row})
		}
	}
	for id, row := range before {
		if _, ok := after[id]; !ok {
			diff.Removed = append(diff.Removed, row)
		}
	}
	sort.Slice(diff.Added, func(i, j int) bool { return diff.Added[i].ID < diff.Added[j].ID })
	sort.Slice(diff.Removed, func(i, j int) bool { return diff.Removed[i].ID < diff.Removed[j].ID })
	sort.Slice(diff.Changed, func(i, j int) bool { return diff.Changed[i].ID < diff.Changed[j].ID })
	sort.Strings(diff.Unchanged)
	return diff
}

func changedFields(before, after domain.ImportRow) []string {
	fields := make([]string, 0, 8)
	checks := []struct {
		name      string
		different bool
	}{
		{"product", before.Product != after.Product},
		{"version", before.Version != after.Version},
		{"checksum", before.Checksum != after.Checksum},
		{"download_url", before.DownloadURL != after.DownloadURL},
		{"platform", before.Platform != after.Platform},
		{"score", before.Score != after.Score},
		{"owner", before.Owner != after.Owner},
		{"source", before.Source != after.Source},
	}
	for _, check := range checks {
		if check.different {
			fields = append(fields, check.name)
		}
	}
	return fields
}

func ValidateDiff(diff BatchDiff) error {
	seen := make(map[string]bool)
	for _, row := range diff.Added {
		if seen[row.ID] {
			return fmt.Errorf("duplicate added id %s", row.ID)
		}
		seen[row.ID] = true
	}
	for _, change := range diff.Changed {
		if len(change.Fields) == 0 {
			return fmt.Errorf("changed row %s has no changed fields", change.ID)
		}
	}
	return nil
}

func (d BatchDiff) Counts() map[string]int {
	return map[string]int{"added": len(d.Added), "removed": len(d.Removed), "changed": len(d.Changed), "unchanged": len(d.Unchanged)}
}

func (d BatchDiff) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Changed) == 0
}
