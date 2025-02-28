package main

import (
	"log"
	"os"
	"strconv"

	"github.com/w0ikid/megacalc/internal/agent"
)

func main() {
	log.Println("Starting agent...")
	
	// Get orchestrator URL from environment variable
	orchestratorURL := os.Getenv("ORCHESTRATOR_URL")
	if orchestratorURL == "" {
		orchestratorURL = "http://orchestrator:8080"
	}
	
	// Get computing power from environment variable
	computingPowerStr := os.Getenv("COMPUTING_POWER")
	computingPower := 4 // default value
	if computingPowerStr != "" {
		var err error
		computingPower, err = strconv.Atoi(computingPowerStr)
		if err != nil {
			log.Printf("Invalid COMPUTING_POWER value: %s, using default: %d", computingPowerStr, computingPower)
		}
	}
	
	// Create agent
	a := agent.NewAgent(orchestratorURL, computingPower)
	
	// Start agent
	a.Start()
}