package review

import (
	"sort"
	"strings"

	"firmwarehub/internal/domain"
)

type QueueItem struct {
	Record   domain.Record
	Priority int
	Reasons  []string
}

type Queue struct {
	Items    []QueueItem
	Total    int
	Eligible int
}

func BuildQueue(records []domain.Record, policy Policy) Queue {
	items := make([]QueueItem, 0)
	for _, record := range records {
		if record.Status != domain.StatusPending {
			continue
		}
		reviewability := domain.ReviewRequirements(record, policy.MinimumScore)
		priority := queuePriority(record, reviewability)
		items = append(items, QueueItem{Record: record, Priority: priority, Reasons: append([]string(nil), reviewability.Reasons...)})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Priority == items[j].Priority {
			return items[i].Record.ID < items[j].Record.ID
		}
		return items[i].Priority > items[j].Priority
	})
	eligible := 0
	for _, item := range items {
		if len(item.Reasons) == 0 {
			eligible++
		}
	}
	return Queue{Items: items, Total: len(items), Eligible: eligible}
}

func queuePriority(record domain.Record, reviewability domain.Reviewability) int {
	priority := 100 - record.Score
	if record.State == domain.VerificationFail {
		priority += 30
	}
	if record.UpdatedAt == record.CreatedAt {
		priority += 5
	}
	if len(reviewability.Reasons) > 0 {
		priority += len(reviewability.Reasons) * 4
	}
	return priority
}

func FilterQueue(queue Queue, text string) Queue {
	needle := strings.ToLower(strings.TrimSpace(text))
	if needle == "" {
		return queue
	}
	filtered := make([]QueueItem, 0, len(queue.Items))
	for _, item := range queue.Items {
		if strings.Contains(strings.ToLower(item.Record.Summary()), needle) {
			filtered = append(filtered, item)
		}
	}
	queue.Items = filtered
	queue.Total = len(filtered)
	queue.Eligible = 0
	for _, item := range filtered {
		if len(item.Reasons) == 0 {
			queue.Eligible++
		}
	}
	return queue
}

func QueueIDs(queue Queue) []string {
	ids := make([]string, 0, len(queue.Items))
	for _, item := range queue.Items {
		ids = append(ids, item.Record.ID)
	}
	return ids
}

func (s *Service) ExplainDecision(decision domain.ReviewDecision) string {
	if decision.Approved {
		return "approval requested: " + strings.TrimSpace(decision.Reason)
	}
	return "rejection requested: " + strings.TrimSpace(decision.Reason)
}
