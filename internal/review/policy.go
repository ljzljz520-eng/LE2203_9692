package review

import (
	"fmt"
	"strings"

	"firmwarehub/internal/domain"
)

type Policy struct {
	MinimumScore int
	RequireNotes bool
	AllowedRoles map[string]bool
}

func DefaultPolicy() Policy {
	return Policy{MinimumScore: 50, RequireNotes: true, AllowedRoles: map[string]bool{"reviewer": true, "admin": true}}
}

func (p Policy) Validate(decision domain.ReviewDecision, record domain.Record) error {
	if decision.Actor == "" || !p.AllowedRoles[decision.Actor] {
		return domain.ErrUnauthorized
	}
	if decision.Approved && record.Score < p.MinimumScore {
		return fmt.Errorf("%w: score %d is below %d", domain.ErrValidation, record.Score, p.MinimumScore)
	}
	if p.RequireNotes && strings.TrimSpace(decision.Reason) == "" {
		return fmt.Errorf("%w: review reason is required", domain.ErrValidation)
	}
	return nil
}

func (p Policy) DecisionLabel(decision domain.ReviewDecision) string {
	if decision.Approved {
		return "approved"
	}
	return "rejected"
}
