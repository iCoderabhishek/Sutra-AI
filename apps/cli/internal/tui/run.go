package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iCoderabhishek/Sutra-AI/internal/api"
)

// Run view: streams a run's trace live. For finished runs the backend replays
// the saved trace over the same stream, so history uses this view too.

type (
	traceEventMsg struct {
		runID string
		event api.TraceEvent
	}
	traceEndMsg struct {
		runID string
		err   error
	}
	runFinalMsg struct {
		runID string
		run   *api.JobRun
		err   error
	}
)

type runModel struct {
	client    *api.Client
	runID     string
	agentName string

	cancel context.CancelFunc
	events <-chan api.TraceEvent
	errs   <-chan error

	trace   []api.TraceEvent
	tokens  int
	cost    float64
	started time.Time
	ended   time.Time

	streaming bool
	streamErr error
	final     *api.JobRun

	vp            viewport.Model
	spin          spinner.Model
	width, height int
}

// newRunView opens the stream right away; Init starts reading it.
func newRunView(client *api.Client, runID, agentName string) *runModel {
	ctx, cancel := context.WithCancel(context.Background())
	events, errs := client.StreamRunLogs(ctx, runID)
	return &runModel{
		client:    client,
		runID:     runID,
		agentName: agentName,
		cancel:    cancel,
		events:    events,
		errs:      errs,
		started:   time.Now(),
		streaming: true,
		vp:        viewport.New(80, 10),
		spin:      spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(colorPrimary))),
	}
}

func (m *runModel) Init() tea.Cmd { return tea.Batch(m.spin.Tick, m.next()) }

// next waits for one stream event. Update calls it again after every event,
// which keeps the stream flowing without blocking the UI.
func (m *runModel) next() tea.Cmd {
	runID, events, errs := m.runID, m.events, m.errs
	return func() tea.Msg {
		ev, ok := <-events
		if ok {
			return traceEventMsg{runID, ev}
		}
		return traceEndMsg{runID, <-errs} // errs is buffered and closed, so this never blocks
	}
}

func (m *runModel) fetchFinal() tea.Cmd {
	client, runID := m.client, m.runID
	return func() tea.Msg {
		run, err := client.GetRun(runID)
		return runFinalMsg{runID, run, err}
	}
}

func (m *runModel) Update(msg tea.Msg) (screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.vp.Width = max(msg.Width-4, 20)
		m.vp.Height = max(msg.Height-11, 5)
		m.refresh(true)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case traceEventMsg:
		if msg.runID != m.runID {
			return m, nil
		}
		m.add(msg.event)
		m.refresh(m.vp.AtBottom())
		return m, m.next()

	case traceEndMsg:
		if msg.runID != m.runID {
			return m, nil
		}
		m.streaming, m.streamErr, m.ended = false, msg.err, time.Now()
		return m, m.fetchFinal()

	case runFinalMsg:
		if msg.runID == m.runID && msg.err == nil {
			m.final = msg.run
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.cancel()
			return m, pop("")
		case "g":
			m.vp.GotoTop()
			return m, nil
		case "G":
			m.vp.GotoBottom()
			return m, nil
		}
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	return m, nil
}

// add appends an event. A "done"/"error" event for the step that is still
// "running" replaces it, so each step shows once with its final state.
func (m *runModel) add(ev api.TraceEvent) {
	if ev.Cost != nil {
		m.tokens += ev.Cost.TotalTokens
		m.cost += ev.Cost.CostUsd
	}
	if n := len(m.trace); n > 0 {
		last := m.trace[n-1]
		if last.Status == api.TraceEventRunning && ev.Status != api.TraceEventRunning &&
			last.Step == ev.Step && sameIteration(last.Iteration, ev.Iteration) {
			m.trace[n-1] = ev
			return
		}
	}
	m.trace = append(m.trace, ev)
}

func sameIteration(a, b *int) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}

func (m *runModel) refresh(follow bool) {
	m.vp.SetContent(m.renderTrace(m.vp.Width))
	if follow {
		m.vp.GotoBottom()
	}
}

