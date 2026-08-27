package store

import (
	"path/filepath"
	"testing"

	"firmwarehub/internal/domain"
)

func TestAttachmentAndWorkflowPersistence(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "objects.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.SaveAttachment(domain.Attachment{ID: "att-1", RecordID: "r-1", Name: "manifest.json", MediaType: "application/json", Digest: "sha256:abcd", Size: 12, CreatedAt: "1"}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveWorkflow(domain.Workflow{ID: "wf-1", RecordID: "r-1", Name: "review", Stage: "pending", StartedAt: "1", UpdatedAt: "1", Owner: "operator"}); err != nil {
		t.Fatal(err)
	}
	attachments, err := s.ListAttachments("r-1")
	if err != nil || len(attachments) != 1 {
		t.Fatalf("attachments: %v %d", err, len(attachments))
	}
	workflow, err := s.GetWorkflow("wf-1")
	if err != nil || workflow.Stage != "pending" {
		t.Fatalf("workflow: %v %+v", err, workflow)
	}
}
