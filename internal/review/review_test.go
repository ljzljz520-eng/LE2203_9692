package review

import (
	"path/filepath"
	"testing"

	"firmwarehub/internal/catalog"
	"firmwarehub/internal/domain"
	"firmwarehub/internal/store"
)

func TestReviewApprovalAndRejection(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "review.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := catalog.NewService(s)
	r := NewService(c, DefaultPolicy())
	good, err := c.Register(catalog.RegisterRequest{ID: "good", Product: "switch", Version: "1", Checksum: "sha256:12345678", DownloadURL: "https://example.invalid/good", Platform: "linux-amd64", Owner: "operator", Score: 80, At: "1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Submit(good.ID, "operator", "2"); err != nil {
		t.Fatal(err)
	}
	approved, err := r.Decide(domain.ReviewDecision{RecordID: good.ID, Approved: true, Actor: "reviewer", Reason: "verified", At: "3"})
	if err != nil || approved.Status != domain.StatusApproved {
		t.Fatalf("approval: %v %+v", err, approved)
	}
	bad, err := c.Register(catalog.RegisterRequest{ID: "bad", Product: "switch", Version: "1", Checksum: "sha256:12345678", DownloadURL: "https://example.invalid/bad", Platform: "linux-amd64", Owner: "operator", Score: 20, At: "1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Submit(bad.ID, "operator", "2"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Decide(domain.ReviewDecision{RecordID: bad.ID, Approved: true, Actor: "reviewer", Reason: "verified", At: "3"}); err == nil {
		t.Fatal("expected low score rejection")
	}
	rejected, err := r.Decide(domain.ReviewDecision{RecordID: bad.ID, Approved: false, Actor: "reviewer", Reason: "score too low", At: "4"})
	if err != nil || rejected.Status != domain.StatusRejected {
		t.Fatalf("rejection: %v %+v", err, rejected)
	}
}
