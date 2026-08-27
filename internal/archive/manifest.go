package archive

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"firmwarehub/internal/domain"
)

type ManifestEntry struct {
	RecordID    string
	Product     string
	Version     string
	Platform    string
	Checksum    string
	Score       int
	Fingerprint string
}

type Manifest struct {
	GeneratedAt string
	Entries     []ManifestEntry
	Digest      string
}

type RetentionPlan struct {
	Eligible []domain.Record
	Blocked  []string
	Reason   map[string]string
}

func BuildManifest(records []domain.Record, generatedAt string) Manifest {
	entries := make([]ManifestEntry, 0, len(records))
	for _, record := range records {
		if !record.IsVisible() {
			continue
		}
		entries = append(entries, ManifestEntry{RecordID: record.ID, Product: record.Product, Version: record.Version, Platform: record.Platform, Checksum: record.Checksum, Score: record.Score, Fingerprint: domain.BuildFingerprint(record)})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Product == entries[j].Product {
			if entries[i].Version == entries[j].Version {
				return entries[i].Platform < entries[j].Platform
			}
			return entries[i].Version < entries[j].Version
		}
		return entries[i].Product < entries[j].Product
	})
	manifest := Manifest{GeneratedAt: generatedAt, Entries: entries}
	manifest.Digest = digestManifest(manifest)
	return manifest
}

func digestManifest(manifest Manifest) string {
	parts := []string{manifest.GeneratedAt}
	for _, entry := range manifest.Entries {
		parts = append(parts, entry.RecordID, entry.Product, entry.Version, entry.Platform, entry.Checksum, fmt.Sprintf("%d", entry.Score), entry.Fingerprint)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

func VerifyManifest(manifest Manifest) bool {
	if manifest.Digest == "" {
		return false
	}
	return digestManifest(manifest) == manifest.Digest
}

func PlanRetention(records []domain.Record, cutoff string) RetentionPlan {
	plan := RetentionPlan{Eligible: make([]domain.Record, 0), Blocked: make([]string, 0), Reason: make(map[string]string)}
	for _, record := range records {
		if record.Status != domain.StatusArchived {
			plan.Blocked = append(plan.Blocked, record.ID)
			plan.Reason[record.ID] = "record is not archived"
			continue
		}
		if cutoff != "" && record.ArchivedAt >= cutoff {
			plan.Blocked = append(plan.Blocked, record.ID)
			plan.Reason[record.ID] = "retention cutoff has not elapsed"
			continue
		}
		plan.Eligible = append(plan.Eligible, record)
	}
	sort.Strings(plan.Blocked)
	sort.Slice(plan.Eligible, func(i, j int) bool { return plan.Eligible[i].ArchivedAt < plan.Eligible[j].ArchivedAt })
	return plan
}

func FormatManifest(manifest Manifest) string {
	lines := []string{fmt.Sprintf("generated_at=%s", manifest.GeneratedAt), fmt.Sprintf("digest=%s", manifest.Digest), fmt.Sprintf("entries=%d", len(manifest.Entries))}
	for _, entry := range manifest.Entries {
		lines = append(lines, fmt.Sprintf("%s %s/%s %s score=%d checksum=%s fingerprint=%s", entry.RecordID, entry.Product, entry.Version, entry.Platform, entry.Score, entry.Checksum, entry.Fingerprint))
	}
	return strings.Join(lines, "\n")
}

func EligibleForArchive(record domain.Record) error {
	if record.Status != domain.StatusApproved {
		return fmt.Errorf("%w: status %s is not approved", domain.ErrInvalidState, record.Status)
	}
	if record.Score < 50 {
		return fmt.Errorf("%w: score below archive threshold", domain.ErrValidation)
	}
	if record.ArchivedAt != "" {
		return fmt.Errorf("%w: record has already been archived", domain.ErrInvalidState)
	}
	return nil
}

func GroupByProduct(records []domain.Record) map[string][]domain.Record {
	groups := make(map[string][]domain.Record)
	for _, record := range records {
		groups[record.Product] = append(groups[record.Product], record)
	}
	for product := range groups {
		groups[product] = domain.SortRecordsByScore(groups[product])
	}
	return groups
}
