package domain

import "strings"

func ScoreFromSignals(checksumOK, signatureOK, mirrorOK bool) int {
	score := 0
	if checksumOK {
		score += 50
	}
	if signatureOK {
		score += 30
	}
	if mirrorOK {
		score += 20
	}
	return score
}

func ExplainScore(r Record) []string {
	parts := make([]string, 0, 3)
	if r.Score >= 80 {
		parts = append(parts, "verification passed")
	} else if r.Score >= 50 {
		parts = append(parts, "verification needs review")
	} else {
		parts = append(parts, "verification failed")
	}
	if strings.HasPrefix(r.Checksum, "sha256:") {
		parts = append(parts, "sha256 checksum")
	}
	if r.DownloadURL == "" {
		parts = append(parts, "download location missing")
	}
	return parts
}
