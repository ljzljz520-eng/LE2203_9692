package archive

import (
	"fmt"
	"sort"
	"strings"

	"firmwarehub/internal/domain"
	"firmwarehub/internal/store"
)

type Report struct {
	GeneratedAt  string
	Total        int
	Approved     int
	Archived     int
	Rejected     int
	AverageScore int
	ByPlatform   map[string]int
}

type Reporter struct {
	store *store.Store
}

func NewReporter(s *store.Store) *Reporter {
	return &Reporter{store: s}
}

func (r *Reporter) Build(at string) (Report, error) {
	records, err := r.store.ListRecords()
	if err != nil {
		return Report{}, err
	}
	report := Report{GeneratedAt: at, Total: len(records), ByPlatform: make(map[string]int)}
	for _, record := range records {
		switch record.Status {
		case domain.StatusApproved:
			report.Approved++
		case domain.StatusArchived:
			report.Archived++
		case domain.StatusRejected:
			report.Rejected++
		}
		report.AverageScore += record.Score
		report.ByPlatform[record.Platform]++
	}
	if report.Total > 0 {
		report.AverageScore /= report.Total
	}
	return report, nil
}

func (r *Reporter) Render(report Report) string {
	platforms := make([]string, 0, len(report.ByPlatform))
	for platform := range report.ByPlatform {
		platforms = append(platforms, platform)
	}
	sort.Strings(platforms)
	parts := []string{fmt.Sprintf("generated=%s", report.GeneratedAt), fmt.Sprintf("total=%d", report.Total), fmt.Sprintf("approved=%d", report.Approved), fmt.Sprintf("archived=%d", report.Archived), fmt.Sprintf("rejected=%d", report.Rejected), fmt.Sprintf("average_score=%d", report.AverageScore)}
	for _, platform := range platforms {
		parts = append(parts, fmt.Sprintf("%s=%d", platform, report.ByPlatform[platform]))
	}
	return strings.Join(parts, " ")
}
