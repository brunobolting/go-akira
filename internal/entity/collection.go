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
