package domain

import "fmt"

func CanTransition(from, to RecordStatus) bool {
	switch from {
	case StatusDraft:
		return to == StatusPending || to == StatusRejected
	case StatusPending:
		return to == StatusApproved || to == StatusRejected
	case StatusApproved:
		return to == StatusArchived
	case StatusRejected:
		return to == StatusDraft || to == StatusPending
	case StatusArchived:
		return false
	default:
		return false
	}
}

func Transition(r Record, to RecordStatus, at string) (Record, error) {
	if !CanTransition(r.Status, to) {
		return r, fmt.Errorf("%w: %s -> %s", ErrInvalidState, r.Status, to)
	}
	r.Status = to
	r.UpdatedAt = at
	if to == StatusApproved {
		r.ReviewedAt = at
	}
	if to == StatusArchived {
		r.ArchivedAt = at
	}
	r.Revision++
	return r, nil
}

func StartRecord(id, product, version, checksum, url, platform, owner, at string, score int) Record {
	return Record{
		ID: id, Product: product, Version: version, Checksum: checksum,
		DownloadURL: url, Platform: platform, Owner: owner, Score: NormalizeScore(score),
		State: StateForScore(score), Status: StatusDraft, Revision: 1,
		CreatedAt: at, UpdatedAt: at,
	}
}

func ApplyChange(r Record, checksum, url, notes, at string, score int) (Record, error) {
	if !r.IsMutable() {
		return r, fmt.Errorf("%w: only draft or rejected records can change", ErrInvalidState)
	}
	if checksum != "" {
		r.Checksum = checksum
	}
	if url != "" {
		r.DownloadURL = url
	}
	r.Notes = notes
	r.Score = NormalizeScore(score)
	r.State = StateForScore(r.Score)
	r.UpdatedAt = at
	r.Revision++
	return r, nil
}
