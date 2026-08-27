package domain

import (
	"errors"
	"fmt"
)

type RecordStatus string

const (
	StatusDraft    RecordStatus = "draft"
	StatusPending  RecordStatus = "pending_review"
	StatusApproved RecordStatus = "approved"
	StatusRejected RecordStatus = "rejected"
	StatusArchived RecordStatus = "archived"
)

type VerificationState string

const (
	VerificationUnknown VerificationState = "unknown"
	VerificationPass    VerificationState = "pass"
	VerificationWarn    VerificationState = "warn"
	VerificationFail    VerificationState = "fail"
)

type Record struct {
	ID          string            `json:"id"`
	Product     string            `json:"product"`
	Version     string            `json:"version"`
	Checksum    string            `json:"checksum"`
	DownloadURL string            `json:"download_url"`
	Platform    string            `json:"platform"`
	Score       int               `json:"score"`
	State       VerificationState `json:"state"`
	Status      RecordStatus      `json:"status"`
	Owner       string            `json:"owner"`
	Notes       string            `json:"notes"`
	Revision    int               `json:"revision"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	ReviewedAt  string            `json:"reviewed_at"`
	ArchivedAt  string            `json:"archived_at"`
}

type AuditEvent struct {
	ID       string `json:"id"`
	RecordID string `json:"record_id"`
	Action   string `json:"action"`
	Actor    string `json:"actor"`
	Message  string `json:"message"`
	At       string `json:"at"`
	Revision int    `json:"revision"`
}

type Workflow struct {
	ID        string `json:"id"`
	RecordID  string `json:"record_id"`
	Name      string `json:"name"`
	Stage     string `json:"stage"`
	StartedAt string `json:"started_at"`
	UpdatedAt string `json:"updated_at"`
	Owner     string `json:"owner"`
}

type Attachment struct {
	ID        string `json:"id"`
	RecordID  string `json:"record_id"`
	Name      string `json:"name"`
	MediaType string `json:"media_type"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
}

type ReviewDecision struct {
	RecordID string
	Approved bool
	Actor    string
	Reason   string
	At       string
}

type ImportRow struct {
	ID          string
	Product     string
	Version     string
	Checksum    string
	DownloadURL string
	Platform    string
	Score       int
	Owner       string
	Source      string
}

type ImportResult struct {
	BatchID    string
	Imported   []Record
	Rejected   []string
	Warnings   []string
	TotalScore int
}

var (
	ErrNotFound      = errors.New("record not found")
	ErrConflict      = errors.New("revision conflict")
	ErrInvalidState  = errors.New("invalid record state")
	ErrValidation    = errors.New("validation failed")
	ErrAlreadyExists = errors.New("record already exists")
	ErrUnauthorized  = errors.New("actor is not authorized")
)

func (r Record) Summary() string {
	return fmt.Sprintf("%s %s/%s score=%d state=%s status=%s", r.ID, r.Product, r.Version, r.Score, r.State, r.Status)
}

func (r Record) IsMutable() bool {
	return r.Status == StatusDraft || r.Status == StatusRejected
}

func (r Record) IsVisible() bool {
	return r.Status == StatusApproved || r.Status == StatusArchived
}

func (r Record) Clone() Record {
	return r
}
