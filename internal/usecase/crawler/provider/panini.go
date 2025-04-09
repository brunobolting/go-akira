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

var _ entity.SiteProvider = (*PaniniProvider)(nil)

type PaniniProvider struct {
	collector    *colly.Collector
	opts         entity.CrawlerOptions
	pagesCrawled int
	foundVolumes map[int]bool
}

func NewPaniniProvider() *PaniniProvider {
	return &PaniniProvider{}
}

func (p *PaniniProvider) SiteName() string {
	return "panini"
}

func (p *PaniniProvider) Setup(opts entity.CrawlerOptions) error {
	p.opts = opts
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.AllowedDomains("www.panini.com.br", "panini.com.br"),
	)
	c.SetRequestTimeout(30 * time.Second)
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Cache-Control", "max-age=0")
		fmt.Printf("Visitando: %s\n", r.URL)
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Erro ao visitar %s: %s (Status: %d)\n", r.Request.URL, err, r.StatusCode)
	})
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*panini.*",
		Parallelism: p.opts.MaxConcurrency,
		Delay:       p.opts.RequestInterval,
		RandomDelay: 1 * time.Second,
	})

	p.collector = c
	p.foundVolumes = make(map[int]bool)
	return nil
}

func (p *PaniniProvider) Fetch(ctx context.Context, searchTerms []string) ([]entity.CrawledResult, error) {
	allResults := make([]entity.CrawledResult, 0)
	visitedURLs := make(map[string]bool)
	visitedMutex := &sync.Mutex{}
	resultsMutex := &sync.Mutex{}

	// Visitar página inicial para configurar cookies
	err := p.collector.Visit("https://panini.com.br/")
	if err != nil {
		fmt.Printf("Aviso: Não foi possível visitar a página principal da Panini: %v\n", err)
	}
	time.Sleep(1 * time.Second)

	var wg sync.WaitGroup

	for _, baseSearchTerm := range searchTerms {
		select {
		case <-ctx.Done():
			return p.processResults(allResults), nil
		default:
			wg.Add(1)
			go func(term string) {
				defer wg.Done()

				// Tenta encontrar o volume 1 ou volume inicial da série
				initialVolumesFound := p.findInitialVolumes(ctx, term, &allResults, visitedURLs, visitedMutex, resultsMutex)

				// Se não encontramos volume específico, fazemos uma busca genérica
				if !initialVolumesFound {
					genericResults := p.searchGenericTerm(ctx, term, visitedURLs, visitedMutex)
					if len(genericResults) > 0 {
						resultsMutex.Lock()
						allResults = append(allResults, genericResults...)
						resultsMutex.Unlock()
					}
				}
			}(baseSearchTerm)

			// Adiciona um pequeno atraso entre termos de busca
			select {
			case <-time.After(2 * time.Second):
			case <-ctx.Done():
				break
			}
		}
	}

	// Aguarda todas as buscas terminarem
	wgDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(wgDone)
	}()

	// Espera pela conclusão ou timeout
	select {
	case <-wgDone:
		// Buscas concluídas normalmente
	case <-ctx.Done():
		fmt.Println("Contexto cancelado, retornando resultados parciais")
	}

	return p.processResults(allResults), nil
}

