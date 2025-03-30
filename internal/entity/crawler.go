package entity

import (
	"context"
	"time"
)

const (
	EventCrawlerStarted     EventType = "crawler:started"
	EventCrawlerCompleted   EventType = "crawler:completed"
	EventCrawlerFailed      EventType = "crawler:failed"
	EventCrawlerItemFounded EventType = "crawler:item-founded"
)

type CrawledReview struct {
	Title   string  `json:"title"`
	Author  string  `json:"author"`
	Content string  `json:"content"`
	Rating  float64 `json:"rating"`
	Date    string  `json:"date"`
}

type CrawledResult struct {
	Title       string            `json:"title"`
	Volume      int               `json:"volume"`
	ISBN        string            `json:"isbn"`
	Price       float64           `json:"price"`
	CoverImage  string            `json:"cover_image"`
	URL         string            `json:"url"`
	Description string            `json:"description"`
	Publisher   string            `json:"publisher"`
	Author      []string          `json:"author"`
	Source      string            `json:"source"`
	Tags        []string          `json:"tags"`
	Rating      float64           `json:"rating"`
	Reviews     []CrawledReview   `json:"reviews"`
	ReleasedAt  string            `json:"released_at"`
	Metadata    map[string]string `json:"metadata"`
	Language    string            `json:"language"`
}

type CrawlerOptions struct {
	MaxPages        int
	Timeout         time.Duration
	MaxConcurrency  int
	RequestInterval time.Duration
	SiteOptions     map[string]map[string]any
}

type SiteProvider interface {
	SiteName() string
	Setup(opts CrawlerOptions) error
	Fetch(ctx context.Context, searchTerms []string) ([]CrawledResult, error)
}

type CrawlerRequest struct {
	CollectionID string
	SearchTerms  []string
	Sites        []string
	Opts         CrawlerOptions
}

type CrawlerService interface {
	FetchCollection(ctx context.Context, req CrawlerRequest) error
	GetStatus(collectionID string) (SyncStatus, error)
	CancelFetch(collectionID string) error
}

type CrawlerConsumer interface {
	ConsumeEvents()
	Shutdown() error
}
