package flow020

import (
	"fmt"

	"firmwarehub/internal/archive"
	"firmwarehub/internal/catalog"
	"firmwarehub/internal/domain"
	"firmwarehub/internal/importer"
	"firmwarehub/internal/review"
)

type Application struct {
	Catalog  *catalog.Service
	Review   *review.Service
	Archive  *archive.Service
	Importer *importer.Processor
	Reporter *archive.Reporter
}

func NewApplication(c *catalog.Service, r *review.Service, a *archive.Service, i *importer.Processor, reporter *archive.Reporter) *Application {
	return &Application{Catalog: c, Review: r, Archive: a, Importer: i, Reporter: reporter}
}

func (a *Application) CreateReviewArchive(req catalog.RegisterRequest, reviewer, archivist, at string) (domain.Record, error) {
	record, err := a.Catalog.Register(req)
	if err != nil {
		return domain.Record{}, err
	}
	if _, err := a.Review.Submit(record.ID, reviewer, at); err != nil {
		return domain.Record{}, err
	}
	if _, err := a.Review.Decide(domain.ReviewDecision{RecordID: record.ID, Approved: true, Actor: reviewer, Reason: "checks verified", At: at}); err != nil {
		return domain.Record{}, err
	}
	return a.Archive.Archive(record.ID, archivist, at)
}

func (a *Application) SearchUpdatePublish(id string, change catalog.ChangeRequest, at string) ([]domain.Record, error) {
	if _, err := a.Catalog.Change(id, change); err != nil {
		return nil, err
	}
	if _, err := a.Review.Submit(id, change.Actor, at); err != nil {
		return nil, err
	}
	if _, err := a.Review.Decide(domain.ReviewDecision{RecordID: id, Approved: true, Actor: change.Actor, Reason: "updated artifact accepted", At: at}); err != nil {
		return nil, err
	}
	return a.Catalog.Search(catalog.SearchQuery{Text: id, Status: string(domain.StatusApproved), Limit: 20})
}

func (a *Application) ImportReport(batch importer.Batch, at string) (domain.ImportResult, archive.Report, error) {
	result, err := a.Importer.Import(batch)
	if err != nil {
		return domain.ImportResult{}, archive.Report{}, err
	}
	report, err := a.Reporter.Build(at)
	if err != nil {
		return domain.ImportResult{}, archive.Report{}, err
	}
	return result, report, nil
}

func (a *Application) RequireImportedScore(result domain.ImportResult, id string, expected int) error {
	for _, record := range result.Imported {
		if record.ID == id {
			if record.Score != expected {
				return fmt.Errorf("imported score for %s: got %d want %d", id, record.Score, expected)
			}
			return nil
		}
	}
	return fmt.Errorf("imported record %s not found", id)
}
