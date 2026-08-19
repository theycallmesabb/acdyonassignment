package ingestor

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"job-ingestion-service/internal/models"

	"github.com/mmcdole/gofeed"
)

// Engine orchestrates low-risk sources, parsing, and fallback execution.
type Engine struct {
	client  *ResilientClient
	sources []models.SourceConfig

	mu       sync.RWMutex
	jobs     []models.Job
	metrics  models.IngestionMetrics
	parser   *gofeed.Parser
}

// NewEngine initializes the ingestion pipeline engine.
func NewEngine() *Engine {
	client := NewResilientClient(3, 30*time.Second)

	sources := []models.SourceConfig{
		{
			Name:        "RemoteOK RSS",
			URL:         "https://remoteok.com/remote-jobs.rss",
			Type:        "rss",
			Priority:    1,
			IsFallback:  false,
		},
		{
			Name:        "Jobspire Public RSS Backup",
			URL:         "https://weworkremotely.com/remote-jobs.rss",
			Type:        "rss",
			Priority:    2,
			IsFallback:  false,
		},
		{
			Name:        "Sandbox Mock Feed",
			URL:         "mock://sandbox-feed",
			Type:        "sandbox",
			Priority:    3,
			IsFallback:  true,
		},
	}

	return &Engine{
		client:  client,
		sources: sources,
		jobs:    make([]models.Job, 0),
		parser:  gofeed.NewParser(),
		metrics: models.IngestionMetrics{
			CircuitBreakerStatus: string(StateClosed),
			SourceStats:          make(map[string]int64),
		},
	}
}

// Ingest runs the ingestion strategy with humanized fallback across configured sources.
func (e *Engine) Ingest() ([]models.Job, error) {
	e.mu.Lock()
	e.metrics.TotalRuns++
	e.metrics.LastRunTime = time.Now()
	e.mu.Unlock()

	var fetchedJobs []models.Job
	var lastErr error
	var activeSourceUsed string

	for _, src := range e.sources {
		var jobs []models.Job
		var err error

		if src.Type == "sandbox" {
			jobs = e.generateSandboxJobs()
			activeSourceUsed = src.Name
			err = nil
		} else {
			jobs, err = e.fetchAndParseRSS(src)
		}

		if err == nil && len(jobs) > 0 {
			fetchedJobs = jobs
			activeSourceUsed = src.Name
			break
		}

		lastErr = err
		// Delay slightly before trying fallback source
		e.client.HumanizedDelay(400, 800)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.metrics.CircuitBreakerStatus = string(e.client.GetState())

	if len(fetchedJobs) > 0 {
		e.metrics.SuccessfulRuns++
		e.metrics.TotalJobsFetched += int64(len(fetchedJobs))
		e.metrics.ActiveSource = activeSourceUsed
		e.metrics.SourceStats[activeSourceUsed] += int64(len(fetchedJobs))
		e.jobs = fetchedJobs
		return fetchedJobs, nil
	}

	e.metrics.FailedRuns++
	if lastErr != nil {
		e.metrics.LastError = lastErr.Error()
	} else {
		e.metrics.LastError = "No job listings parsed from available sources"
	}
	return nil, fmt.Errorf("ingestion failed across all sources: %v", e.metrics.LastError)
}

// fetchAndParseRSS retrieves an RSS XML payload via ResilientClient and parses jobs.
func (e *Engine) fetchAndParseRSS(src models.SourceConfig) ([]models.Job, error) {
	body, _, err := e.client.FetchBody("rss", src.URL)
	if err != nil {
		return nil, fmt.Errorf("source %s fetch error: %w", src.Name, err)
	}

	feed, err := e.parser.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("source %s parse error: %w", src.Name, err)
	}

	var parsedJobs []models.Job
	for idx, item := range feed.Items {
		pubTime := time.Now()
		if item.PublishedParsed != nil {
			pubTime = *item.PublishedParsed
		}

		company := "Various Companies"
		if item.Author != nil && item.Author.Name != "" {
			company = item.Author.Name
		}

		job := models.Job{
			ID:          fmt.Sprintf("%s-%d", src.Name, idx+1),
			Title:       item.Title,
			Company:     company,
			Location:    "Remote",
			URL:         item.Link,
			Description: item.Description,
			Source:      src.Name,
			PublishedAt: pubTime,
			FetchedAt:   time.Now(),
		}
		parsedJobs = append(parsedJobs, job)
	}

	return parsedJobs, nil
}

// generateSandboxJobs returns mock job data when external sources are blocked or unavailable.
func (e *Engine) generateSandboxJobs() []models.Job {
	now := time.Now()
	return []models.Job{
		{
			ID:          "sandbox-1",
			Title:       "Senior Go Backend Engineer (Distributed Systems)",
			Company:     "CloudScale Solutions",
			Location:    "Remote / US-Only",
			URL:         "https://example.com/jobs/sandbox-1",
			Description: "Designing resilient ingestion microservices with Go, Gin, and Docker.",
			Source:      "Sandbox Mock Feed",
			PublishedAt: now.Add(-2 * time.Hour),
			FetchedAt:   now,
		},
		{
			ID:          "sandbox-2",
			Title:       "Lead Data Pipeline Architect",
			Company:     "DataStream HQ",
			Location:    "Remote / Worldwide",
			URL:         "https://example.com/jobs/sandbox-2",
			Description: "Building anti-fragile streaming data crawlers with circuit breaker patterns.",
			Source:      "Sandbox Mock Feed",
			PublishedAt: now.Add(-5 * time.Hour),
			FetchedAt:   now,
		},
		{
			ID:          "sandbox-3",
			Title:       "Site Reliability & Security Engineer",
			Company:     "Resilient Net",
			Location:    "Remote / EU",
			URL:         "https://example.com/jobs/sandbox-3",
			Description: "Specializing in proxy mesh orchestration, TLS fingerprints, and rate mitigation.",
			Source:      "Sandbox Mock Feed",
			PublishedAt: now.Add(-12 * time.Hour),
			FetchedAt:   now,
		},
	}
}

// GetJobs returns currently stored jobs.
func (e *Engine) GetJobs() []models.Job {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.jobs
}

// GetMetrics returns ingestion execution metrics.
func (e *Engine) GetMetrics() models.IngestionMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()
	e.metrics.CircuitBreakerStatus = string(e.client.GetState())
	return e.metrics
}
