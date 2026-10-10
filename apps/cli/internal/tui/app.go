package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/iCoderabhishek/Sutra-AI/internal/api"
)

// app is the root model. Home is always at the bottom; other screens are
// stacked on top of it. Keys go to the top screen only. Other messages (API
// results, spinner ticks) go to home and the top screen, and each ignores
// what isn't meant for it.
type app struct {
	client        *api.Client
	home          homeModel
	stack         []screen
	width, height int
}

func newApp(client *api.Client) app {
	return app{client: client, home: newHome(client)}
}

func (a app) Init() tea.Cmd { return a.home.Init() }

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
		if len(a.stack) == 0 {
			if msg.String() == "q" {
				return a, tea.Quit
			}
			return a.updateHome(msg)
		}
		return a.updateTop(msg)

	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		cmds := []tea.Cmd{}
		var cmd tea.Cmd
		a.home, cmd = a.home.Update(msg)
		cmds = append(cmds, cmd)
		for i := range a.stack {
			a.stack[i], cmd = a.stack[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)

	case NavigateMsg:
		return a, push(a.screenFor(msg.To))

	case pushMsg:
		a.stack = append(a.stack, msg.screen)
		return a, a.start(len(a.stack) - 1)

	case replaceMsg:
		if len(a.stack) == 0 {
			return a, push(msg.screen)
		}
		a.stack[len(a.stack)-1] = msg.screen
		return a, a.start(len(a.stack) - 1)

	case popMsg:
		if len(a.stack) > 0 {
			a.stack = a.stack[:len(a.stack)-1]
		}
		next, cmd := a.updateActive(resumedMsg{})
		if msg.toast == "" {
			return next, cmd
		}
		next2, cmd2 := next.(app).updateActive(toastMsg{msg.toast})
		return next2, tea.Batch(cmd, cmd2)

	case toastMsg, resumedMsg:
		return a.updateActive(msg)
	}

	// Everything else: home and the top screen.
	next, cmd := a.updateHome(msg)
	a = next.(app)
	if len(a.stack) == 0 {
		return a, cmd
	}
	next, cmd2 := a.updateTop(msg)
	return next, tea.Batch(cmd, cmd2)
}

// start sizes a newly shown screen and runs its Init.
func (a app) start(i int) tea.Cmd {
	var sizeCmd tea.Cmd
	if a.width > 0 {
		a.stack[i], sizeCmd = a.stack[i].Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
	}
	return tea.Batch(a.stack[i].Init(), sizeCmd)
}

func (a app) screenFor(to Screen) screen {
	switch to {
	case ScreenRunAgent:
		return newAgents(a.client, true)
	case ScreenCreateAgent:
		return newAgentForm(a.client)
	case ScreenRuns:
		return newHistory(a.client, "", "")
	default:
		return newAgents(a.client, false)
	}
}

func (a app) updateActive(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(a.stack) == 0 {
		return a.updateHome(msg)
	}
	return a.updateTop(msg)
}

func (a app) updateHome(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.home, cmd = a.home.Update(msg)
	return a, cmd
}

func (a app) updateTop(msg tea.Msg) (tea.Model, tea.Cmd) {
	i := len(a.stack) - 1
	var cmd tea.Cmd
	a.stack[i], cmd = a.stack[i].Update(msg)
	return a, cmd
}

func (a app) View() string {
	if len(a.stack) > 0 {
		return a.stack[len(a.stack)-1].View()
	}
	return a.home.View()
}

// Run starts the full-screen TUI and blocks until the user quits.
func Run(client *api.Client) error {
	_, err := tea.NewProgram(newApp(client), tea.WithAltScreen()).Run()
	return err
}
