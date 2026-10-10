package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iCoderabhishek/Sutra-AI/internal/api"
)

// Run history: past runs, optionally for one agent. Enter opens the trace.

type runsLoadedMsg struct {
	runs []api.RecentRun
	err  error
}

// Status filters cycled with f. "" means all.
var historyFilters = []api.RunStatus{"", api.RunSucceeded, api.RunFailed, api.RunRunning, api.RunQueued}

type historyModel struct {
	client             *api.Client
	agentID, agentName string // empty = all agents
	spin               spinner.Model

	runs    []api.RecentRun
	cursor  int
	filter  int
	loading bool
	err     error

	width, height int
}

func newHistory(client *api.Client, agentID, agentName string) *historyModel {
	return &historyModel{
		client:    client,
		agentID:   agentID,
		agentName: agentName,
		spin:      spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(lipgloss.NewStyle().Foreground(colorPrimary))),
		loading:   true,
	}
}

func (m *historyModel) Init() tea.Cmd { return tea.Batch(m.spin.Tick, m.load()) }

func (m *historyModel) load() tea.Cmd {
	client, agentID, status := m.client, m.agentID, string(historyFilters[m.filter])
	return func() tea.Msg {
		runs, err := client.ListRuns(agentID, status)
		return runsLoadedMsg{runs, err}
	}
}

func (m *historyModel) Update(msg tea.Msg) (screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case resumedMsg:
		m.loading = true
		return m, m.load()

	case runsLoadedMsg:
		m.loading, m.err = false, msg.err
		if msg.err == nil {
			m.runs = msg.runs
			m.cursor = min(m.cursor, max(len(m.runs)-1, 0))
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, pop("")
		case "up", "k":
			if len(m.runs) > 0 {
				m.cursor = (m.cursor - 1 + len(m.runs)) % len(m.runs)
			}
		case "down", "j":
			if len(m.runs) > 0 {
				m.cursor = (m.cursor + 1) % len(m.runs)
			}
		case "f":
			m.filter = (m.filter + 1) % len(historyFilters)
			m.cursor, m.loading = 0, true
			return m, m.load()
		case "r":
			m.loading = true
			return m, m.load()
		case "enter":
			if m.cursor < len(m.runs) {
				r := m.runs[m.cursor]
				return m, push(newRunView(m.client, r.ID, r.Agent.Name))
			}
		}
	}
	return m, nil
}

func (m *historyModel) View() string {
	title := "Run history"
	if m.agentName != "" {
		title += " · " + m.agentName
	}
	w := max(m.width-6, 40)

	filter := "all"
	if f := historyFilters[m.filter]; f != "" {
		filter = strings.ToLower(string(f))
	}
	header := styleMuted.Render("showing ") + styleKeyHint.Render(filter) + styleMuted.Render(fmt.Sprintf(" · %d runs", len(m.runs)))
	if m.loading {
		header += "  " + m.spin.View()
	}

	var body string
	switch {
	case m.err != nil:
		body = styleError.Render("Couldn't load runs\n\n" + errText(m.err))
	case !m.loading && len(m.runs) == 0:
		body = styleMuted.Render("No runs here yet.")
	default:
		body = m.tableView(w)
	}

	keys := keyHints("↑/↓", "select", "enter", "open trace", "f", "filter", "r", "reload", "esc", "back")
	return page{Title: title, Body: header + "\n\n" + body, Keys: keys}.render(m.width)
}

func (m *historyModel) tableView(w int) string {
	const statusW, whenW, durW, tokW, costW = 22, 10, 8, 9, 9
	nameW := max(w-2-1-2-statusW-whenW-durW-tokW-costW-6, 8) // padding, marker, icon, gaps

	row := func(cols ...string) string {
		widths := []int{2, nameW, statusW, whenW, durW, tokW, costW}
		out := make([]string, len(cols))
		for i, c := range cols {
			out[i] = lipgloss.NewStyle().Width(widths[i]).MaxWidth(widths[i]).Render(c)
		}
		return strings.Join(out, " ")
	}

	lines := []string{" " + styleCardLabel.Render(row("", "AGENT", "STATUS", "WHEN", "TOOK", "TOKENS", "COST"))}

	visible := max(m.height-12, 5)
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}
	end := min(start+visible, len(m.runs))

	for i := start; i < end; i++ {
		r := m.runs[i]
		took := "–"
		if r.StartedAt != nil && r.FinishedAt != nil {
			took = r.FinishedAt.Sub(*r.StartedAt).Round(time.Second).String()
		}
		name := styleText.Render(truncate(r.Agent.Name, nameW))
		marker := " "
		if i == m.cursor {
			marker = styleMenuSelected.Render("▸")
			name = styleMenuSelected.Render(truncate(r.Agent.Name, nameW))
		}
		lines = append(lines, marker+row(runIcon(r.Status), name, strings.ToLower(strings.ReplaceAll(string(r.Status), "_", " ")),
			relTime(r.CreatedAt), took, humanize(r.TotalTokens), fmt.Sprintf("$%.4f", r.TotalCost)))
	}
	if end < len(m.runs) {
		lines = append(lines, styleMuted.Render(fmt.Sprintf("  … %d more", len(m.runs)-end)))
	}
	return stylePanel.Width(w).Render(strings.Join(lines, "\n"))
}
