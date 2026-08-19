package ingestor

import (
	"testing"
	"time"
)

func TestRandomUserAgent(t *testing.T) {
	ua := GetRandomUserAgent()
	if ua == "" {
		t.Errorf("expected non-empty User-Agent string")
	}
}

func TestBuildHumanizedHeaders(t *testing.T) {
	ua := "TestUA/1.0"
	headers := BuildHumanizedHeaders(ua)
	if headers["User-Agent"] != ua {
		t.Errorf("expected User-Agent header %s, got %s", ua, headers["User-Agent"])
	}
	if headers["Accept-Language"] == "" {
		t.Errorf("expected Accept-Language header to be present")
	}
}

func TestCircuitBreaker(t *testing.T) {
	client := NewResilientClient(2, 500*time.Millisecond)

	if client.GetState() != StateClosed {
		t.Errorf("expected initial state to be CLOSED, got %s", client.GetState())
	}

	client.RecordFailure()
	if client.GetState() != StateClosed {
		t.Errorf("expected state to remain CLOSED after 1 failure, got %s", client.GetState())
	}

	client.RecordFailure()
	if client.GetState() != StateOpen {
		t.Errorf("expected state to be OPEN after reaching failure threshold, got %s", client.GetState())
	}

	time.Sleep(600 * time.Millisecond)
	if client.GetState() != StateHalfOpen {
		t.Errorf("expected state to be HALF-OPEN after cooldown, got %s", client.GetState())
	}

	client.RecordSuccess()
	if client.GetState() != StateClosed {
		t.Errorf("expected state to reset to CLOSED after success, got %s", client.GetState())
	}
}

func TestEngineFallback(t *testing.T) {
	engine := NewEngine()
	jobs, err := engine.Ingest()
	if err != nil {
		t.Fatalf("expected ingestion to succeed (via public feed or fallback sandbox), got error: %v", err)
	}

	if len(jobs) == 0 {
		t.Errorf("expected non-zero jobs returned")
	}

	metrics := engine.GetMetrics()
	if metrics.TotalRuns != 1 {
		t.Errorf("expected TotalRuns to be 1, got %d", metrics.TotalRuns)
	}
	if metrics.ActiveSource == "" {
		t.Errorf("expected ActiveSource to be set")
	}
}
