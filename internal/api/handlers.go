package api

import (
	"net/http"
	"strings"

	"job-ingestion-service/internal/ingestor"

	"github.com/gin-gonic/gin"
)

// Handler contains HTTP handlers for the Gin service.
type Handler struct {
	engine *ingestor.Engine
}

// NewHandler initializes a Handler.
func NewHandler(engine *ingestor.Engine) *Handler {
	return &Handler{engine: engine}
}

// RegisterRoutes sets up API endpoints.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", h.HealthCheck)
	r.GET("/api/jobs", h.GetJobs)
	r.POST("/api/ingest", h.TriggerIngest)
	r.GET("/api/metrics", h.GetMetrics)
}

// HealthCheck returns service status and circuit breaker health.
func (h *Handler) HealthCheck(c *gin.Context) {
	metrics := h.engine.GetMetrics()
	c.JSON(http.StatusOK, gin.H{
		"status":                 "healthy",
		"service":                "job-ingestion-engine",
		"circuit_breaker_status": metrics.CircuitBreakerStatus,
		"active_source":          metrics.ActiveSource,
	})
}

// GetJobs returns retrieved job listings with optional title filter.
func (h *Handler) GetJobs(c *gin.Context) {
	jobs := h.engine.GetJobs()
	filter := strings.ToLower(c.Query("q"))

	if filter != "" {
		var filtered []interface{}
		for _, j := range jobs {
			if strings.Contains(strings.ToLower(j.Title), filter) || strings.Contains(strings.ToLower(j.Company), filter) {
				filtered = append(filtered, j)
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"count": len(filtered),
			"jobs":  filtered,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(jobs),
		"jobs":  jobs,
	})
}

// TriggerIngest forces an immediate ingestion run across sources.
func (h *Handler) TriggerIngest(c *gin.Context) {
	jobs, err := h.engine.Ingest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"metrics": h.engine.GetMetrics(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Ingestion successful",
		"count":         len(jobs),
		"active_source": h.engine.GetMetrics().ActiveSource,
		"jobs":          jobs,
	})
}

// GetMetrics returns detailed operational statistics.
func (h *Handler) GetMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, h.engine.GetMetrics())
}
