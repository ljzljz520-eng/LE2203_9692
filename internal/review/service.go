package review

import (
	"fmt"

	"firmwarehub/internal/catalog"
	"firmwarehub/internal/domain"
)

type Service struct {
	catalog *catalog.Service
	policy  Policy
}

func NewService(c *catalog.Service, policy Policy) *Service {
	return &Service{catalog: c, policy: policy}
}

func (s *Service) Submit(recordID, actor, at string) (domain.Record, error) {
	return s.catalog.Submit(recordID, actor, at)
}

func (s *Service) Decide(decision domain.ReviewDecision) (domain.Record, error) {
	record, err := s.catalog.Get(decision.RecordID)
	if err != nil {
		return domain.Record{}, err
	}
	if record.Status != domain.StatusPending {
		return domain.Record{}, fmt.Errorf("%w: review requires pending status", domain.ErrInvalidState)
	}
	if err := s.policy.Validate(decision, record); err != nil {
		return domain.Record{}, err
	}
	target := domain.StatusRejected
	if decision.Approved {
		target = domain.StatusApproved
	}
	updated, err := s.catalogTransition(decision.RecordID, target, decision.Actor, decision.At)
	if err != nil {
		return domain.Record{}, err
	}
	return updated, nil
}

func (s *Service) catalogTransition(id string, target domain.RecordStatus, actor, at string) (domain.Record, error) {
	if target == domain.StatusApproved {
		return s.catalogApprove(id, actor, at)
	}
	return s.catalogReject(id, actor, at)
}

func (s *Service) catalogApprove(id, actor, at string) (domain.Record, error) {
	return invokeTransition(s.catalog, id, domain.StatusApproved, actor, at)
}

func (s *Service) catalogReject(id, actor, at string) (domain.Record, error) {
	return invokeTransition(s.catalog, id, domain.StatusRejected, actor, at)
}

func invokeTransition(c *catalog.Service, id string, target domain.RecordStatus, actor, at string) (domain.Record, error) {
	return c.TransitionForReview(id, target, actor, at)
}

func (s *Service) Explain(recordID string) ([]string, error) {
	record, err := s.catalog.Get(recordID)
	if err != nil {
		return nil, err
	}
	return domain.ExplainScore(record), nil
}
