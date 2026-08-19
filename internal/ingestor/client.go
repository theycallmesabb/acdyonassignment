package ingestor

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// CircuitBreakerState represents the status of the circuit breaker.
type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "CLOSED"    // Healthy, operating normally
	StateOpen     CircuitBreakerState = "OPEN"      // Tripped due to blocks/failures, blocking requests
	StateHalfOpen CircuitBreakerState = "HALF-OPEN" // Testing recovery
)

// ResilientClient manages HTTP requests with jittered delays, rotation, retries, and circuit breaker.
type ResilientClient struct {
	httpClient       *http.Client
	failureThreshold int
	cooldownDuration time.Duration

	mu              sync.RWMutex
	state           CircuitBreakerState
	consecutiveFails int
	lastStateChange time.Time
}

// NewResilientClient creates an instance of ResilientClient.
func NewResilientClient(failureThreshold int, cooldownDuration time.Duration) *ResilientClient {
	return &ResilientClient{
		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},
		failureThreshold: failureThreshold,
		cooldownDuration: cooldownDuration,
		state:            StateClosed,
		lastStateChange:  time.Now(),
	}
}

// GetState returns current circuit breaker state.
func (c *ResilientClient) GetState() CircuitBreakerState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.state == StateOpen && time.Since(c.lastStateChange) > c.cooldownDuration {
		return StateHalfOpen
	}
	return c.state
}

// RecordSuccess registers a successful HTTP request.
func (c *ResilientClient) RecordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.consecutiveFails = 0
	if c.state != StateClosed {
		c.state = StateClosed
		c.lastStateChange = time.Now()
	}
}

// RecordFailure registers a failed request or detected block.
func (c *ResilientClient) RecordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.consecutiveFails++
	if c.consecutiveFails >= c.failureThreshold {
		c.state = StateOpen
		c.lastStateChange = time.Now()
	}
}

// HumanizedDelay adds a gaussian-like jittered delay to simulate human browsing intervals.
func (c *ResilientClient) HumanizedDelay(minMs, maxMs int) {
	if maxMs <= minMs {
		maxMs = minMs + 100
	}
	delayMs := minMs + rand.Intn(maxMs-minMs)
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
}

// FetchBody makes an HTTP GET request with rotated headers, jitter, and error detection.
func (c *ResilientClient) FetchBody(ctx string, targetURL string) ([]byte, int, error) {
	state := c.GetState()
	if state == StateOpen {
		return nil, http.StatusServiceUnavailable, fmt.Errorf("circuit breaker is OPEN: target host rate-limiting or blocking requests")
	}

	// Apply humanized timing delay before sending request
	c.HumanizedDelay(300, 900)

	req, err := http.NewRequestWithContext(context.Background(), "GET", targetURL, nil)
	if err != nil {
		return nil, 0, err
	}

	// Attach humanized headers with User-Agent rotation
	ua := GetRandomUserAgent()
	headers := BuildHumanizedHeaders(ua)
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.RecordFailure()
		return nil, 0, fmt.Errorf("network error fetching %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	// Detection: Block/Rate-limit status codes (403 Forbidden, 429 Too Many Requests, 503 Service Unavailable)
	if resp.StatusCode == 403 || resp.StatusCode == 429 || resp.StatusCode == 503 {
		c.RecordFailure()
		return nil, resp.StatusCode, fmt.Errorf("bot detection / rate-limiting triggered (HTTP %d)", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		c.RecordFailure()
		return nil, resp.StatusCode, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.RecordFailure()
		return nil, resp.StatusCode, fmt.Errorf("error reading response body: %w", err)
	}

	c.RecordSuccess()
	return body, resp.StatusCode, nil
}
