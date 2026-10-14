package main

import (
	"fmt"
	"os"

	"github.com/iCoderabhishek/Sutra-AI/apps/cli/internal/api"
	"github.com/iCoderabhishek/Sutra-AI/apps/cli/internal/config"
	"github.com/iCoderabhishek/Sutra-AI/apps/cli/internal/tui"
)

// Set by release builds via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Println("sutra", version)
			return
		case "--help", "-h", "help":
			fmt.Println("sutra: scheduled AI agents in your terminal")
			fmt.Println()
			fmt.Println("Usage: sutra            open the terminal UI")
			fmt.Println("       sutra --version  print the version")
			fmt.Println()
			fmt.Println("Server:", config.GetEnvConfig().BackendURL, "(override with SUTRA_BACKEND_URL)")
			return
		}
	}

	env := config.GetEnvConfig()
	client := api.NewClient(env.BackendURL, "")

	// A missing session is fine: the TUI shows the sign-in screen on 401.
	_ = client.InitAuth()

	if err := tui.Run(client); err != nil {
		fmt.Fprintln(os.Stderr, "sutra:", err)
		os.Exit(1)
	}
}
