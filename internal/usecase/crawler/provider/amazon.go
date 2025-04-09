package provider

import (
	"akira/internal/entity"
	"context"
	"fmt"
	"regexp"
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
		colly.AllowedDomains("www.amazon.com.br", "amazon.com.br"),
		colly.MaxDepth(3),
	)
    c.OnRequest(func(r *colly.Request) {
        fmt.Printf("Visiting: %s\n", r.URL)
    })
	c.OnError(func(r *colly.Response, err error) {
        fmt.Printf("Error on %s: %s\n", r.Request.URL, err)
    })
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*amazon.*",
		Parallelism: p.opts.MaxConcurrency,
		Delay:       p.opts.RequestInterval,
		RandomDelay: 2 * time.Second,
	})
	p.collector = c
	return nil
}

func (p *AmazonProvider) Fetch(ctx context.Context, searchTerms []string) ([]entity.CrawledResult, error) {
	allResults := make([]entity.CrawledResult, 0)
	for _, term := range searchTerms {
		results, err := p.searchTerm(ctx, term)
		if err != nil {
			fmt.Printf("Warning: Initial search for '%s' failed: %v\n", term, err)
			continue
		}
		allResults = append(allResults, results...)
		select {
        case <-time.After(3 * time.Second):
            // Continue after delay
        case <-ctx.Done():
            return p.processResults(allResults), nil
        }
	}
	series := p.identifySeries(allResults)
	for seriesName, info := range series {
		if info.coverage >= 0.5 && info.count >= 5 {
			continue
		}
		maxVolumeToSearch := info.maxVolume
        if info.isMangaSeries && maxVolumeToSearch > 10 {
            fmt.Printf("Doing targeted search for manga series: %s (missing volumes)\n", seriesName)

            // Use a much simpler search strategy for manga series
            for vol := 1; vol <= maxVolumeToSearch; vol++ {
                if info.volumes[vol] {
                    continue // Skip volumes we already found
                }

                // Try just a simple search with quotes
                searchQuery := fmt.Sprintf(`"%s" vol %d`, seriesName, vol)

                // Add delay between searches (crucial for Amazon)
                select {
                case <-time.After(3 * time.Second):
                    // Continue after delay
                case <-ctx.Done():
                    return p.processResults(allResults), nil
                }

                results, err := p.searchTerm(ctx, searchQuery)
                if err != nil {
                    continue
                }

                // Check if we found the right volume
                foundRightVolume := false
                for _, result := range results {
                    if result.Volume == vol {
                        foundRightVolume = true
                        break
                    }
                }

                if foundRightVolume {
                    allResults = append(allResults, results...)
                    fmt.Printf("Found volume %d for %s\n", vol, seriesName)
                }
            }
        }
	}
	return p.processResults(allResults), nil
}

type seriesInfo struct {
	volumes  map[int]bool
	maxVolume int
	count int
	coverage float64
	isMangaSeries bool
}

func (p *AmazonProvider) identifySeries(results []entity.CrawledResult) map[string]seriesInfo {
	series := make(map[string]seriesInfo)
	for _, result := range results {
		if result.Volume <= 0 {
			continue
		}
		seriesTitle := entity.ExtractSeriesTitle(result.Title)
		if seriesTitle == "" {
			continue
		}
		seriesTitle = strings.ToLower(seriesTitle)
		seriesTitle = strings.TrimSpace(seriesTitle)
		info, exists := series[seriesTitle]
		if !exists {
			info = seriesInfo{
				volumes:      make(map[int]bool),
				maxVolume:    0,
				count:        0,
				isMangaSeries: p.looksLikeManga(result.Title, result.Publisher),
			}
		}
		info.volumes[result.Volume] = true
		info.count++
		if result.Volume > info.maxVolume {
			info.maxVolume = result.Volume
		}
		if info.maxVolume > 0 {
			info.coverage = float64(len(info.volumes)) / float64(info.maxVolume)
		}
		series[seriesTitle] = info
	}
	return series
}

func (p *AmazonProvider) looksLikeManga(title, publisher string) bool {
	title = strings.ToLower(title)
	knownMangaPublishers := []string{"panini", "jbc", "newpop", "viz media", "kodansha"}
	for _, pub := range knownMangaPublishers {
		if publisher != "" && strings.Contains(strings.ToLower(publisher), pub) {
			return true
		}
	}
	mangaIndicators := []string{"manga", "tankobon", "shonen", "shoujo", "seinen", "kodomo"}
	for _, indicator := range mangaIndicators {
		if strings.Contains(title, indicator) {
			return true
		}
	}
	volOfPattern := regexp.MustCompile(`vol\D+\d+\D+of\D+\d+`)
	if volOfPattern.MatchString(title) {
		return true
	}
	return false
}

