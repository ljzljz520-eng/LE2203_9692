package flow020

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

func newApplication(t *testing.T) (*Application, *store.Store) {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "flow.db"))
	if err != nil {
		t.Fatal(err)
	}
	c := catalog.NewService(s)
	return NewApplication(c, review.NewService(c, review.DefaultPolicy()), archive.NewService(c), importer.NewProcessor(c, s), archive.NewReporter(s)), s
}

func TestWorkflowCreateReviewArchive(t *testing.T) {
	app, s := newApplication(t)
	defer s.Close()
	record, err := app.CreateReviewArchive(catalog.RegisterRequest{ID: "flow-1", Product: "router", Version: "1", Checksum: "sha256:12345678", DownloadURL: "https://example.invalid/flow-1", Platform: "linux-amd64", Owner: "operator", Score: 90, At: "1"}, "reviewer", "archivist", "2")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != domain.StatusArchived {
		t.Fatalf("unexpected workflow status: %s", record.Status)
	}
}

func TestWorkflowSearchUpdatePublish(t *testing.T) {
	app, s := newApplication(t)
	defer s.Close()
	record, err := app.Catalog.Register(catalog.RegisterRequest{ID: "flow-2", Product: "camera", Version: "1", Checksum: "sha256:12345678", DownloadURL: "file:///flow-2", Platform: "linux-arm64", Owner: "operator", Score: 70, At: "1"})
	if err != nil {
		t.Fatal(err)
	}
	items, err := app.SearchUpdatePublish(record.ID, catalog.ChangeRequest{Checksum: "sha256:87654321", Score: 95, ExpectedRev: 1, At: "2", Actor: "reviewer"}, "3")
	if err != nil || len(items) != 1 || items[0].Status != domain.StatusApproved {
		t.Fatalf("workflow search/update/publish: %v %+v", err, items)
	}
}

func TestWorkflowImportReport(t *testing.T) {
	app, s := newApplication(t)
	defer s.Close()
	rows := []domain.ImportRow{{ID: "flow-3", Product: "sensor", Version: "1", Checksum: "sha256:12345678", DownloadURL: "https://example.invalid/flow-3", Platform: "linux-amd64", Score: 85, Owner: "operator", Source: "import"}}
	batch, err := importer.NewBatch("flow-batch", "import", "1", rows)
	if err != nil {
		t.Fatal(err)
	}
	result, report, err := app.ImportReport(batch, "2")
	if err != nil || len(result.Imported) != 1 || report.Total != 1 {
		t.Fatalf("workflow import/report: %v %+v %+v", err, result, report)
	}
}

func Test2203BusinessRegression(t *testing.T) {
	app, s := newApplication(t)
	defer s.Close()
	rows := []domain.ImportRow{
		{ID: "broken-row", Product: "sensor", Version: "1", Checksum: "", DownloadURL: "https://example.invalid/broken", Platform: "linux-amd64", Score: 15, Owner: "operator", Source: "sync"},
		{ID: "second-row", Product: "sensor", Version: "2", Checksum: "sha256:12345678", DownloadURL: "https://example.invalid/second", Platform: "linux-amd64", Score: 92, Owner: "operator", Source: "sync"},
	}
	batch, err := importer.NewBatch("sync-second-batch", "sync", "1", rows)
	if err != nil {
		t.Fatal(err)
	}
	result, _, err := app.ImportReport(batch, "2")
	if err != nil {
		t.Fatal(err)
	}
	if err := app.RequireImportedScore(result, "second-row", 92); err != nil {
		t.Fatal(err)
	}
}
