package archive

import (
	"path/filepath"
	"strings"
	"testing"

	"firmwarehub/internal/catalog"
	"firmwarehub/internal/domain"
	"firmwarehub/internal/store"
)

func TestArchiveReport(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "archive.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := catalog.NewService(s)
	r := NewService(c)
	record, err := c.Register(catalog.RegisterRequest{ID: "archive-1", Product: "sensor", Version: "5", Checksum: "sha256:12345678", DownloadURL: "file:///sensor", Platform: "linux-arm64", Owner: "operator", Score: 88, At: "1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Submit(record.ID, "operator", "2"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.TransitionForReview(record.ID, domain.StatusApproved, "reviewer", "3"); err != nil {
		t.Fatal(err)
	}
	archived, err := r.Archive(record.ID, "archivist", "4")
	if err != nil || archived.Status != domain.StatusArchived {
		t.Fatalf("archive: %v %+v", err, archived)
	}
	report, err := NewReporter(s).Build("5")
	if err != nil || report.Archived != 1 || report.Total != 1 || report.AverageScore != 88 {
		t.Fatalf("report: %v %+v", err, report)
	}
	if !strings.Contains(NewReporter(s).Render(report), "archived=1") {
		t.Fatal("render missing archive count")
	}
}