func (p *AmazonProvider) processResults(results []entity.CrawledResult) []entity.CrawledResult {
	results = p.deduplicateResults(results)
	return p.postProcessResults(results)
}

func (p *AmazonProvider) searchTerm(ctx context.Context, term string) ([]entity.CrawledResult, error) {
	p.pagesCrawled = 0
	done := ctx.Done()
	searchURL := fmt.Sprintf("https://www.amazon.com.br/s?k=%s&i=stripbooks", strings.ReplaceAll(term, " ", "+"))
	c := p.collector.Clone()
	resultsMutex := &sync.Mutex{}
	pageResults := make([]entity.CrawledResult, 0)

	noResultsFound := false
    c.OnHTML("div.s-no-results-result", func(e *colly.HTMLElement) {
        noResultsFound = true
    })
	c.OnHTML("form[action*='validateCaptcha']", func(e *colly.HTMLElement) {
        fmt.Printf("CAPTCHA detected when searching for: %s\n", term)
    })

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

	err := c.Visit(searchURL)
	if err != nil {
		fmt.Println("Error visiting search URL:", err)
		return nil, err
	}

	c.Wait()
	if noResultsFound {
        fmt.Printf("No results found for: %s\n", term)
    }

	return pageResults, nil
}

func (p *AmazonProvider) deduplicateResults(results []entity.CrawledResult) []entity.CrawledResult {
	for i := range results {
		if results[i].Metadata == nil {
			results[i].Metadata = make(map[string]string)
		}
		results[i].Metadata["normalized_title"] = strings.ToLower(
			entity.ExtractSeriesTitle(results[i].Title),
		)
	}
	volMap := make(map[string]map[int][]entity.CrawledResult)
	for _, result := range results {
		seriesTitle := result.Metadata["normalized_title"]
		if result.Volume <= 0 || seriesTitle == "" {
			continue
		}
		if _, exists := volMap[seriesTitle]; !exists {
			volMap[seriesTitle] = make(map[int][]entity.CrawledResult)
		}
		volMap[seriesTitle][result.Volume] = append(volMap[seriesTitle][result.Volume], result)
	}

	unique := make([]entity.CrawledResult, 0)
	for _, vmap := range volMap {
		for _, candidates := range vmap {
			sort.Slice(candidates, func(i, j int) bool {
				if candidates[i].Price > 0 && candidates[j].Price == 0 {
					return true
				}
				if candidates[i].Price == 0 && candidates[j].Price > 0 {
					return false
				}
				if candidates[i].CoverImage != "" && candidates[j].CoverImage == "" {
					return true
				}
				if candidates[i].CoverImage == "" && candidates[j].CoverImage != "" {
					return false
				}
				if candidates[i].Description != "" && candidates[j].Description == "" {
					return true
				}
				return len(candidates[i].Title) < len(candidates[j].Title)
			})
			if len(candidates) > 0 {
				unique = append(unique, candidates[0])
			}
		}
	}
	for _, result := range results {
		if result.Volume <= 0 {
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
	for signature, group := range seriesGroups {
		sort.Slice(group, func(i, j int) bool {
			return group[i].Volume < group[j].Volume
		})
		if len(group) > 3 {
			minVol, maxVol := group[0].Volume, group[len(group)-1].Volume
			if minVol == 0 {
				minVol = 1
			}
			missingCount := 0
			foundedVolumes := make(map[int]bool)
			for _, result := range group {
				if result.Volume > 0 {
					foundedVolumes[result.Volume] = true
				}
			}
			for vol := minVol; vol <= maxVol; vol++ {
				if !foundedVolumes[vol] {
					missingCount++
				}
			}
			completeness := 100.0
			if maxVol > 0 {
				completeness = float64(len(foundedVolumes)) / float64(maxVol) * 100
			}
			baseTitle := entity.ExtractSeriesTitle(group[0].Title)
			for i := range group {
				if group[i].Metadata == nil {
					group[i].Metadata = make(map[string]string)
				}
                group[i].Metadata["part_of_series"] = "true"
                group[i].Metadata["series_volumes_found"] = fmt.Sprintf("%d", len(group))
                group[i].Metadata["series_max_volume"] = fmt.Sprintf("%d", maxVol)
                group[i].Metadata["series_completeness"] = fmt.Sprintf("%.1f", completeness)
                group[i].Metadata["series_title"] = baseTitle
                group[i].Metadata["series_signature"] = signature
				if group[i].Publisher == "" {
					for _, other := range group {
						if other.Publisher != "" {
							group[i].Publisher = other.Publisher
							break
						}
					}
				}
			}
		}
		enhancedResults = append(enhancedResults, group...)
	}
	return enhancedResults
}