func (p *PaniniProvider) findInitialVolumes(
	ctx context.Context,
	searchTerm string,
	allResults *[]entity.CrawledResult,
	visitedURLs map[string]bool,
	visitedMutex *sync.Mutex,
	resultsMutex *sync.Mutex,
) bool {
	// Formatos de busca para tentar encontrar o volume 1
	searchFormats := []string{
		fmt.Sprintf("%s vol 1", searchTerm),
		fmt.Sprintf("%s volume 1", searchTerm),
		fmt.Sprintf("%s vol. 1", searchTerm),
		fmt.Sprintf("%s #1", searchTerm),
	}

	foundVolumes := false

	for _, term := range searchFormats {
		select {
		case <-ctx.Done():
			return foundVolumes
		default:
			searchURL := fmt.Sprintf("https://panini.com.br/catalogsearch/result/?q=%s", strings.ReplaceAll(term, " ", "+"))
			fmt.Printf("Procurando volume inicial: %s\n", term)

			c := p.collector.Clone()
			productURLs := make([]string, 0)

			// Encontra resultados na página de busca
			c.OnHTML("div.products div.product-item-info", func(e *colly.HTMLElement) {
				productURL := e.ChildAttr("a.product-item-photo", "href")
				if productURL == "" {
					productURL = e.ChildAttr("a.product-item-link", "href")
				}

				if productURL != "" {
					visitedMutex.Lock()
					if !visitedURLs[productURL] {
						visitedURLs[productURL] = true
						productURLs = append(productURLs, productURL)
					}
					visitedMutex.Unlock()
				}
			})

			// Verifica se não encontrou resultados
			c.OnHTML("div.message.notice", func(e *colly.HTMLElement) {
				if strings.Contains(e.Text, "não encontramos") {
					fmt.Printf("Nenhum resultado encontrado para: %s\n", term)
				}
			})

			err := c.Visit(searchURL)
			if err != nil {
				fmt.Printf("Erro ao visitar página de busca %s: %v\n", searchURL, err)
				continue
			}
			c.Wait()

			// Processa cada URL de produto encontrada
			for _, productURL := range productURLs {
				results := p.processProductAndFollowSeries(ctx, productURL, visitedURLs, visitedMutex)
				if len(results) > 0 {
					resultsMutex.Lock()
					*allResults = append(*allResults, results...)
					resultsMutex.Unlock()
					foundVolumes = true
				}

				// Adiciona um pequeno atraso entre produtos para evitar sobrecarga
				select {
				case <-time.After(1 * time.Second):
				case <-ctx.Done():
					return foundVolumes
				}
			}
		}
	}

	return foundVolumes
}

