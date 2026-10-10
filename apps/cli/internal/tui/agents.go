package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iCoderabhishek/Sutra-AI/internal/api"
)

// Agents screen: list on the left, details of the selected agent on the right.
// In pick mode (home → "Run an agent") enter runs the agent.

type (
	agentsLoadedMsg struct {
		agents []api.Agent
		err    error
	}
	agentChangedMsg struct {
		toast string
		err   error
	}
	runTriggeredMsg struct {
		run   *api.JobRun
		agent api.Agent
		err   error
	}
)

type agentsModel struct {
	client   *api.Client
	pickMode bool
	spin     spinner.Model

	agents        []api.Agent
	cursor        int
	loading       bool
	busy          bool // an action is in flight
	confirmDelete bool
	err           error
	toast         string

	width, height int
}

func newAgents(client *api.Client, pickMode bool) *agentsModel {
	return &agentsModel{
		client:   client,
		pickMode: pickMode,
		spin:     spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(lipgloss.NewStyle().Foreground(colorPrimary))),
		loading:  true,
	}
}

func (m *agentsModel) Init() tea.Cmd { return tea.Batch(m.spin.Tick, m.load()) }

func (m *agentsModel) load() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		agents, err := client.ListAgents()
		if err != nil {
			return agentsLoadedMsg{err: err}
		}
		// The backend soft-deletes agents by setting them INACTIVE, so hide those.
		visible := agents[:0]
		for _, a := range agents {
			if a.Status != api.StatusInactive {
				visible = append(visible, a)
			}
		}
		return agentsLoadedMsg{agents: visible}
	}
}

func (m *agentsModel) selected() (api.Agent, bool) {
	if m.cursor < 0 || m.cursor >= len(m.agents) {
		return api.Agent{}, false
	}
	return m.agents[m.cursor], true
}

func (m *agentsModel) Update(msg tea.Msg) (screen, tea.Cmd) {
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

	case toastMsg:
		m.toast = msg.text

	case agentsLoadedMsg:
		m.loading, m.err = false, msg.err
		if msg.err == nil {
			m.agents = msg.agents
			m.cursor = min(m.cursor, max(len(m.agents)-1, 0))
		}

	case agentChangedMsg:
		m.busy = false
		if msg.err != nil {
			m.toast = "✘ " + errText(msg.err)
			return m, nil
		}
		m.toast = msg.toast
		return m, m.load()

	case runTriggeredMsg:
		m.busy = false
		if msg.err != nil {
			m.toast = "✘ " + errText(msg.err)
			return m, nil
		}
		return m, push(newRunView(m.client, msg.run.ID, msg.agent.Name))

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *agentsModel) handleKey(msg tea.KeyMsg) (screen, tea.Cmd) {
	k := msg.String()
	m.toast = ""

	if m.confirmDelete {
		m.confirmDelete = false
		if a, ok := m.selected(); ok && (k == "y" || k == "Y") {
			m.busy = true
			client := m.client
			return m, func() tea.Msg {
				return agentChangedMsg{toast: "Deleted “" + a.Name + "”", err: client.DeleteAgent(a.ID)}
			}
		}
		return m, nil
	}

	switch k {
	case "esc", "q":
		return m, pop("")
	case "up", "k":
		if len(m.agents) > 0 {
			m.cursor = (m.cursor - 1 + len(m.agents)) % len(m.agents)
		}
		return m, nil
	case "down", "j":
		if len(m.agents) > 0 {
			m.cursor = (m.cursor + 1) % len(m.agents)
		}
		return m, nil
	case "r":
		m.loading = true
		return m, m.load()
	case "n":
		return m, push(newAgentForm(m.client))
	}

	a, ok := m.selected()
	if !ok || m.busy {
		return m, nil
	}

	switch k {
	case "enter", "x":
		if a.Status != api.StatusActive {
			m.toast = "“" + a.Name + "” is paused. Press p to resume it first."
			return m, nil
		}
		m.busy = true
		client := m.client
		return m, func() tea.Msg {
			run, err := client.TriggerRun(a.ID)
			return runTriggeredMsg{run: run, agent: a, err: err}
		}

	case "p":
		next, verb := api.StatusPaused, "Paused"
		if a.Status != api.StatusActive {
			next, verb = api.StatusActive, "Resumed"
		}
		m.busy = true
		client := m.client
		return m, func() tea.Msg {
			_, err := client.SetAgentStatus(a.ID, next)
			return agentChangedMsg{toast: verb + " “" + a.Name + "”", err: err}
		}

	case "d":
		m.confirmDelete = true

	case "h":
		return m, push(newHistory(m.client, a.ID, a.Name))
	}
	return m, nil
}

