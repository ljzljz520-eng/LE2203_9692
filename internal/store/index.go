package store

import (
	"fmt"
	"strings"

	"firmwarehub/internal/domain"
)

type IndexEntry struct {
	RecordID string
	Product  string
	Platform string
	Status   domain.RecordStatus
	Score    int
}

type CatalogIndex struct {
	Entries   []IndexEntry
	Products  map[string]int
	Platforms map[string]int
	Statuses  map[domain.RecordStatus]int
}

func (s *Store) BuildIndex() (CatalogIndex, error) {
	records, err := s.ListRecords()
	if err != nil {
		return CatalogIndex{}, err
	}
	index := CatalogIndex{Entries: make([]IndexEntry, 0, len(records)), Products: make(map[string]int), Platforms: make(map[string]int), Statuses: make(map[domain.RecordStatus]int)}
	for _, record := range records {
		index.Entries = append(index.Entries, IndexEntry{RecordID: record.ID, Product: record.Product, Platform: record.Platform, Status: record.Status, Score: record.Score})
		index.Products[record.Product]++
		index.Platforms[record.Platform]++
		index.Statuses[record.Status]++
	}
	return index, nil
}

func (index CatalogIndex) LookupProduct(product string) []IndexEntry {
	items := make([]IndexEntry, 0)
	for _, entry := range index.Entries {
		if entry.Product == product {
			items = append(items, entry)
		}
	}
	return items
}

func (index CatalogIndex) LookupPlatform(platform string) []IndexEntry {
	items := make([]IndexEntry, 0)
	for _, entry := range index.Entries {
		if entry.Platform == platform {
			items = append(items, entry)
		}
	}
	return items
}

func (index CatalogIndex) Describe() string {
	parts := []string{fmt.Sprintf("entries=%d", len(index.Entries)), fmt.Sprintf("products=%d", len(index.Products)), fmt.Sprintf("platforms=%d", len(index.Platforms)), fmt.Sprintf("statuses=%d", len(index.Statuses))}
	return strings.Join(parts, " ")
}

func (index CatalogIndex) Empty() bool {
	return len(index.Entries) == 0
}
