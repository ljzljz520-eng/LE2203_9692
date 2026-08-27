package catalog

import (
	"fmt"

	"firmwarehub/internal/domain"
	"firmwarehub/internal/store"
)

type Service struct {
	store *store.Store
}

type RegisterRequest struct {
	ID          string
	Product     string
	Version     string
	Checksum    string
	DownloadURL string
	Platform    string
	Owner       string
	Score       int
	At          string
}

type ChangeRequest struct {
	Checksum    string
	DownloadURL string
	Notes       string
	Score       int
	ExpectedRev int
	At          string
	Actor       string
}

func NewService(s *store.Store) *Service {
	return &Service{store: s}
}

func (s *Service) Register(req RegisterRequest) (domain.Record, error) {
	record := domain.StartRecord(req.ID, req.Product, req.Version, req.Checksum, req.DownloadURL, req.Platform, req.Owner, req.At, req.Score)
	if !domain.IsSupportedPlatform(record.Platform) {
		return domain.Record{}, fmt.Errorf("%w: unsupported platform", domain.ErrValidation)
	}
	if err := domain.ValidateRecord(record); err != nil {
		return domain.Record{}, err
	}
	if _, err := s.store.GetRecord(record.ID); err == nil {
		return domain.Record{}, domain.ErrAlreadyExists
	}
	if err := s.store.PutRecord(record); err != nil {
		return domain.Record{}, err
	}
	if err := s.audit(record, "registered", req.Owner, "record registered", req.At); err != nil {
		return domain.Record{}, err
	}
	return record, nil
}

func (s *Service) Get(id string) (domain.Record, error) {
	return s.store.GetRecord(id)
}

func (s *Service) Change(id string, req ChangeRequest) (domain.Record, error) {
	record, err := s.store.GetRecord(id)
	if err != nil {
		return domain.Record{}, err
	}
	if req.ExpectedRev != 0 && req.ExpectedRev != record.Revision {
		return domain.Record{}, domain.ErrConflict
	}
	updated, err := domain.ApplyChange(record, req.Checksum, req.DownloadURL, req.Notes, req.At, req.Score)
	if err != nil {
		return domain.Record{}, err
	}
	if err := domain.ValidateRecord(updated); err != nil {
		return domain.Record{}, err
	}
	if err := s.store.PutRecord(updated); err != nil {
		return domain.Record{}, err
	}
	if err := s.audit(updated, "changed", req.Actor, "record details changed", req.At); err != nil {
		return domain.Record{}, err
	}
	return updated, nil
}

func (s *Service) Submit(id, actor, at string) (domain.Record, error) {
	return s.transition(id, domain.StatusPending, actor, at, "submitted for review")
}

func (s *Service) TransitionForReview(id string, target domain.RecordStatus, actor, at string) (domain.Record, error) {
	message := "review decision recorded"
	if target == domain.StatusApproved {
		message = "record approved"
	} else if target == domain.StatusRejected {
		message = "record rejected"
	}
	return s.transition(id, target, actor, at, message)
}

func (s *Service) transition(id string, target domain.RecordStatus, actor, at, message string) (domain.Record, error) {
	record, err := s.store.GetRecord(id)
	if err != nil {
		return domain.Record{}, err
	}
	updated, err := domain.Transition(record, target, at)
	if err != nil {
		return domain.Record{}, err
	}
	if err := s.store.PutRecord(updated); err != nil {
		return domain.Record{}, err
	}
	if err := s.audit(updated, string(target), actor, message, at); err != nil {
		return domain.Record{}, err
	}
	return updated, nil
}

func (s *Service) audit(record domain.Record, action, actor, message, at string) error {
	event := domain.AuditEvent{ID: fmt.Sprintf("%s-%02d-%s", record.ID, record.Revision, action), RecordID: record.ID, Action: action, Actor: actor, Message: message, At: at, Revision: record.Revision}
	return s.store.AppendAudit(event)
}
