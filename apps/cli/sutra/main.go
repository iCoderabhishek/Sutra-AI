package main

import (
	"fmt"
	"os"

	"github.com/iCoderabhishek/Sutra-AI/internal/api"
	"github.com/iCoderabhishek/Sutra-AI/internal/config"
	"github.com/iCoderabhishek/Sutra-AI/internal/tui"
)

func main() {
	env := config.GetEnvConfig()
	client := api.NewClient(env.BackendURL, "")

	// A missing session is fine: the TUI shows the sign-in screen on 401.
	_ = client.InitAuth()

	if err := tui.Run(client); err != nil {
		fmt.Fprintln(os.Stderr, "sutra:", err)
		os.Exit(1)
	}
}
