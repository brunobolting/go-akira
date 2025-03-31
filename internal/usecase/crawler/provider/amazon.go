package provider

import (
	"akira/internal/entity"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

var _ entity.SiteProvider = (*AmazonProvider)(nil)

type AmazonProvider struct {
	collector    *colly.Collector
	opts         entity.CrawlerOptions
	pagesCrawled int
}

func NewAmazonProvider() entity.SiteProvider {
	return &AmazonProvider{}
}

func (p *AmazonProvider) SiteName() string {
	return "amazon"
}

func (p *AmazonProvider) Setup(opts entity.CrawlerOptions) error {
	p.opts = opts
	if opts.MaxPages < 5 {
		opts.MaxPages = 5
	}
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.amazon.com.br"),
		colly.MaxDepth(3),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*amazon.*",
		Parallelism: p.opts.MaxConcurrency,
		Delay:       p.opts.RequestInterval,
	})
	p.collector = c
	return nil
}

func (p *AmazonProvider) Fetch(ctx context.Context, searchTerms []string) ([]entity.CrawledResult, error) {
	originalTerms := searchTerms
	enhancedTerms := []string{}
	for _, term := range searchTerms {
		enhancedTerms = append(enhancedTerms, term)
		enhancedTerms = append(enhancedTerms, term+" volume")
		enhancedTerms = append(enhancedTerms, term+" vol")
	}

	allResults := make([]entity.CrawledResult, 0)

	for _, term := range enhancedTerms {
		results, err := p.searchTerm(ctx, term)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	maxVolume := 0
	foundVolumes := make(map[int]bool)
	for _, result := range allResults {
		if result.Volume > 0 {
			foundVolumes[result.Volume] = true
			if result.Volume > maxVolume {
				maxVolume = result.Volume
			}
		}
	}

	if maxVolume > 3 && len(foundVolumes) < (maxVolume/2) {
		for i := 1; i <= maxVolume; i++ {
			if !foundVolumes[i] {
				for _, baseTerm := range originalTerms {
					volumeSearchTerm := fmt.Sprintf("%s volume %d", baseTerm, i)
					results, err := p.searchTerm(ctx, volumeSearchTerm)
					if err != nil {
						continue
					}
					if len(results) > 0 {
						allResults = append(allResults, results...)
						foundVolumes[i] = true
						break
					}
				}
			}
		}
	}

	return p.deduplicateResults(allResults), nil
}

func (p *AmazonProvider) searchTerm(ctx context.Context, term string) ([]entity.CrawledResult, error) {
	done := ctx.Done()
	searchURL := fmt.Sprintf("https://www.amazon.com.br/s?k=%s&i=stripbooks", strings.ReplaceAll(term, " ", "+"))
	c := p.collector.Clone()
	resultsMutex := &sync.Mutex{}
	pageResults := make([]entity.CrawledResult, 0)

	c.OnHTML("div.s-result-item", func(e *colly.HTMLElement) {
		select {
		case <-done:
			return
		default:
		}
		asin := e.Attr("data-asin")
		if asin == "" {
			return
		}
		title := strings.TrimSpace(e.ChildText("span.a-size-base-plus"))
		if title == "" {
			title = strings.TrimSpace(e.ChildText("h2 span"))
		}
		if title == "" {
			return
		}
		productURL := fmt.Sprintf("https://www.amazon.com.br/dp/%s", asin)

		coverURL := e.ChildAttr("img.s-image", "src")

		priceStr := e.ChildText("span.a-price-whole")
		priceStr = strings.ReplaceAll(priceStr, ".", "")
		priceStr = strings.ReplaceAll(priceStr, ",", ".")
		priceStr = strings.TrimSpace(priceStr)
		var price float64
		if priceStr != "" {
			price, _ = strconv.ParseFloat(priceStr, 64)
		}

		volume := entity.ExtractVolumeNumber(title)

		result := entity.CrawledResult{
			Title:      title,
			Volume:     volume,
			ISBN:       asin,
			Price:      price,
			CoverImage: coverURL,
			URL:        productURL,
			Source:     p.SiteName(),
			Metadata: map[string]string{
				"asin": asin,
			},
		}

		if price == 0 {
			productCollector := p.collector.Clone()
			productCollector.OnHTML("span#price", func(e *colly.HTMLElement) {
				priceStr := strings.TrimSpace(e.Text)
				priceStr = strings.ReplaceAll(priceStr, "R$", "")
				priceStr = strings.ReplaceAll(priceStr, ".", "")
				priceStr = strings.ReplaceAll(priceStr, ",", ".")
				priceStr = strings.TrimSpace(priceStr)
				if priceStr != "" {
					if p, err := strconv.ParseFloat(priceStr, 64); err == nil {
						result.Price = p
					}
				}

			})

			productCollector.OnHTML("div#bookDescription_feature_div", func(e *colly.HTMLElement) {
				result.Description = strings.TrimSpace(e.Text)
			})

			productCollector.OnHTML("div#detailBullets_feature_div ul li", func(e *colly.HTMLElement) {
				text := strings.TrimSpace(e.Text)

				if strings.Contains(text, "ISBN") {
					parts := strings.Split(text, ":")
					if len(parts) > 1 {
						isbn := strings.TrimSpace(parts[1])
						isbn = strings.ReplaceAll(isbn, "-", "")
						isbn = strings.ReplaceAll(isbn, " ", "")
						isbn = strings.ReplaceAll(isbn, "\n", "")
						result.ISBN = isbn
					}
				}

				if strings.Contains(text, "Editora") {
					parts := strings.Split(text, ":")
					if len(parts) > 1 {
						publisher := strings.TrimSpace(parts[1])
						publisher = strings.ReplaceAll(publisher, "-", "")
						publisher = strings.ReplaceAll(publisher, " ", "")
						publisher = strings.ReplaceAll(publisher, "\n", "")
						result.Publisher = publisher
					}
				}

				if strings.Contains(text, "publicação") {
					parts := strings.Split(text, ":")
					if len(parts) > 1 {
						dateStr := strings.TrimSpace(parts[1])
						if t, err := time.Parse("02 de janeiro de 2006", dateStr); err == nil {
							result.ReleasedAt = t.Format("2006-01-02")
						}
					}
				}
			})

			productCollector.Visit(productURL)
		}

		resultsMutex.Lock()
		pageResults = append(pageResults, result)
		resultsMutex.Unlock()
	})

	pageCount := 0
	c.OnHTML("ul.a-pagination li.a-disabled", func(e *colly.HTMLElement) {
		if pageCount == 0 {
			pageCountStr := e.Text
			if p, err := strconv.Atoi(pageCountStr); err == nil && p > 0 {
				pageCount = p
			}
		}
	})
	c.OnHTML("ul.a-pagination li.a-last a", func(e *colly.HTMLElement) {
		select {
		case <-done:
			return
		default:
		}
		if p.opts.MaxPages > 0 && p.pagesCrawled < p.opts.MaxPages {
			p.pagesCrawled++
			nextPage := e.Attr("href")
			if nextPage != "" {
				c.Visit(e.Request.AbsoluteURL(nextPage))
			}
		}
	})

	c.Visit(searchURL)

	return pageResults, nil
}

func (p *AmazonProvider) deduplicateResults(results []entity.CrawledResult) []entity.CrawledResult {
	seen := make(map[string]bool)
	unique := make([]entity.CrawledResult, 0)
	for _, result := range results {
		identifier := result.ISBN
		if identifier == "" {
			identifier = fmt.Sprintf("%s_vol%d", result.Title, result.Volume)
		}
		if !seen[identifier] {
			seen[identifier] = true
			unique = append(unique, result)
		}
	}
	return unique
}

func (p *AmazonProvider) postProcessResults(results []entity.CrawledResult) []entity.CrawledResult {
	seriesGroups := make (map[string][]entity.CrawledResult)
	for _, result := range results {
		seriesSignature := entity.CalculateSeriesSignature(result)
		seriesGroups[seriesSignature] = append(seriesGroups[seriesSignature], result)
	}
	enhancedResults := make([]entity.CrawledResult, 0)
	for _, group := range seriesGroups {
		sort.Slice(group, func(i, j int) bool {
			return group[i].Volume < group[j].Volume
		})
		if len(group) > 3 {
			for i := range group {
				group[i].Metadata["part_of_series"] = "true"
				group[i].Metadata["series_volumes_found"] = fmt.Sprintf("%d", len(group))
				baseTitle := entity.ExtractSeriesTitle(group[i].Title)
				group[i].Metadata["series_title"] = baseTitle
			}
		}
		enhancedResults = append(enhancedResults, group...)
	}
	return enhancedResults
}

