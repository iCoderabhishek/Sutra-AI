package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/iCoderabhishek/Sutra-AI/internal/api"
)

// app is the root model. It owns the active screen and routes messages to it.
// Only the home screen exists so far; other screens plug in via NavigateMsg.
type app struct {
	client *api.Client
	home   homeModel
}

func newApp(client *api.Client) app {
	return app{client: client, home: newHome(client)}
}

func (a app) Init() tea.Cmd { return a.home.Init() }

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if nav, ok := msg.(NavigateMsg); ok {
		// Placeholder until the other screens are built.
		names := map[Screen]string{
			ScreenRunAgent:    "Run an agent",
			ScreenCreateAgent: "Create agent",
			ScreenAgents:      "My agents",
			ScreenRuns:        "Run history",
		}
		msg = toastMsg{text: names[nav.To] + " is coming soon"}
	}

	if k, ok := msg.(tea.KeyMsg); ok && (k.String() == "ctrl+c" || k.String() == "q") {
		return a, tea.Quit
	}

	var cmd tea.Cmd
	a.home, cmd = a.home.Update(msg)
	return a, cmd
}

func (a app) View() string { return a.home.View() }

// Run starts the full-screen TUI and blocks until the user quits.
func Run(client *api.Client) error {
	_, err := tea.NewProgram(newApp(client), tea.WithAltScreen()).Run()
	return err
}
