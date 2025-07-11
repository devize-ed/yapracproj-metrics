package main

import (
	"log"

	"github.com/devize-ed/yapracproj-metrics.git/internal/agent"
	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/go-resty/resty/v2"
)

func main() {
	cfg, err := config.GetAgentConfig() // call the function to get agent configuration
	if err != nil {
		log.Fatal("Failed to get agent config:", err)
	}
	client := resty.New() // init client

	a := agent.NewAgent(client, cfg) // create a new agent instance
	a.Run()                          // start the agent to collect and report metrics
	log.Println("Agent started with config:", cfg)
}
