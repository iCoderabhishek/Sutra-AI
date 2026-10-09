package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iCoderabhishek/Sutra-AI/internal/api"
)

// Create-agent form. Fields are visited with tab / shift+tab (or ↑/↓ outside
// the multi-line goal box); ctrl+s saves from anywhere.

type formField int

const (
	fieldName formField = iota
	fieldDesc
	fieldGoal
	fieldTools
	fieldSchedule
	fieldActive
	fieldSubmit
	fieldCount
)

type agentCreatedMsg struct {
	agent *api.Agent
	err   error
}

type agentForm struct {
	client *api.Client

	name, desc, cron textinput.Model
	goal             textarea.Model
	tools            map[api.ToolName]bool
	toolCursor       int
	active           bool

	focus  formField
	saving bool
	err    string
	width  int
}

func newAgentForm(client *api.Client) *agentForm {
	input := func(placeholder string, limit int) textinput.Model {
		t := textinput.New()
		t.Placeholder = placeholder
		t.CharLimit = limit
		t.Prompt = ""
		return t
	}

	goal := textarea.New()
	goal.Placeholder = "What should this agent do? e.g. Find today's top 5 AI news stories and email me a summary."
	goal.ShowLineNumbers = false
	goal.CharLimit = 4000
	goal.SetHeight(4)
	goal.Prompt = ""

	f := &agentForm{
		client: client,
		name:   input("Daily AI news", 80),
		desc:   input("Optional one-line description", 200),
		cron:   input("Optional, e.g. 0 9 * * *", 64),
		goal:   goal,
		tools:  map[api.ToolName]bool{api.ToolWebSearch: true},
		active: true,
	}
	f.setFocus(fieldName)
	return f
}

func (f *agentForm) Init() tea.Cmd { return textinput.Blink }

func (f *agentForm) setFocus(field formField) {
	f.focus = (field + fieldCount) % fieldCount
	f.name.Blur()
	f.desc.Blur()
	f.cron.Blur()
	f.goal.Blur()
	switch f.focus {
	case fieldName:
		f.name.Focus()
	case fieldDesc:
		f.desc.Focus()
	case fieldGoal:
		f.goal.Focus()
	case fieldSchedule:
		f.cron.Focus()
	}
}

func (f *agentForm) Update(msg tea.Msg) (screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.width = msg.Width
		w := min(max(msg.Width-10, 30), 80)
		f.name.Width, f.desc.Width, f.cron.Width = w, w, w
		f.goal.SetWidth(w)
		return f, nil

	case agentCreatedMsg:
		f.saving = false
		if msg.err != nil {
			f.err = errText(msg.err)
			return f, nil
		}
		return f, pop("Created agent “" + msg.agent.Name + "”")

	case tea.KeyMsg:
		if f.saving {
			return f, nil
		}
		return f.handleKey(msg)
	}
	return f, nil
}

func (f *agentForm) handleKey(msg tea.KeyMsg) (screen, tea.Cmd) {
	inGoal := f.focus == fieldGoal

	switch msg.String() {
	case "esc":
		return f, pop("")
	case "ctrl+s":
		return f.submit()
	case "tab":
		f.setFocus(f.focus + 1)
		return f, nil
	case "shift+tab":
		f.setFocus(f.focus - 1)
		return f, nil
	case "down":
		if !inGoal {
			f.setFocus(f.focus + 1)
			return f, nil
		}
	case "up":
		if !inGoal {
			f.setFocus(f.focus - 1)
			return f, nil
		}
	}

	switch f.focus {
	case fieldTools:
		switch msg.String() {
		case "left", "h":
			f.toolCursor = (f.toolCursor - 1 + len(api.AllTools)) % len(api.AllTools)
		case "right", "l":
			f.toolCursor = (f.toolCursor + 1) % len(api.AllTools)
		case " ", "enter":
			t := api.AllTools[f.toolCursor].Name
			f.tools[t] = !f.tools[t]
		}
		return f, nil

	case fieldActive:
		if s := msg.String(); s == " " || s == "enter" {
			f.active = !f.active
		}
		return f, nil

	case fieldSubmit:
		if msg.String() == "enter" {
			return f.submit()
		}
		return f, nil
	}

	// Enter moves on from single-line inputs; in the goal box it adds a newline.
	if msg.String() == "enter" && !inGoal {
		f.setFocus(f.focus + 1)
		return f, nil
	}

	var cmd tea.Cmd
	switch f.focus {
	case fieldName:
		f.name, cmd = f.name.Update(msg)
	case fieldDesc:
		f.desc, cmd = f.desc.Update(msg)
	case fieldGoal:
		f.goal, cmd = f.goal.Update(msg)
	case fieldSchedule:
		f.cron, cmd = f.cron.Update(msg)
	}
	f.err = ""
	return f, cmd
}

