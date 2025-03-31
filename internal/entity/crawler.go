package entity

import (
	"context"
	"regexp"
	"strconv"
	"strings"
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

func ExtractVolumeNumber(title string) int {
	title = strings.ToLower(title)
	patterns := []string{
		`vol\D*(\d+)`,            // "Vol. 3", "Vol 3", "Vol-3"
		`volume\D*(\d+)`,         // "Volume 3", "Volume: 3"
		`\b(\d+)\D*vol`,          // "3 Vol", "3º Vol"
		`#\s*(\d+)`,              // "#3", "# 3"
		`\btomo\D*(\d+)`,         // "Tomo 3" (Spanish/Portuguese)
		`\D(\d{1,2})\D*$`,        // Title ending with a number
		`\D(\d{1,2})\D.*edition`, // "3rd edition" type patterns
	}
	for _, pattern := range patterns {
		rx := regexp.MustCompile(pattern)
		matches := rx.FindStringSubmatch(title)
		if len(matches) > 1 {
			if vol, err := strconv.Atoi(matches[1]); err == nil {
				return vol
			}
		}
	}
	return 0
}

func ExtractSeriesTitle(title string) string {
	patterns := []string{
		`\bvol\.?\s*\d+`,
		`\bvolume\s*\d+`,
		`\btomo\s*\d+`,
		`\#\d+`,
		`\d+ª?\s*edição`,
		`\bedition\b`,
	}
	cleanTitle := title
	for _, pattern := range patterns {
		rx := regexp.MustCompile(pattern)
		cleanTitle = rx.ReplaceAllString(cleanTitle, "")
	}
	cleanTitle = strings.Trim(cleanTitle, " -:,.")
	return cleanTitle
}

func CalculateSeriesSignature(result CrawledResult) string {
	baseTitle := ExtractSeriesTitle(result.Title)
	signature := strings.ToLower(baseTitle)
	if result.Publisher != "" {
		signature += "|" + strings.ToLower(result.Publisher)
	}
	return signature
}