func (m *agentsModel) View() string {
	title := "My agents"
	if m.pickMode {
		title = "Run an agent"
	}
	w := max(m.width-4, 40)

	var body string
	switch {
	case m.loading && len(m.agents) == 0:
		body = m.spin.View() + " Loading agents…"
	case m.err != nil:
		body = styleError.Render("Couldn't load agents\n\n" + errText(m.err))
	case len(m.agents) == 0:
		body = styleText.Render("No agents yet. Press ") + styleKeyHint.Render("n") + styleText.Render(" to create one.")
	case w >= 90:
		listW := 38
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.listView(listW), " ", m.detailView(w-listW-3))
	default:
		body = lipgloss.JoinVertical(lipgloss.Left, m.listView(w-2), m.detailView(w-2))
	}

	status := ""
	switch {
	case m.confirmDelete:
		a, _ := m.selected()
		status = lipgloss.NewStyle().Foreground(colorDanger).Render("Delete “"+a.Name+"”? ") +
			styleKeyHint.Render("y") + styleMuted.Render(" to confirm, any other key to cancel")
	case m.busy:
		status = m.spin.View() + styleMuted.Render(" Working…")
	case m.toast != "":
		status = styleToast.Render("● " + m.toast)
	}

	keys := keyHints("↑/↓", "select", "enter", "run", "p", "pause/resume", "d", "delete", "h", "history", "n", "new", "esc", "back")
	if m.pickMode {
		keys = keyHints("↑/↓", "select", "enter", "run", "n", "new", "esc", "back")
	}
	return page{Title: title, Body: body, Status: status, Keys: keys}.render(m.width)
}

func (m *agentsModel) listView(inner int) string {
	// Keep the cursor in view on long lists.
	visible := max(m.height-12, 5)
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}
	end := min(start+visible, len(m.agents))

	lines := []string{stylePanelTitle.MarginBottom(0).Render(fmt.Sprintf("Agents (%d)", len(m.agents)))}
	for i := start; i < end; i++ {
		a := m.agents[i]
		badge := statusBadge(a.Status)
		nameW := max(inner-lipgloss.Width(badge)-5, 4)
		name := truncate(a.Name, nameW)
		if i == m.cursor {
			name = styleMenuSelected.Render("▸ " + name)
		} else {
			name = styleText.Render("  " + name)
		}
		gap := max(inner-2-lipgloss.Width(name)-lipgloss.Width(badge), 1)
		lines = append(lines, name+strings.Repeat(" ", gap)+badge)
	}
	if end < len(m.agents) {
		lines = append(lines, styleMuted.Render(fmt.Sprintf("  … %d more", len(m.agents)-end)))
	}
	return stylePanel.Width(inner).Render(strings.Join(lines, "\n"))
}

func (m *agentsModel) detailView(inner int) string {
	a, ok := m.selected()
	if !ok {
		return ""
	}
	textW := max(inner-2, 20)
	wrap := lipgloss.NewStyle().Width(textW)
	field := func(label, value string) string {
		return styleCardLabel.Render(label) + "\n" + wrap.Render(value)
	}

	goal := agentGoal(a)
	if goal == "" {
		goal = lipgloss.NewStyle().Foreground(colorDanger).Render("No goal set: runs will fail.")
	} else {
		goal = styleText.Render(truncate(goal, 600))
	}

	var tools []string
	for _, t := range a.Tools {
		if t.Valid() {
			tools = append(tools, styleText.Render(string(t)))
		} else {
			tools = append(tools, styleToast.Render(string(t)+" (unknown)"))
		}
	}
	toolText := styleMuted.Render("none")
	if len(tools) > 0 {
		toolText = strings.Join(tools, styleMuted.Render(", "))
	}

	schedule := styleMuted.Render("manual only")
	if c := agentCron(a); c != "" {
		schedule = styleText.Render(c)
		if a.Status != api.StatusActive {
			schedule += styleMuted.Render("  (not running while paused)")
		}
	}

	parts := []string{
		styleGreet.Render(a.Name) + "  " + statusBadge(a.Status),
	}
	if a.Desc != nil && *a.Desc != "" {
		parts = append(parts, styleMuted.Render(*a.Desc))
	}
	parts = append(parts, "",
		field("GOAL", goal), "",
		field("TOOLS", toolText), "",
		field("SCHEDULE", schedule), "",
		styleMuted.Render("created "+relTime(a.CreatedAt)+" · "+a.ID),
	)
	return stylePanel.Width(inner).Render(strings.Join(parts, "\n"))
}