func (p *PaniniProvider) processProductAndFollowSeries(
	ctx context.Context,
	startURL string,
	visitedURLs map[string]bool,
	visitedMutex *sync.Mutex,
) []entity.CrawledResult {
	results := make([]entity.CrawledResult, 0)
	toVisit := []string{startURL}

	// Manter controle do volume mais recente encontrado
	seriesMaxVolume := 0
	seriesTitle := ""
	maxVolumeDetected := false

	// Para sites de busca, precisamos fazer uma extração especial
	// isSearchURL := strings.Contains(startURL, "/catalogsearch/result/")

	for len(toVisit) > 0 {
		select {
		case <-ctx.Done():
			return results
		default:
			currentURL := toVisit[0]
			toVisit = toVisit[1:] // Remove o primeiro item

			// Se for uma URL de busca, precisamos tratar diferente
			if strings.Contains(currentURL, "/catalogsearch/result/") {
				searchResults := p.handleSearchPage(ctx, currentURL, visitedURLs, visitedMutex)
				for _, searchURL := range searchResults {
					if !visitedURLs[searchURL] {
						visitedMutex.Lock()
						visitedURLs[searchURL] = true
						visitedMutex.Unlock()
						toVisit = append(toVisit, searchURL)
					}
				}
				continue
			}

			fmt.Printf("Processando produto: %s\n", currentURL)

			c := p.collector.Clone()
			result := entity.CrawledResult{
				Metadata: make(map[string]string),
			}
			resultFound := false
			var nextVolURL, prevVolURL, latestVolURL string
			latestVolNumber := 0

			// Extrai informações principais do produto
			c.OnHTML("div.product-info-main", func(e *colly.HTMLElement) {
				title := strings.TrimSpace(e.ChildText("h1.page-title span"))
				if title == "" {
					return
				}

				resultFound = true
				result = entity.CrawledResult{
					Title:     title,
					URL:       currentURL,
					Source:    p.SiteName(),
					Publisher: "Panini",
					Metadata:  make(map[string]string),
				}

				// Extrai o número do volume
				result.Volume = entity.ExtractVolumeNumber(title)

				// Se é o primeiro item e não temos um título de série ainda,
				// extraímos o título da série para usar como referência
				if seriesTitle == "" && result.Volume > 0 {
					seriesTitle = p.extractSeriesName(title)
				}

				// Marca este volume como encontrado - CORREÇÃO AQUI
				if result.Volume > 0 {
					p.foundVolumes[result.Volume] = true
					fmt.Printf(">> Marcando volume %d como encontrado para %s\n", result.Volume, seriesTitle)

					// Atualiza o volume máximo se necessário
					if result.Volume > seriesMaxVolume {
						seriesMaxVolume = result.Volume
					}
				}

				// Extrai o preço
				priceStr := strings.TrimSpace(e.ChildText("span.price"))
				if priceStr == "" {
					priceStr = e.ChildAttr("span.price-wrapper", "data-price-amount")
				}

				priceStr = strings.ReplaceAll(priceStr, "R$", "")
				priceStr = strings.ReplaceAll(priceStr, ".", "")
				priceStr = strings.ReplaceAll(priceStr, ",", ".")
				priceStr = strings.TrimSpace(priceStr)

				if priceStr != "" {
					if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
						result.Price = price
					}
				}

				// Verifica status do estoque
				stockStatus := strings.TrimSpace(e.ChildText("div.product-info-stock-sku div.stock span"))
				if stockStatus != "" {
					result.Metadata["stock_status"] = stockStatus
				}

				// Verifica status de pré-venda
				presaleText := strings.TrimSpace(e.ChildText("div.presale"))
				if presaleText != "" {
					result.Metadata["presale"] = "true"
					result.Metadata["presale_text"] = presaleText
				}
			})

			c.OnHTML("div.product.media", func(e *colly.HTMLElement) {
				coverImage := e.ChildAttr("img.gallery-placeholder__image", "src")
				if coverImage == "" {
					coverImage = e.ChildAttr("img.fotorama__img", "src")
				}
				if coverImage == "" {
					coverImage = e.ChildAttr("div.fotorama__stage__frame.fotorama__active img", "src")
				}
				if coverImage != "" {
					result.CoverImage = coverImage
				}
			})

			// Extrai detalhes do produto
			c.OnHTML("table#product-attribute-specs-table", func(e *colly.HTMLElement) {
				e.ForEach("tbody tr", func(_ int, row *colly.HTMLElement) {
					attrLabel := strings.TrimSpace(row.ChildText("th.col.label"))
					attrValue := strings.TrimSpace(row.ChildText("td.col.data"))

					// Fallback para data-th attribute
					if attrLabel == "" {
						attrLabel = strings.TrimSpace(row.ChildAttr("th.col", "data-th"))
					}

					switch strings.ToLower(attrLabel) {
					case "isbn", "código do produto", "código", "referência":
						result.ISBN = attrValue
					case "autor", "autores", "autor(es)":
						result.Author = strings.Split(attrValue, ",")
						// Limpa espaços nos nomes dos autores
						for i := range result.Author {
							result.Author[i] = strings.TrimSpace(result.Author[i])
						}
					case "data de lançamento", "lançamento":
						result.Metadata["release_date"] = attrValue
						// Tenta analisar formatos de data comuns no Brasil
						if parsedDate, err := time.Parse("02/01/2006", attrValue); err == nil {
							result.ReleasedAt = parsedDate.Format("2006-01-02")
						}
					default:
						if attrLabel != "" && attrValue != "" {
							result.Metadata[strings.ToLower(strings.ReplaceAll(attrLabel, " ", "_"))] = attrValue
						}
					}
				})
			})

			// Extrai descrição do produto
			c.OnHTML("div.product.attribute.overview div.value", func(e *colly.HTMLElement) {
				result.Description = strings.TrimSpace(e.Text)
			})

			// Esta é a parte crucial - extrair links para volumes relacionados
			c.OnHTML("div.volumes-container div.volume", func(e *colly.HTMLElement) {
				volumeLabel := strings.TrimSpace(e.ChildText("span"))
				volumeURL := e.ChildAttr("div.volume-actions div.volume-buy a", "href")
				volumeName := strings.TrimSpace(e.ChildText("div.volume-info span.name"))

				if volumeURL == "" || volumeName == "" {
					return
				}

				fmt.Printf("Encontrado volume relacionado: %s - %s\n", volumeLabel, volumeName)

				// Detecta o volume mais recente para usar como critério de parada
				volumeLower := strings.ToLower(volumeLabel)
				fmt.Println("Volume Lower:", volumeLower)
				if (strings.Contains(volumeLower, "mais recente") || strings.Contains(volumeLower, "recente")) && !maxVolumeDetected{
					latestVolURL = volumeURL
					result.Metadata["latest_volume"] = volumeName
					result.Metadata["latest_volume_url"] = volumeURL

					// Extrai número do volume mais recente
					latestVolNumber = entity.ExtractVolumeNumber(volumeName)
					fmt.Println("Volume Name:", volumeName)
					fmt.Println("Volume Number:", latestVolNumber)
					if latestVolNumber > 0 {
						result.Metadata["latest_volume_number"] = strconv.Itoa(latestVolNumber)

						// Marca que encontramos informação sobre o volume máximo
						maxVolumeDetected = true

						// Atualiza o volume máximo da série se o volume mais recente for maior
						if latestVolNumber > seriesMaxVolume {
							seriesMaxVolume = latestVolNumber
							fmt.Printf(">> Volume máximo da série atualizado para %d\n", seriesMaxVolume)
						}
					}
				} else if strings.Contains(volumeLower, "próximo") {
					nextVolURL = volumeURL
					result.Metadata["next_volume"] = volumeName
					result.Metadata["next_volume_url"] = volumeURL
				} else if strings.Contains(volumeLower, "anterior") {
					prevVolURL = volumeURL
					result.Metadata["previous_volume"] = volumeName
					result.Metadata["previous_volume_url"] = volumeURL
				}

				// Também extraímos o volume deste item relacionado para contabilizar
				relatedVolume := entity.ExtractVolumeNumber(volumeName)
				if relatedVolume > 0 {
					if relatedVolume > seriesMaxVolume {
						seriesMaxVolume = relatedVolume
					}
				}
			})

			// Quando a página for totalmente processada, adiciona o resultado e programa visitas adicionais
			c.OnScraped(func(r *colly.Response) {
				if resultFound {
					results = append(results, result)
					fmt.Printf("Processado: %s (Vol. %d) - R$ %.2f\n", result.Title, result.Volume, result.Price)

					// Dump dos volumes encontrados para depuração
					volList := make([]int, 0)
					for vol := range p.foundVolumes {
						volList = append(volList, vol)
					}
					sort.Ints(volList)
					fmt.Printf(">> Volumes encontrados: %v (total: %d)\n", volList, len(p.foundVolumes))

					// Verifica se encontramos todos os volumes da série
					if maxVolumeDetected && seriesMaxVolume > 0 {
						// Conta quantos volumes encontramos
						foundCount := len(p.foundVolumes)

						// Se encontramos todos ou quase todos (90%) dos volumes, podemos parar
						completeness := float64(foundCount) / float64(seriesMaxVolume)
						if completeness >= 0.9 {
							fmt.Printf("Encontrados %d/%d volumes (%.2f%%). Considerando a série completa.\n",
								foundCount, seriesMaxVolume, completeness*100)
							// Esvazia a lista toVisit para encerrar o loop
							toVisit = nil
							return
						}

						fmt.Printf("Progresso: %d/%d volumes encontrados (%.2f%%).\n",
							foundCount, seriesMaxVolume, completeness*100)
					}

					// Adiciona próximos URLs para visitar se ainda não foram visitados
					visitedMutex.Lock()

					// Se identificamos o volume mais recente, vamos priorizar a visita a volumes faltantes
					// em vez de seguir cegamente os links "próximo" e "anterior"
					if maxVolumeDetected && seriesMaxVolume > 0 {
						// Se o volume mais recente foi visitado, podemos priorizar volumes faltantes
						if len(toVisit) == 0 {
							// Encontra volumes faltantes
							for vol := 1; vol <= seriesMaxVolume; vol++ {
								if !p.foundVolumes[vol] {
									// Gera URL para tentar encontrar o volume faltante
									searchURL := fmt.Sprintf("https://panini.com.br/catalogsearch/result/?q=%s+vol+%d",
										strings.ReplaceAll(seriesTitle, " ", "+"), vol)

									if !visitedURLs[searchURL] {
										visitedURLs[searchURL] = true
										toVisit = append(toVisit, searchURL)
										fmt.Printf("Adicionando busca por volume faltante: %d\n", vol)
									}
								}
							}
						}
					} else {
						// Comportamento normal - seguir links próximo/anterior/mais recente
						if nextVolURL != "" && !visitedURLs[nextVolURL] {
							visitedURLs[nextVolURL] = true
							toVisit = append(toVisit, nextVolURL)
						}

						if prevVolURL != "" && !visitedURLs[prevVolURL] {
							visitedURLs[prevVolURL] = true
							toVisit = append(toVisit, prevVolURL)
						}

						if latestVolURL != "" && !visitedURLs[latestVolURL] {
							visitedURLs[latestVolURL] = true
							toVisit = append(toVisit, latestVolURL)
						}
					}

					visitedMutex.Unlock()
				}
			})

			err := c.Visit(currentURL)
			if err != nil {
				fmt.Printf("Erro ao visitar página do produto %s: %v\n", currentURL, err)
			}

			// Espera pela conclusão da visita
			c.Wait()

			// Adiciona um pequeno atraso entre visitas
			select {
			case <-time.After(2 * time.Second):
			case <-ctx.Done():
				return results
			}
		}
	}

	// Adiciona metadados sobre a completude da série nos resultados
	if seriesMaxVolume > 0 {
		seriesTitle = strings.TrimSpace(seriesTitle)

		for i := range results {
			// Se o título da série já foi estabelecido, use-o para verificar se este item pertence à série
			if seriesTitle != "" {
				itemSeriesTitle := p.extractSeriesName(results[i].Title)
				if itemSeriesTitle == seriesTitle {
					results[i].Metadata["series_max_volume"] = strconv.Itoa(seriesMaxVolume)
					results[i].Metadata["series_total_expected"] = strconv.Itoa(seriesMaxVolume)
					results[i].Metadata["series_found_count"] = strconv.Itoa(len(p.foundVolumes))

					completeness := float64(len(p.foundVolumes)) / float64(seriesMaxVolume)
					results[i].Metadata["series_completeness"] = fmt.Sprintf("%.2f", completeness)
				}
			}
		}
	}

	fmt.Printf("Processamento da série finalizado. Encontrados %d/%d volumes.\n",
		len(p.foundVolumes), seriesMaxVolume)
	return results
}

