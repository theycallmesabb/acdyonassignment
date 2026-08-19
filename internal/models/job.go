package models

import "time"

// Job represents a normalized job listing.
type Job struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
	FetchedAt   time.Time `json:"fetched_at"`
}

// IngestionMetrics tracks statistics about ingestion runs.
type IngestionMetrics struct {
	TotalRuns            int64            `json:"total_runs"`
	SuccessfulRuns       int64            `json:"successful_runs"`
	FailedRuns           int64            `json:"failed_runs"`
	TotalJobsFetched     int64            `json:"total_jobs_fetched"`
	ActiveSource         string           `json:"active_source"`
	CircuitBreakerStatus string           `json:"circuit_breaker_status"` // "CLOSED" (healthy), "OPEN" (tripped), "HALF-OPEN"
	LastRunTime          time.Time        `json:"last_run_time"`
	LastError            string           `json:"last_error,omitempty"`
	SourceStats          map[string]int64 `json:"source_stats"`
}

// SourceConfig defines a feed target.
type SourceConfig struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Type        string `json:"type"` // "rss", "json_api", "sandbox"
	Priority    int    `json:"priority"`
	IsFallback  bool   `json:"is_fallback"`
}
