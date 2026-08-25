package store

import (
	"path/filepath"
	"testing"

	"firmwarehub/internal/domain"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "records.db")
	record := domain.StartRecord("persist-1", "gateway", "2.4", "sha256:12345678", "file:///firmware.bin", "linux-arm64", "operator", "2026-02-01T00:00:00Z", 90)
	first, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.PutRecord(record); err != nil {
		t.Fatal(err)
	}
	if err := first.AppendAudit(domain.AuditEvent{ID: "audit-1", RecordID: record.ID, Action: "registered", Actor: "operator", At: record.CreatedAt, Revision: record.Revision}); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	got, err := second.GetRecord(record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Summary() != record.Summary() {
		t.Fatalf("reopened record mismatch: %s != %s", got.Summary(), record.Summary())
	}
	events, err := second.ListAudit(record.ID)
	if err != nil || len(events) != 1 {
		t.Fatalf("reopened audit mismatch: %v %d", err, len(events))
	}
}

func TestStoreFiltering(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "filter.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, record := range []domain.Record{
		domain.StartRecord("a", "router", "1", "sha256:aaaaaaaa", "file:///a", "linux-amd64", "o", "1", 90),
		domain.StartRecord("b", "router", "2", "sha256:bbbbbbbb", "file:///b", "linux-arm64", "o", "1", 40),
	} {
		if err := s.PutRecord(record); err != nil {
			t.Fatal(err)
		}
	}
	items, err := s.FindRecords("router", "linux-amd64", "", 80)
	if err != nil || len(items) != 1 || items[0].ID != "a" {
		t.Fatalf("unexpected filtered records: %v %+v", err, items)
	}
}