// searchGenericTerm realiza uma busca genérica quando não conseguimos encontrar volumes específicos
func (p *PaniniProvider) searchGenericTerm(
	ctx context.Context,
	term string,
	visitedURLs map[string]bool,
	visitedMutex *sync.Mutex,
) []entity.CrawledResult {
	searchURL := fmt.Sprintf("https://panini.com.br/catalogsearch/result/?q=%s", strings.ReplaceAll(term, " ", "+"))
	fmt.Printf("Realizando busca genérica: %s\n", term)

	c := p.collector.Clone()
	results := make([]entity.CrawledResult, 0)
	productURLs := make([]string, 0)
	pageURLs := make(map[string]bool) // Para evitar loops infinitos na paginação

	// Coleta URLs de produtos na página de resultados
	c.OnHTML("div.products div.product-item-info", func(e *colly.HTMLElement) {
		productURL := e.ChildAttr("a.product-item-photo", "href")
		if productURL == "" {
			productURL = e.ChildAttr("a.product-item-link", "href")
		}

		if productURL != "" {
			visitedMutex.Lock()
			if !visitedURLs[productURL] {
				visitedURLs[productURL] = true
				productURLs = append(productURLs, productURL)
			}
			visitedMutex.Unlock()
		}
	})

	// Verifica paginação - CORRIGIDO para evitar loops
	c.OnHTML("ul.pages-items li.pages-item-next a", func(e *colly.HTMLElement) {
		nextPageURL := e.Request.AbsoluteURL(e.Attr("href"))
		if nextPageURL != "" && p.pagesCrawled < p.opts.MaxPages {
			// Verifica se já visitamos esta página
			if _, exists := pageURLs[nextPageURL]; !exists {
				pageURLs[nextPageURL] = true
				p.pagesCrawled++
				fmt.Printf("Seguindo para próxima página: %s (página %d)\n", nextPageURL, p.pagesCrawled+1)

				// Usar Visit em vez de VisitURL para manter o contexto
				c.Visit(nextPageURL)
			}
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visitando página de busca: %s\n", r.URL)
	})

	err := c.Visit(searchURL)
	if err != nil {
		fmt.Printf("Erro ao visitar página de busca %s: %v\n", searchURL, err)
		return results
	}

	// Esperar pela conclusão das visitas - importante para paginação
	c.Wait()
	fmt.Printf("Concluída busca genérica: %s - Encontrados %d produtos\n", term, len(productURLs))

	// Processa cada produto encontrado
	for _, productURL := range productURLs {
		select {
		case <-ctx.Done():
			return results
		default:
			seriesResults := p.processProductAndFollowSeries(ctx, productURL, visitedURLs, visitedMutex)
			results = append(results, seriesResults...)

			// Adiciona um pequeno atraso entre produtos
			time.Sleep(1 * time.Second)
		}
	}

	return results
}

// processResults processa os resultados finais, agrupando-os por série e adicionando metadados úteis
func (p *PaniniProvider) processResults(results []entity.CrawledResult) []entity.CrawledResult {
	if len(results) == 0 {
		return results
	}

	// Agrupa resultados por série (baseado no título sem o número do volume)
	seriesGroups := make(map[string][]entity.CrawledResult)

	for _, result := range results {
		seriesName := p.extractSeriesName(result.Title)
		seriesGroups[seriesName] = append(seriesGroups[seriesName], result)
	}

	processedResults := make([]entity.CrawledResult, 0)

	// Processa cada grupo de série
	for seriesName, seriesResults := range seriesGroups {
		// Ordena os volumes em ordem crescente
		sort.Slice(seriesResults, func(i, j int) bool {
			return seriesResults[i].Volume < seriesResults[j].Volume
		})

		// Encontra o volume máximo da série
		maxVolume := 0
		for _, result := range seriesResults {
			if result.Volume > maxVolume {
				maxVolume = result.Volume
			}
		}

		// Adiciona informações da série aos metadados
		for i := range seriesResults {
			if seriesResults[i].Metadata == nil {
				seriesResults[i].Metadata = make(map[string]string)
			}

			seriesResults[i].Metadata["series_name"] = seriesName
			seriesResults[i].Metadata["series_total_found"] = strconv.Itoa(len(seriesResults))
			seriesResults[i].Metadata["series_max_volume"] = strconv.Itoa(maxVolume)

			// Calcula a completude da série (quantos volumes encontrados vs. esperados)
			volumeMap := make(map[int]bool)
			for _, res := range seriesResults {
				if res.Volume > 0 {
					volumeMap[res.Volume] = true
				}
			}

			if maxVolume > 0 {
				completeness := float64(len(volumeMap)) / float64(maxVolume)
				seriesResults[i].Metadata["series_completeness"] = fmt.Sprintf("%.2f", completeness)
			}

			// Adiciona flag para parte de série
			seriesResults[i].Metadata["part_of_series"] = "true"
		}

		// Adiciona todos os resultados processados
		processedResults = append(processedResults, seriesResults...)
	}

	// Remove duplicatas por URL
	seen := make(map[string]bool)
	uniqueResults := make([]entity.CrawledResult, 0)

	for _, result := range processedResults {
		if !seen[result.URL] {
			seen[result.URL] = true
			uniqueResults = append(uniqueResults, result)
		}
	}

	fmt.Printf("Panini: processamento finalizado com %d resultados, %d resultados únicos\n",
		len(processedResults), len(uniqueResults))

	return uniqueResults
}

// extractSeriesName remove informações de volume do título para obter o nome da série
func (p *PaniniProvider) extractSeriesName(title string) string {
	title = strings.TrimSpace(title)

	// Padrões comuns para volumes em títulos
	patterns := []string{
		`\s+[Vv]ol(\.|ume)?\s*\d+$`,
		`\s+[Vv]ol(\.|ume)?\s*\d+\s*$`,
		`\s+[Tt]omo\s*\d+$`,
		`\s+#\d+$`,
		`\s+-\s+[Vv]ol(ume)?\s*\d+\s*$`,
		`\s+\d+$`, // Apenas número no final
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		title = re.ReplaceAllString(title, "")
	}

	// Limpa pontuação final
	title = strings.TrimRight(title, " -:,.")

	return strings.TrimSpace(title)
}

func (p *PaniniProvider) handleSearchPage(
	ctx context.Context,
	searchURL string,
	visitedURLs map[string]bool,
	visitedMutex *sync.Mutex,
) []string {
	productURLs := make([]string, 0)
	c := p.collector.Clone()

	c.OnHTML("div.products div.product-item-info", func(e *colly.HTMLElement) {
		productURL := e.ChildAttr("a.product-item-photo", "href")
		if productURL == "" {
			productURL = e.ChildAttr("a.product-item-link", "href")
		}

		if productURL != "" && !strings.Contains(productURL, "/catalogsearch/result/") {
			visitedMutex.Lock()
			if !visitedURLs[productURL] {
				visitedURLs[productURL] = true
				productURLs = append(productURLs, productURL)
			}
			visitedMutex.Unlock()
		}
	})

	err := c.Visit(searchURL)
	if err != nil {
		fmt.Printf("Erro ao visitar página de busca %s: %v\n", searchURL, err)
	}

	c.Wait()
	return productURLs
}
