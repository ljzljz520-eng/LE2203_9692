package domain

import (
	"fmt"
	"strings"
)

func ValidateRecord(r Record) error {
	checks := []struct {
		ok  bool
		msg string
	}{
		{strings.TrimSpace(r.ID) != "", "id is required"},
		{strings.TrimSpace(r.Product) != "", "product is required"},
		{strings.TrimSpace(r.Version) != "", "version is required"},
		{strings.TrimSpace(r.Checksum) != "", "checksum is required"},
		{strings.TrimSpace(r.Platform) != "", "platform is required"},
		{r.Score >= 0 && r.Score <= 100, "score must be between 0 and 100"},
		{r.Revision >= 0, "revision cannot be negative"},
	}
	for _, check := range checks {
		if !check.ok {
			return fmt.Errorf("%w: %s", ErrValidation, check.msg)
		}
	}
	if len(r.Checksum) < 8 {
		return fmt.Errorf("%w: checksum is too short", ErrValidation)
	}
	if !strings.HasPrefix(r.DownloadURL, "https://") && !strings.HasPrefix(r.DownloadURL, "file://") {
		return fmt.Errorf("%w: download URL must be https or file", ErrValidation)
	}
	return nil
}

func ValidateImportRow(row ImportRow) error {
	if row.ID == "" || row.Product == "" || row.Version == "" {
		return fmt.Errorf("%w: import identity is incomplete", ErrValidation)
	}
	if row.Score < 0 || row.Score > 100 {
		return fmt.Errorf("%w: import score is outside range", ErrValidation)
	}
	if row.Source == "" {
		return fmt.Errorf("%w: import source is required", ErrValidation)
	}
	return nil
}

func NormalizeScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func StateForScore(score int) VerificationState {
	score = NormalizeScore(score)
	if score >= 80 {
		return VerificationPass
	}
	if score >= 50 {
		return VerificationWarn
	}
	return VerificationFail
}

func IsSupportedPlatform(platform string) bool {
	switch strings.ToLower(platform) {
	case "linux-amd64", "linux-arm64", "windows-amd64", "darwin-arm64":
		return true
	default:
		return false
	}
}
