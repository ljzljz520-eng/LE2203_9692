package catalog

import (
	"path/filepath"
	"testing"

	"firmwarehub/internal/domain"
	"firmwarehub/internal/store"
)

func TestRegisterChangeAndHistory(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := NewService(s)
	record, err := c.Register(RegisterRequest{ID: "catalog-1", Product: "router", Version: "1", Checksum: "sha256:12345678", DownloadURL: "https://example.invalid/a", Platform: "linux-amd64", Owner: "operator", Score: 60, At: "1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Register(RegisterRequest{ID: record.ID, Product: "router", Version: "1", Checksum: "sha256:12345678", DownloadURL: "https://example.invalid/a", Platform: "linux-amd64", Owner: "operator", Score: 60, At: "1"}); err != domain.ErrAlreadyExists {
		t.Fatalf("duplicate error: %v", err)
	}
	changed, err := c.Change(record.ID, ChangeRequest{Checksum: "sha256:87654321", Notes: "updated", Score: 90, ExpectedRev: 1, At: "2", Actor: "operator"})
	if err != nil || changed.Score != 90 {
		t.Fatalf("change failed: %v %+v", err, changed)
	}
	if _, err := c.Change(record.ID, ChangeRequest{Score: 40, ExpectedRev: 1, At: "3", Actor: "operator"}); err != domain.ErrConflict {
		t.Fatalf("expected conflict: %v", err)
	}
	history, err := c.History(record.ID)
	if err != nil || len(history) < 2 {
		t.Fatalf("history: %v %d", err, len(history))
	}
}

func TestSearchVisibleRecords(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "search.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := NewService(s)
	r, err := c.Register(RegisterRequest{ID: "search-1", Product: "camera", Version: "3", Checksum: "sha256:12345678", DownloadURL: "file:///camera", Platform: "linux-amd64", Owner: "operator", Score: 95, At: "1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Submit(r.ID, "operator", "2"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.TransitionForReview(r.ID, domain.StatusApproved, "reviewer", "3"); err != nil {
		t.Fatal(err)
	}
	items, err := c.Search(SearchQuery{Text: "camera", Status: string(domain.StatusApproved)})
	if err != nil || len(items) != 1 {
		t.Fatalf("search: %v %+v", err, items)
	}
}
