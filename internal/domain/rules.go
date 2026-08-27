package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type Reviewability struct {
	Allowed bool
	Reasons []string
}

type ActionRequirement struct {
	Action   string
	Required bool
	Message  string
}

type IntegrityReport struct {
	Valid    bool
	Checks   []string
	Failures []string
}

func ValidateIdentifier(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%w: identifier is empty", ErrValidation)
	}
	if len(trimmed) > 80 {
		return fmt.Errorf("%w: identifier is longer than 80 characters", ErrValidation)
	}
	for _, char := range trimmed {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			continue
		}
		return fmt.Errorf("%w: identifier contains unsupported character", ErrValidation)
	}
	return nil
}

func NormalizePlatform(platform string) string {
	value := strings.ToLower(strings.TrimSpace(platform))
	value = strings.ReplaceAll(value, "_", "-")
	switch value {
	case "amd64-linux", "linux-x86-64":
		return "linux-amd64"
	case "arm64-linux", "linux-aarch64":
		return "linux-arm64"
	case "x86-windows", "windows-x86-64":
		return "windows-amd64"
	default:
		return value
	}
}

func CompareRevision(expected, actual int) error {
	if expected < 1 {
		return fmt.Errorf("%w: expected revision must be positive", ErrValidation)
	}
	if actual < 1 {
		return fmt.Errorf("%w: actual revision must be positive", ErrValidation)
	}
	if expected != actual {
		return ErrConflict
	}
	return nil
}

func StatusLabel(status RecordStatus) string {
	switch status {
	case StatusDraft:
		return "Draft"
	case StatusPending:
		return "Pending review"
	case StatusApproved:
		return "Approved"
	case StatusRejected:
		return "Rejected"
	case StatusArchived:
		return "Archived"
	default:
		return "Unknown"
	}
}

func ScoreBand(score int) string {
	score = NormalizeScore(score)
	switch {
	case score >= 90:
		return "excellent"
	case score >= 80:
		return "strong"
	case score >= 50:
		return "review"
	default:
		return "critical"
	}
}

func BuildFingerprint(record Record) string {
	parts := []string{record.Product, record.Version, record.Checksum, record.Platform, fmt.Sprintf("%d", record.Score)}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

func CheckRecordIntegrity(record Record) IntegrityReport {
	report := IntegrityReport{Valid: true, Checks: make([]string, 0, 7), Failures: make([]string, 0)}
	checks := []struct {
		name string
		ok   bool
	}{
		{"identifier", ValidateIdentifier(record.ID) == nil},
		{"product", strings.TrimSpace(record.Product) != ""},
		{"version", strings.TrimSpace(record.Version) != ""},
		{"checksum", strings.HasPrefix(record.Checksum, "sha256:")},
		{"download_url", strings.HasPrefix(record.DownloadURL, "https://") || strings.HasPrefix(record.DownloadURL, "file://")},
		{"platform", IsSupportedPlatform(record.Platform)},
		{"score", record.Score >= 0 && record.Score <= 100},
	}
	for _, check := range checks {
		if check.ok {
			report.Checks = append(report.Checks, check.name)
		} else {
			report.Valid = false
			report.Failures = append(report.Failures, check.name)
		}
	}
	return report
}

func ReviewRequirements(record Record, policyMinimum int) Reviewability {
	reasons := make([]string, 0, 4)
	if record.Status != StatusPending {
		reasons = append(reasons, "record is not pending")
	}
	if record.Score < policyMinimum {
		reasons = append(reasons, "score is below policy minimum")
	}
	if record.State == VerificationFail {
		reasons = append(reasons, "verification state is fail")
	}
	if !CheckRecordIntegrity(record).Valid {
		reasons = append(reasons, "record integrity checks failed")
	}
	return Reviewability{Allowed: len(reasons) == 0, Reasons: reasons}
}

func RequiredActions(record Record) []ActionRequirement {
	actions := make([]ActionRequirement, 0, 5)
	switch record.Status {
	case StatusDraft:
		actions = append(actions, ActionRequirement{Action: "submit", Required: true, Message: "submit record for review"})
	case StatusPending:
		actions = append(actions, ActionRequirement{Action: "review", Required: true, Message: "reviewer decision required"})
	case StatusApproved:
		actions = append(actions, ActionRequirement{Action: "archive", Required: false, Message: "archive when retention policy permits"})
	case StatusRejected:
		actions = append(actions, ActionRequirement{Action: "change", Required: true, Message: "correct rejected record"})
	case StatusArchived:
		actions = append(actions, ActionRequirement{Action: "none", Required: false, Message: "record is immutable"})
	default:
		actions = append(actions, ActionRequirement{Action: "inspect", Required: true, Message: "inspect unknown status"})
	}
	if record.Score < 80 {
		actions = append(actions, ActionRequirement{Action: "evidence", Required: true, Message: "attach verification evidence"})
	}
	return actions
}

func MergeNotes(existing, incoming string) string {
	left := strings.TrimSpace(existing)
	right := strings.TrimSpace(incoming)
	if left == "" {
		return right
	}
	if right == "" {
		return left
	}
	if left == right {
		return left
	}
	return left + " | " + right
}

func IsTerminal(status RecordStatus) bool {
	return status == StatusArchived
}

func StatusOrder(status RecordStatus) int {
	switch status {
	case StatusDraft:
		return 1
	case StatusPending:
		return 2
	case StatusRejected:
		return 3
	case StatusApproved:
		return 4
	case StatusArchived:
		return 5
	default:
		return 0
	}
}

func CanPublish(record Record) bool {
	return record.Status == StatusApproved && record.Score >= 80 && record.State != VerificationFail
}

func IsChecksumFormatValid(checksum string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(checksum)), "sha256:") && len(strings.TrimSpace(checksum)) >= 15
}

func NextStatus(status RecordStatus) RecordStatus {
	switch status {
	case StatusDraft:
		return StatusPending
	case StatusPending:
		return StatusApproved
	case StatusApproved:
		return StatusArchived
	default:
		return status
	}
}

func StateLabel(state VerificationState) string {
	switch state {
	case VerificationPass:
		return "Pass"
	case VerificationWarn:
		return "Warning"
	case VerificationFail:
		return "Fail"
	default:
		return "Unknown"
	}
}

func ScoreDelta(before, after Record) int {
	return after.Score - before.Score
}

func SortRecordsByScore(records []Record) []Record {
	copyRecords := append([]Record(nil), records...)
	sort.SliceStable(copyRecords, func(i, j int) bool {
		if copyRecords[i].Score == copyRecords[j].Score {
			return copyRecords[i].ID < copyRecords[j].ID
		}
		return copyRecords[i].Score > copyRecords[j].Score
	})
	return copyRecords
}
