package domain

import "testing"

func TestRecordValidationAndTransitions(t *testing.T) {
	record := StartRecord("r1", "router", "1.0", "sha256:abcdef12", "https://example.invalid/r1", "linux-amd64", "operator", "2026-01-01T00:00:00Z", 85)
	if err := ValidateRecord(record); err != nil {
		t.Fatal(err)
	}
	if record.State != VerificationPass || !record.IsMutable() {
		t.Fatalf("unexpected initial state: %+v", record)
	}
	pending, err := Transition(record, StatusPending, "2026-01-01T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Transition(pending, StatusArchived, "2026-01-01T02:00:00Z"); err == nil {
		t.Fatal("expected invalid transition")
	}
}

func TestScoringSignals(t *testing.T) {
	if ScoreFromSignals(true, true, false) != 80 {
		t.Fatal("unexpected score")
	}
	if StateForScore(49) != VerificationFail || StateForScore(50) != VerificationWarn || StateForScore(80) != VerificationPass {
		t.Fatal("unexpected score states")
	}
}
