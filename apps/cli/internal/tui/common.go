package tui

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iCoderabhishek/Sutra-AI/internal/api"
)

// ---- navigation ----

// screen is a page shown on top of the home screen.
// The root app keeps them in a stack: push opens one, pop closes it.
type screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (screen, tea.Cmd)
	View() string
}

type (
	pushMsg    struct{ screen screen }
	replaceMsg struct{ screen screen } // swap the top screen, e.g. form → run view
	popMsg     struct{ toast string }  // close the top screen, optionally with a note
	resumedMsg struct{}                // sent to a screen when it is on top again
	toastMsg   struct{ text string }   // short status note shown by the active screen
)

func push(s screen) tea.Cmd         { return send(pushMsg{s}) }
func replace(s screen) tea.Cmd      { return send(replaceMsg{s}) }
func pop(toast string) tea.Cmd      { return send(popMsg{toast}) }
func showToast(text string) tea.Cmd { return send(toastMsg{text}) }

func send(msg tea.Msg) tea.Cmd { return func() tea.Msg { return msg } }

// ---- agent fields ----

// The backend stores prompt and schedule as free-form JSON. These read them
// with the same keys the backend checks (libs/queue.ts, agents/scheduler.ts).
var (
	promptKeys   = []string{"goal", "prompt", "text", "content", "instruction", "value"}
	scheduleKeys = []string{"pattern", "cron"}
)

// agentGoal returns the prompt text, or "" if the worker would find none.
func agentGoal(a api.Agent) string {
	var s string
	if json.Unmarshal(a.Prompt, &s) == nil {
		return strings.TrimSpace(s)
	}
	return firstString(a.Prompt, promptKeys)
}

// agentCron returns the cron pattern, or "" if the agent has no schedule.
func agentCron(a api.Agent) string { return firstString(a.Schedule, scheduleKeys) }

func firstString(raw json.RawMessage, keys []string) string {
	var obj map[string]any
	if json.Unmarshal(raw, &obj) != nil {
		return ""
	}
	for _, k := range keys {
		if v, ok := obj[k].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// ---- layout ----

// page is the shared layout for sub screens: brand and title, body,
// an optional status line, then key hints.
type page struct {
	Title, Body, Status, Keys string
}

func (p page) render(width int) string {
	parts := []string{brandLine() + styleMuted.Render("  ›  ") + styleGreet.Render(p.Title), "", p.Body, ""}
	if p.Status != "" {
		parts = append(parts, p.Status)
	}
	parts = append(parts, p.Keys)
	return stylePage.MaxWidth(max(width, 40)).Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

// brandLine is the one-line logo used where the full logo doesn't fit.
func brandLine() string {
	return styleBrandStart.Render("◆ SUTRA") + styleBrandEnd.Render(" AI") + " " + styleSignature.Render("by Abhishek")
}

// keyHints renders "key label" pairs: keyHints("enter", "run", "esc", "back").
func keyHints(pairs ...string) string {
	hints := make([]string, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		hints = append(hints, styleKeyHint.Render(pairs[i])+" "+styleMuted.Render(pairs[i+1]))
	}
	return strings.Join(hints, styleMuted.Render("  ·  "))
}

func statusBadge(s api.AgentStatus) string {
	switch s {
	case api.StatusActive:
		return styleStatusActive.Render("● active")
	case api.StatusPaused:
		return styleStatusPaused.Render("‖ paused")
	default:
		return styleStatusInactive.Render("○ inactive")
	}
}

// ---- terminal safety ----

var (
	ansiEscape = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|\x1b[@-Z\\-_]`)
	mdLink     = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^)\s]+)\)`)
	mdEmphasis = regexp.MustCompile("(\\*\\*|__|`)(\\S.*?\\S|\\S)(\\*\\*|__|`)")
	mdHeading  = regexp.MustCompile(`^#{1,6}\s+`)
	mdBullet   = regexp.MustCompile(`^[-*+•]\s+`)
)

// sanitize strips escape sequences and control characters from server text,
// so a tool result can't move the cursor, recolor or clear the terminal.
func sanitize(s string) string {
	s = ansiEscape.ReplaceAllString(s, "")
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
}

// stripInline removes markdown emphasis and turns links into "text (url)".
func stripInline(s string) string {
	s = mdLink.ReplaceAllString(s, "$1 ($2)")
	return strings.ReplaceAll(mdEmphasis.ReplaceAllString(s, "$2"), "**", "")
}

// errText prefers the backend's message over the full API error string.
func errText(err error) string {
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Message
	}
	return err.Error()
}