// validate returns the request to send, or a message and the field to fix.
func (f *agentForm) validate() (api.CreateAgentRequest, string, formField) {
	name := strings.TrimSpace(f.name.Value())
	goal := strings.TrimSpace(f.goal.Value())
	cron := strings.Join(strings.Fields(f.cron.Value()), " ")

	switch {
	case name == "":
		return api.CreateAgentRequest{}, "Give the agent a name.", fieldName
	case goal == "":
		return api.CreateAgentRequest{}, "Describe what the agent should do.", fieldGoal
	case cron != "" && !validCron(cron):
		return api.CreateAgentRequest{}, "Schedule must be a cron pattern with 5 or 6 fields, e.g. 0 9 * * *", fieldSchedule
	}

	req := api.CreateAgentRequest{
		Name:   name,
		Desc:   strings.TrimSpace(f.desc.Value()),
		Prompt: map[string]string{"goal": goal},
		Tools:  []api.ToolName{},
		Status: api.StatusPaused,
	}
	for _, t := range api.AllTools {
		if f.tools[t.Name] {
			req.Tools = append(req.Tools, t.Name)
		}
	}
	if cron != "" {
		req.Schedule = map[string]string{"cron": cron}
	}
	// PAUSED, not INACTIVE: the backend marks deleted agents INACTIVE.
	if f.active {
		req.Status = api.StatusActive
	}
	return req, "", 0
}

func (f *agentForm) submit() (screen, tea.Cmd) {
	req, problem, field := f.validate()
	if problem != "" {
		f.err = problem
		f.setFocus(field)
		return f, nil
	}
	f.saving, f.err = true, ""
	client := f.client
	return f, func() tea.Msg {
		agent, err := client.CreateAgent(req)
		return agentCreatedMsg{agent, err}
	}
}

// validCron is a shape check only; the backend scheduler does the real parse.
func validCron(s string) bool {
	n := len(strings.Fields(s))
	return n == 5 || n == 6
}

func (f *agentForm) View() string {
	label := func(field formField, text string) string {
		if f.focus == field {
			return styleMenuSelected.Render("▸ " + text)
		}
		return styleCardLabel.Render("  " + text)
	}
	box := func(field formField, content string) string {
		border := colorBorder
		if f.focus == field {
			border = colorPrimary
		}
		return "  " + stylePanel.BorderForeground(border).Render(content)
	}

	var tools []string
	for i, t := range api.AllTools {
		mark := "[ ]"
		if f.tools[t.Name] {
			mark = "[✓]"
		}
		item := mark + " " + string(t.Name)
		if f.focus == fieldTools && i == f.toolCursor {
			item = styleMenuSelected.Render(item)
		} else {
			item = styleText.Render(item)
		}
		tools = append(tools, item)
	}
	toolDesc := api.AllTools[f.toolCursor].Desc

	activeMark := "[ ]"
	if f.active {
		activeMark = "[✓]"
	}

	button := stylePanel.BorderForeground(colorBorder).Render("Create agent")
	if f.focus == fieldSubmit {
		button = stylePanel.BorderForeground(colorAccent).Foreground(colorAccent).Bold(true).Render("Create agent")
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		label(fieldName, "Name"), box(fieldName, f.name.View()),
		label(fieldDesc, "Description"), box(fieldDesc, f.desc.View()),
		label(fieldGoal, "Goal"), box(fieldGoal, f.goal.View()),
		label(fieldTools, "Tools"),
		"    "+strings.Join(tools, "   ")+"   "+styleMuted.Render(toolDesc),
		"",
		label(fieldSchedule, "Schedule (cron)"), box(fieldSchedule, f.cron.View()),
		"    "+styleMuted.Render("0 9 * * * daily 9:00 · 0 * * * * hourly · 0 9 * * 1 Mondays · empty = manual only"),
		"",
		label(fieldActive, "Status")+"  "+styleText.Render(activeMark+" Active now")+
			styleMuted.Render("  (unchecked = created paused)"),
		"",
		"  "+button,
	)

	status := ""
	switch {
	case f.saving:
		status = styleMuted.Render("Saving…")
	case f.err != "":
		status = lipgloss.NewStyle().Foreground(colorDanger).Render("✘ " + f.err)
	}

	keys := keyHints("tab", "next field", "space", "toggle", "←/→", "pick tool", "ctrl+s", "create", "esc", "cancel")
	return page{Title: "Create agent", Body: body, Status: status, Keys: keys}.render(f.width)
}
