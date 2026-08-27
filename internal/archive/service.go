package archive

import (
	"fmt"

	"firmwarehub/internal/catalog"
	"firmwarehub/internal/domain"
)

type Service struct {
	catalog *catalog.Service
}

func NewService(c *catalog.Service) *Service {
	return &Service{catalog: c}
}

func (s *Service) Archive(recordID, actor, at string) (domain.Record, error) {
	record, err := s.catalog.Get(recordID)
	if err != nil {
		return domain.Record{}, err
	}
	if record.Status != domain.StatusApproved {
		return domain.Record{}, fmt.Errorf("%w: only approved records can be archived", domain.ErrInvalidState)
	}
	return s.catalog.TransitionForReview(recordID, domain.StatusArchived, actor, at)
}

func (s *Service) Restore(recordID, actor, at string) (domain.Record, error) {
	record, err := s.catalog.Get(recordID)
	if err != nil {
		return domain.Record{}, err
	}
	if record.Status != domain.StatusArchived {
		return domain.Record{}, fmt.Errorf("%w: only archived records can be restored", domain.ErrInvalidState)
	}
	return domain.Record{}, fmt.Errorf("%w: archived records are immutable", domain.ErrInvalidState)
}
