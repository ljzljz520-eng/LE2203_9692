package integration

import (
	"path/filepath"
	"testing"

	"firmwarehub/internal/archive"
	"firmwarehub/internal/catalog"
	"firmwarehub/internal/domain"
	"firmwarehub/internal/importer"
	"firmwarehub/internal/review"
	"firmwarehub/internal/store"
)

func TestAllWorkflowBoundaries(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "integration.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := catalog.NewService(s)
	reviewService := review.NewService(c, review.DefaultPolicy())
	archiveService := archive.NewService(c)
	record, err := c.Register(catalog.RegisterRequest{ID: "int-1", Product: "gateway", Version: "4", Checksum: "sha256:12345678", DownloadURL: "file:///gateway", Platform: "linux-amd64", Owner: "operator", Score: 84, At: "1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reviewService.Submit(record.ID, "operator", "2"); err != nil {
		t.Fatal(err)
	}
	if _, err := reviewService.Decide(domain.ReviewDecision{RecordID: record.ID, Approved: true, Actor: "reviewer", Reason: "verified", At: "3"}); err != nil {
		t.Fatal(err)
	}
	if _, err := archiveService.Archive(record.ID, "archivist", "4"); err != nil {
		t.Fatal(err)
	}
	processor := importer.NewProcessor(c, s)
	batch, err := importer.NewBatch("int-batch", "import", "5", []domain.ImportRow{{ID: "int-2", Product: "gateway", Version: "5", Checksum: "sha256:87654321", DownloadURL: "file:///gateway2", Platform: "linux-arm64", Score: 75, Owner: "operator", Source: "import"}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := processor.Import(batch)
	if err != nil || len(result.Imported) != 1 || result.TotalScore != 75 {
		t.Fatalf("import boundary: %v %+v", err, result)
	}
}
