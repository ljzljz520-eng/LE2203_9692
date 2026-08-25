package catalog

import (
	"fmt"
	"strings"

	"firmwarehub/internal/domain"
)

type ValidationIssue struct {
	Field    string
	Message  string
	Severity string
}

func InspectRecord(record domain.Record) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if err := domain.ValidateIdentifier(record.ID); err != nil {
		issues = append(issues, ValidationIssue{Field: "id", Message: err.Error(), Severity: "error"})
	}
	if strings.TrimSpace(record.Product) == "" {
		issues = append(issues, ValidationIssue{Field: "product", Message: "product is required", Severity: "error"})
	}
	if strings.TrimSpace(record.Version) == "" {
		issues = append(issues, ValidationIssue{Field: "version", Message: "version is required", Severity: "error"})
	}
	if !strings.HasPrefix(record.Checksum, "sha256:") {
		issues = append(issues, ValidationIssue{Field: "checksum", Message: "sha256 checksum is required", Severity: "error"})
	}
	if record.Score < 50 {
		issues = append(issues, ValidationIssue{Field: "score", Message: "score will require extra evidence", Severity: "warning"})
	}
	if record.Status == domain.StatusArchived && record.ArchivedAt == "" {
		issues = append(issues, ValidationIssue{Field: "archived_at", Message: "archived timestamp is required", Severity: "error"})
	}
	return issues
}

func ValidationSummary(issues []ValidationIssue) string {
	if len(issues) == 0 {
		return "valid"
	}
	errors := 0
	warnings := 0
	for _, issue := range issues {
		if issue.Severity == "error" {
			errors++
		} else {
			warnings++
		}
	}
	return fmt.Sprintf("errors=%d warnings=%d", errors, warnings)
}

func HasBlockingIssues(issues []ValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}

func (s *Service) ValidateBeforeReview(id string) ([]ValidationIssue, error) {
	record, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	return InspectRecord(record), nil
}
