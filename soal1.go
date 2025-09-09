package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/time/rate"
)

type ScrapResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Title      string `json:"title"`
	Error      string `json:"error,omitempty"`
	Duration   string `json:"duration"`
}

type Scraper interface {
	ScrapeURLs(ctx context.Context, urls []string) ([]ScrapResult, error)
}

type CollyScraper struct {
	limiter *rate.Limiter
}

func NewCollyScraper(rps int) *CollyScraper {
	return &CollyScraper{
		limiter: rate.NewLimiter(rate.Limit(rps), rps),
	}
}

func (s *CollyScraper) ScrapeURLs(ctx context.Context, urls []string) ([]ScrapResult, error) {
	results := make([]ScrapResult, len(urls))

	var wg sync.WaitGroup
	// var mu sync.Mutex

	for i, url := range urls {
		wg.Add(1)

		go func(i int, url string) {
			defer wg.Done()
			var result ScrapResult
			result.URL = url

			timeStart := time.Now()

			for attempt := 1; attempt <= 3; attempt++ {
				if err := s.limiter.Wait(ctx); err != nil {
					result.Error = err.Error()
				}
				reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
				defer cancel()

				c := colly.NewCollector(
					colly.Async(false),
				)

				var reqErr error

				c.OnError(func(r *colly.Response, err error) {
					reqErr = err
				})

				c.OnRequest(func(r *colly.Request) {
					fmt.Println("Scrapping: " + r.URL.String())
				})

				c.OnResponse(func(r *colly.Response) {
					result.StatusCode = r.StatusCode
				})

				c.OnHTML("title", func(h *colly.HTMLElement) {
					result.Title = h.Text
				})

				err := c.Request("GET", url, nil, nil, nil)

				if err != nil {
					reqErr = err
				}

				select {
				case <-reqCtx.Done():
					reqErr = fmt.Errorf("timeout %w", reqCtx.Err())
				default:
				}

				if reqErr == nil {
					break
				} else {
					result.Error = reqErr.Error()
					if attempt < 3 {
						time.Sleep(time.Second)
					}
				}

			}

			result.Duration = time.Since(timeStart).String()
			results[i] = result

		}(i, url)
	}

	wg.Wait()

	return results, nil
}

func Soal1() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	websites := []string{"https://health.detik.com/fotohealth/d-8103170/aksi-dokter-swiss-mogok-makan-desak-pemerintah-bersikap-tegas-soal-gaza", "https://health.detik.com/berita-detikhealth/d-8103532/viral-dokter-anestesi-disebut-dipukul-keluarga-pasien-kemenkes-bilang-gini", "https://health.detik.com/berita-detikhealth/d-8100708/ternyata-ini-alasan-kim-jong-un-hapus-jejak-dan-bawa-toilet-sendiri-saat-di-beijing"}

	scraper := NewCollyScraper(5)

	results, err := scraper.ScrapeURLs(ctx, websites)

	if err != nil {
		fmt.Println(err)
	}

	convertJson, _ := json.MarshalIndent(results, "", "  ")

	fmt.Println(string(convertJson))
}
