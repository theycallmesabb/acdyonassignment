package main

import (
	"log"
	"os"

	"job-ingestion-service/internal/api"
	"job-ingestion-service/internal/ingestor"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Initialize Ingestion Engine
	engine := ingestor.NewEngine()

	// Perform initial warm-up ingestion run
	log.Println("[INFO] Performing initial warm-up ingestion run...")
	if _, err := engine.Ingest(); err != nil {
		log.Printf("[WARN] Initial ingestion encountered issue (will recover on endpoints): %v\n", err)
	} else {
		log.Println("[INFO] Initial ingestion completed successfully.")
	}

	// Register API Routes
	handler := api.NewHandler(engine)
	handler.RegisterRoutes(r)

	log.Printf("[INFO] Starting Job Ingestion Service on port %s...\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}
}
