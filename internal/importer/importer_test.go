package importer

import (
	"path/filepath"
	"testing"

	"firmwarehub/internal/catalog"
	"firmwarehub/internal/domain"
	"firmwarehub/internal/store"
)

func TestCSVParsingAndImport(t *testing.T) {
	input := "id,product,version,checksum,download_url,platform,score,owner,source\n" +
		"i1,router,1,sha256:12345678,file:///i1,linux-amd64,82,operator,import\n"
	rows, warnings, err := ParseCSV(input, "import")
	if err != nil || len(rows) != 1 || len(warnings) != 0 {
		t.Fatalf("parse: %v %+v %+v", err, rows, warnings)
	}
	s, err := store.Open(filepath.Join(t.TempDir(), "import.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p := NewProcessor(catalog.NewService(s), s)
	batch, err := NewBatch("batch-1", "import", "1", rows)
	if err != nil {
		t.Fatal(err)
	}
	result, err := p.Import(batch)
	if err != nil || len(result.Imported) != 1 || result.Imported[0].State != domain.VerificationPass {
		t.Fatalf("import: %v %+v", err, result)
	}
}

func TestImportRejectsInvalidRows(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "invalid.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p := NewProcessor(catalog.NewService(s), s)
	batch, err := NewBatch("batch-2", "import", "1", []domain.ImportRow{{ID: "bad", Product: "router", Version: "1", Score: 10}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := p.Import(batch)
	if err != nil || len(result.Imported) != 0 || len(result.Rejected) != 1 {
		t.Fatalf("invalid import: %v %+v", err, result)
	}
}
