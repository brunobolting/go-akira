package entity

import "time"

type ReleaseStatus string

const (
	ReleaseStatusOnGoing   ReleaseStatus = "ongoing"
	ReleaseStatusCompleted ReleaseStatus = "completed"
	ReleaseStatusHiatus    ReleaseStatus = "hiatus"
	ReleaseStatusCancelled ReleaseStatus = "cancelled"
)

type SyncStatus string

const (
	SyncStatusNotFound SyncStatus = "not_found"
	SyncStatusPending  SyncStatus = "pending"
	SyncStatusFetching SyncStatus = "fetching"
	SyncStatusSynced   SyncStatus = "synced"
	SyncStatusFailed   SyncStatus = "failed"
)

type SyncSources []string

type SyncOptions struct {
	AutoSync        bool
	TrackPrice      bool
	TrackNewVolumes bool
	TrackReviews    bool
}

type Collection struct {
	ID             string
	Name           string
	Edition        string
	Slug           string
	UserID         string
	Author         []string
	Publisher      string
	Tags           []string
	Metadata       map[string]string
	ReleaseStatus  ReleaseStatus
	SyncStatus     SyncStatus
	SyncSources    SyncSources
	TotalVolumes   int
	CrawlerOptions SyncOptions
	Language       string
	LastSync       time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CollectionBook struct {
	CollectionID string
	BookID       string
	CreatedAt    time.Time
}

type CreateCollectionRequest struct {
	Name           string
	Edition        string
	Author         []string
	Publisher      string
	Tags           []string
	Metadata       map[string]string
	SyncSources    SyncSources
	CrawlerOptions SyncOptions
	Language       string
}

func (r *CreateCollectionRequest) Validate() error {
	var e RequestError
	if r.Name == "" {
		e = e.Add("name", ErrCollectionNameInvalid.Error())
	}
	if len(r.Name) > 255 {
		e = e.Add("name", ErrCollectionNameTooLong.Error())
	}
	if e.HasError() {
		return e
	}
	return nil
}

func NewCollection(
	userID string,
	name string,
	edition string,
	slug string,
	author []string,
	publisher string,
	language string,
	tags []string,
	metadata map[string]string,
	syncSources SyncSources,
	opts SyncOptions,
) *Collection {
	return &Collection{
		ID:             NewID(),
		Name:           name,
		Edition:        edition,
		Slug:           slug,
		UserID:         userID,
		Author:         author,
		Publisher:      publisher,
		Tags:           tags,
		Metadata:       metadata,
		ReleaseStatus:  ReleaseStatusOnGoing,
		SyncStatus:     SyncStatusPending,
		SyncSources:    syncSources,
		TotalVolumes:   0,
		CrawlerOptions: opts,
		Language:       language,
		LastSync:       time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

type CollectionService interface {
	CreateCollection(userID string, req CreateCollectionRequest) (*Collection, error)
	// SyncCollection(collectionID string, opts CrawlerOptions) error
}

type CollectionRepository interface {
	CreateCollection(collection *Collection) error
	FindCollectionBySlug(userID, slug string) (*Collection, error)
}
