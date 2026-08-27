package importer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"firmwarehub/internal/domain"
)

type BatchSummary struct {
	BatchID      string
	Source       string
	Rows         int
	Imported     int
	Rejected     int
	TotalScore   int
	AverageScore int
	States       map[domain.VerificationState]int
	RejectedIDs  []string
}

type RowDecision struct {
	RowID    string
	Accepted bool
	State    domain.VerificationState
	Score    int
	Reason   string
}

func (p *Processor) Summarize(batch Batch, result domain.ImportResult) BatchSummary {
	summary := BatchSummary{BatchID: batch.ID, Source: batch.Source, Rows: len(batch.Rows), Imported: len(result.Imported), Rejected: len(result.Rejected), TotalScore: result.TotalScore, States: make(map[domain.VerificationState]int), RejectedIDs: make([]string, 0, len(result.Rejected))}
	for _, record := range result.Imported {
		summary.States[record.State]++
	}
	for _, rejected := range result.Rejected {
		if parts := strings.SplitN(rejected, ":", 2); len(parts) > 0 {
			summary.RejectedIDs = append(summary.RejectedIDs, parts[0])
		}
	}
	if summary.Imported > 0 {
		summary.AverageScore = summary.TotalScore / summary.Imported
	}
	sort.Strings(summary.RejectedIDs)
	return summary
}

func (p *Processor) DecideRows(batch Batch) []RowDecision {
	decisions := make([]RowDecision, 0, len(batch.Rows))
	for _, row := range batch.Rows {
		decision := RowDecision{RowID: row.ID, Score: domain.NormalizeScore(row.Score), State: domain.StateForScore(row.Score)}
		if err := domain.ValidateImportRow(row); err != nil {
			decision.Accepted = false
			decision.Reason = err.Error()
		} else if row.Checksum == "" {
			decision.Accepted = false
			decision.Reason = "checksum is required"
		} else {
			decision.Accepted = true
			decision.Reason = "ready for persistence"
		}
		decisions = append(decisions, decision)
	}
	return decisions
}

func StableBatchDigest(batch Batch) string {
	parts := []string{batch.ID, batch.Source, batch.StartedAt}
	for _, row := range batch.Rows {
		parts = append(parts, row.ID, row.Product, row.Version, row.Checksum, row.DownloadURL, row.Platform, fmt.Sprintf("%d", row.Score), row.Owner, row.Source)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

func RenderSummary(summary BatchSummary) string {
	states := make([]string, 0, len(summary.States))
	for state, count := range summary.States {
		states = append(states, fmt.Sprintf("%s=%d", state, count))
	}
	sort.Strings(states)
	return fmt.Sprintf("batch=%s source=%s rows=%d imported=%d rejected=%d total_score=%d average_score=%d states=%s rejected_ids=%s", summary.BatchID, summary.Source, summary.Rows, summary.Imported, summary.Rejected, summary.TotalScore, summary.AverageScore, strings.Join(states, ","), strings.Join(summary.RejectedIDs, ","))
}

func (p *Processor) ImportAndSummarize(batch Batch) (domain.ImportResult, BatchSummary, error) {
	result, err := p.Import(batch)
	if err != nil {
		return domain.ImportResult{}, BatchSummary{}, err
	}
	return result, p.Summarize(batch, result), nil
}

func MergeWarnings(result domain.ImportResult, extra ...string) domain.ImportResult {
	merged := result
	merged.Warnings = append(append([]string(nil), result.Warnings...), extra...)
	return merged
}
