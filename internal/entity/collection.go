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
	SyncStatusPending  SyncStatus = "pending"
	SyncStatusFetching SyncStatus = "fetching"
	SyncStatusSynced   SyncStatus = "synced"
	SyncStatusFailed   SyncStatus = "failed"
)

type ScrapingSite struct {
	ID        string
	Name      string
	URL       map[string]string
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ScrapingOptions struct {
	AutoSync        bool
	TrackPrice      bool
	TrackNewVolumes bool
	TrackReviews    bool
}

type Collection struct {
	ID              string
	Name            string
	Edition         string
	Slug            string
	UserID          string
	Author          []string
	Publisher       string
	Tags            []string
	Metadata        map[string]string
	ReleaseStatus   ReleaseStatus
	SyncStatus      SyncStatus
	ScrapingSites   []ScrapingSite
	TotalVolumes    int
	ScrapingOptions ScrapingOptions
	Language        string
	LastSync        time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ScrapingReview struct {
	ID        string
	VolumeID  string
	Author    string
	Title     string
	Content   string
	Rating    float64
	Date      time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Book struct {
	ID          string
	Name        string
	Edition     string
	Description string
	Slug        string
	CoverImage  string
	PageCount   int
	Volume      *int
	Rating      float64
	Reviews     []ScrapingReview
	Publisher   string
	Author      []string
	UserID      string
	ISBN        string
	Tags        []string
	Metadata    map[string]string
	Language    string
	LastSync    time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CollectionBook struct {
	CollectionID string
	BookID       string
	CreatedAt    time.Time
}

type CreateCollectionRequest struct {
	Name            string
	Edition         string
	Author          []string
	Publisher       string
	Tags            []string
	Metadata        map[string]string
	ScrapingSites   []ScrapingSite
	ScrapingOptions ScrapingOptions
	Language        string
}

func (r *CreateCollectionRequest) Validate() error {
	var e RequestError
	if r.Name == "" {
		e = e.Add("name", ErrCollectionNameInvalid.Error())
	}
	if len(r.Name) > 255 {
		e = e.Add("name", ErrCollectionNameTooLong.Error())
	}
	return e
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
	scrapingSites []ScrapingSite,
	scrapingOptions ScrapingOptions,
) *Collection {
	return &Collection{
		ID:              NewID(),
		Name:            name,
		Edition:         edition,
		Slug:            slug,
		UserID:          userID,
		Author:          author,
		Publisher:       publisher,
		Tags:            tags,
		Metadata:        metadata,
		ReleaseStatus:   ReleaseStatusOnGoing,
		SyncStatus:      SyncStatusPending,
		ScrapingSites:   scrapingSites,
		TotalVolumes:    0,
		ScrapingOptions: scrapingOptions,
		Language:        language,
		LastSync:        time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

type CollectionService interface {
	CreateCollection(userID string, req CreateCollectionRequest) (*Collection, error)
}

type CollectionRepository interface {
	CreateCollection(collection *Collection) error
	FindCollectionBySlug(userID, slug string) (*Collection, error)
}
