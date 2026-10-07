package main

import (
	"context"
	"fmt"
	"log"

	"github.com/iCoderabhishek/Sutra-AI/internal/api"
	"github.com/iCoderabhishek/Sutra-AI/internal/config"
)

func main() {
	// 1. Initialize environment and base client
	env := config.GetEnvConfig()
	client := api.NewClient(env.BackendURL, "")

	fmt.Println("🤖 Sutra CLI Scratchpad 🤖")
	fmt.Println("---------------------------")
	
	// 2. Try to load existing auth session, or fallback to login flow
	if err := client.InitAuth(); err != nil {
		fmt.Printf("No existing session found: %v\n", err)
	}

	fmt.Println("Checking authentication by fetching dashboard...")
	dashboard, err := client.GetDashboard()
	
	if err != nil {
		fmt.Printf("Auth check failed (%v). Starting interactive login flow...\n", err)
		if err := client.Login(); err != nil {
			log.Fatalf("❌ Login failed: %v", err)
		}
		
		fmt.Println("Fetching dashboard again after login...")
		dashboard, err = client.GetDashboard()
		if err != nil {
			log.Fatalf("❌ Failed to fetch dashboard: %v", err)
		}
	}

	fmt.Println("✅ Successfully authenticated!")
	fmt.Printf("📊 Dashboard Stats -> Total Agents: %d | Credit Balance: %.2f\n", dashboard.Agents.Total, dashboard.Credits.Balance)
	fmt.Println("---------------------------")

	// 3. Fetch Agents
	fmt.Println("Fetching agents...")
	agents, err := client.ListAgents()
	if err != nil {
		log.Fatalf("❌ Failed to list agents: %v", err)
	}

	if len(agents) == 0 {
		fmt.Println("No agents found in your account. Go create one in the web dashboard first!")
		return
	}

	agent := agents[0]
	fmt.Printf("✅ Found Agent: %s (ID: %s)\n", agent.Name, agent.ID)
	fmt.Println("---------------------------")

	// 4. Trigger Run
	fmt.Printf("Triggering a run for Agent '%s'...\n", agent.Name)
	run, err := client.TriggerRun(agent.ID)
	if err != nil {
		log.Fatalf("❌ Failed to trigger run: %v", err)
	}
	fmt.Printf("✅ Run started! Run ID: %s (Status: %s)\n", run.ID, run.Status)
	fmt.Println("---------------------------")

	// 5. Stream Run Logs
	fmt.Println("📡 Connecting to Server-Sent Events stream...")
	events, errors := client.StreamRunLogs(context.Background(), run.ID)

	fmt.Println("--- LIVE TRACE STREAM ---")
	for {
		select {
		case event, ok := <-events:
			if !ok {
				fmt.Println("\n🏁 Stream closed cleanly.")
				return
			}
			
			// Format the event output
			statusIcon := "🔄"
			if event.Status == api.TraceEventDone {
				statusIcon = "✅"
			} else if event.Status == api.TraceEventError {
				statusIcon = "❌"
			}
			
			fmt.Printf("%s [%s] %s\n", statusIcon, event.Status, event.Step)
			
			if event.ResultPreview != nil && *event.ResultPreview != "" {
				fmt.Printf("   -> %s\n", *event.ResultPreview)
			}
			
		case err, ok := <-errors:
			if ok && err != nil {
				log.Fatalf("\n❌ Stream error: %v", err)
			}
			return
		}
	}
}
