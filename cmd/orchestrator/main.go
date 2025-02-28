package main

import (
	"log"
	"os"

	"github.com/w0ikid/megacalc/internal/api"
	"github.com/w0ikid/megacalc/internal/service"
)

func main() {
	log.Println("Starting orchestrator...")
	
	// Get operation times from environment variables
	opTimes := api.GetOperationTimes()
	
	// Create service
	svc := service.NewService(opTimes)
	
	// Create handler
	handler := api.NewHandler(svc)
	
	// Get address from environment variable or use default
	addr := os.Getenv("ORCHESTRATOR_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	
	log.Printf("Orchestrator listening on %s", addr)
	
	// Start server
	err := handler.Start(addr)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}