func (m *runModel) renderTrace(width int) string {
	if len(m.trace) == 0 {
		return styleMuted.Render("Waiting for the agent to start…")
	}
	wrap := lipgloss.NewStyle().Width(max(width-4, 10)).PaddingLeft(4)
	var blocks []string
	for _, ev := range m.trace {
		// The final answer gets the full report layout instead of a step block.
		if ev.Step == "Final Answer" && ev.Content != nil {
			if r, ok := api.ParseReport(*ev.Content); ok {
				blocks = append(blocks, renderReport(r, width-2))
			} else {
				blocks = append(blocks, sectionRule("ANSWER", width-2)+"\n"+renderPlainAnswer(*ev.Content, width-2))
			}
			continue
		}

		head := traceIcon(ev.Status) + " " + styleGreet.Render(sanitize(ev.Step))
		if ev.Iteration != nil {
			head += styleMuted.Render(fmt.Sprintf("  · step %d", *ev.Iteration))
		}
		if ev.Cost != nil {
			head += styleMuted.Render(fmt.Sprintf("  · %s tokens · $%.4f", humanize(ev.Cost.TotalTokens), ev.Cost.CostUsd))
		}
		lines := []string{head}
		if len(ev.Args) > 0 {
			lines = append(lines, wrap.Render(styleKeyHint.Render(sanitize(formatArgs(ev.Args)))))
		}
		if ev.Content != nil && strings.TrimSpace(*ev.Content) != "" {
			text := styleText
			if ev.Status == api.TraceEventError {
				text = lipgloss.NewStyle().Foreground(colorDanger)
			}
			lines = append(lines, wrap.Render(text.Render(stripInline(sanitize(strings.TrimSpace(*ev.Content))))))
		}
		if ev.ResultPreview != nil && strings.TrimSpace(*ev.ResultPreview) != "" {
			lines = append(lines, wrap.Render(styleMuted.Render("→ "+firstLines(sanitize(strings.TrimSpace(*ev.ResultPreview)), 6))))
		}
		blocks = append(blocks, strings.Join(lines, "\n"))
	}
	return strings.Join(blocks, "\n\n")
}

func (m *runModel) View() string {
	elapsed := time.Since(m.started)
	if !m.streaming {
		elapsed = m.ended.Sub(m.started)
	}
	tokens, cost := m.tokens, m.cost
	if m.final != nil {
		tokens, cost = m.final.TotalTokens, m.final.TotalCost
		if m.final.StartedAt != nil && m.final.FinishedAt != nil {
			elapsed = m.final.FinishedAt.Sub(*m.final.StartedAt)
		}
	}

	var state string
	switch {
	case m.streaming:
		state = m.spin.View() + lipgloss.NewStyle().Foreground(colorPrimary).Render(" running")
	case m.final != nil:
		state = runIcon(m.final.Status) + " " + styleText.Render(strings.ToLower(strings.ReplaceAll(string(m.final.Status), "_", " ")))
	default:
		state = styleMuted.Render("stream ended")
	}

	summary := lipgloss.JoinHorizontal(lipgloss.Top,
		state,
		styleMuted.Render(fmt.Sprintf("   %s   %s tokens   $%.4f   run %s",
			elapsed.Round(time.Second), humanize(tokens), cost, shortID(m.runID))),
	)
	body := lipgloss.JoinVertical(lipgloss.Left,
		summary, "",
		stylePanel.Width(max(m.width-6, 20)).Render(m.vp.View()),
	)

	status := ""
	if m.streamErr != nil {
		status = lipgloss.NewStyle().Foreground(colorDanger).Render("✘ " + m.streamErr.Error())
	} else if !m.vp.AtBottom() && m.streaming {
		status = styleToast.Render("● new steps below · press G to follow")
	}

	keys := keyHints("↑/↓", "scroll", "g/G", "top/bottom", "esc", "back")
	return page{Title: m.agentName, Body: body, Status: status, Keys: keys}.render(m.width)
}

func traceIcon(s api.TraceEventStatus) string {
	switch s {
	case api.TraceEventDone:
		return lipgloss.NewStyle().Foreground(colorSuccess).Render("✔")
	case api.TraceEventError:
		return lipgloss.NewStyle().Foreground(colorDanger).Render("✘")
	default:
		return lipgloss.NewStyle().Foreground(colorPrimary).Render("●")
	}
}

// formatArgs renders tool arguments as key=value, sorted for stable output.
func formatArgs(args map[string]any) string {
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v, _ := json.Marshal(args[k])
		parts = append(parts, k+"="+truncate(string(v), 120))
	}
	return strings.Join(parts, "  ")
}

func firstLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n") + "\n…"
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
