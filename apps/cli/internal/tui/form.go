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
	fieldModel
	fieldTools
	fieldTelegram // only shown when the notify tool is checked
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

	name, desc, cron, telegram textinput.Model
	goal                       textarea.Model

	models      []api.LLMModel
	modelCursor int

	tools      map[api.ToolName]bool
	toolCursor int
	active     bool

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
	goal.CharLimit = 2000 // matches the runtime's input guardrail
	goal.SetHeight(3)
	goal.Prompt = ""
	goal.FocusedStyle.CursorLine = lipgloss.NewStyle()
	goal.KeyMap.InsertNewline.SetKeys("alt+enter")

	models, err := client.GetLLMModels()
	if err != nil || len(models) == 0 {
		models = []api.LLMModel{{ID: "auto", Name: "Auto (Best Available)"}}
	}

	f := &agentForm{
		client:   client,
		name:     input("Daily AI news", 80),
		desc:     input("Optional one-line description", 200),
		cron:     input("Optional, e.g. 0 9 * * *", 64),
		telegram: input("e.g. 123456789 (message the bot, then ask @userinfobot)", 32),
		goal:     goal,
		models:   models,
		tools:    map[api.ToolName]bool{api.ToolWebSearch: true},
		active:   true,
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
	f.telegram.Blur()
	f.goal.Blur()
	switch f.focus {
	case fieldName:
		f.name.Focus()
	case fieldDesc:
		f.desc.Focus()
	case fieldGoal:
		f.goal.Focus()
	case fieldTelegram:
		f.telegram.Focus()
	case fieldSchedule:
		f.cron.Focus()
	}
}

// move steps focus by delta, skipping the Telegram field while notify is off.
func (f *agentForm) move(delta formField) {
	next := (f.focus + delta + fieldCount) % fieldCount
	if next == fieldTelegram && !f.tools[api.ToolNotify] {
		next = (next + delta + fieldCount) % fieldCount
	}
	f.setFocus(next)
}

func (f *agentForm) Update(msg tea.Msg) (screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.width = msg.Width
		w := min(max(msg.Width-10, 30), 80)
		f.name.Width, f.desc.Width, f.cron.Width, f.telegram.Width = w, w, w, w
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
		f.move(1)
		return f, nil
	case "shift+tab":
		f.move(-1)
		return f, nil
	case "down":
		// In the goal box, ↓ moves the cursor until the last line, then leaves.
		if !inGoal || f.goal.Line() >= f.goal.LineCount()-1 {
			f.move(1)
			return f, nil
		}
	case "up":
		if !inGoal || f.goal.Line() == 0 {
			f.move(-1)
			return f, nil
		}
	}

	switch f.focus {
	case fieldModel:
		switch msg.String() {
		case "left", "h":
			f.modelCursor = (f.modelCursor - 1 + len(f.models)) % len(f.models)
		case "right", "l":
			f.modelCursor = (f.modelCursor + 1) % len(f.models)
		}
		return f, nil

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

	// Enter moves on from text fields (alt+enter adds a line in the goal box).
	if msg.String() == "enter" {
		f.move(1)
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
	case fieldTelegram:
		f.telegram, cmd = f.telegram.Update(msg)
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
	chatID := strings.TrimSpace(f.telegram.Value())
	useNotify := f.tools[api.ToolNotify]

	switch {
	case name == "":
		return api.CreateAgentRequest{}, "Give the agent a name.", fieldName
	case goal == "":
		return api.CreateAgentRequest{}, "Describe what the agent should do.", fieldGoal
	case useNotify && !validChatID(chatID):
		return api.CreateAgentRequest{}, "Telegram alerts need your numeric chat ID, e.g. 123456789.", fieldTelegram
	case cron != "" && !validCron(cron):
		return api.CreateAgentRequest{}, "Schedule must be a cron pattern with 5 or 6 fields, e.g. 0 9 * * *", fieldSchedule
	}

	prompt := map[string]string{"goal": goal}
	if useNotify {
		prompt["telegram_chat_id"] = chatID // read by the backend worker, not shown to the model
	}

	req := api.CreateAgentRequest{
		Name:   name,
		Desc:   strings.TrimSpace(f.desc.Value()),
		Prompt: prompt,
		Model:  f.models[f.modelCursor].ID,
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

// validChatID accepts Telegram chat IDs: digits, with a leading "-" for groups.
func validChatID(s string) bool {
	s = strings.TrimPrefix(s, "-")
	if len(s) < 5 || len(s) > 20 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
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
		// MarginLeft indents every line; a "  " prefix only shifted the top border.
		return stylePanel.BorderForeground(border).MarginLeft(2).Render(content)
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

	button := stylePanel.BorderForeground(colorBorder).MarginLeft(2).Render("Create agent")
	if f.focus == fieldSubmit {
		button = stylePanel.BorderForeground(colorAccent).Foreground(colorAccent).Bold(true).MarginLeft(2).Render("▸ Create agent  (enter)")
	}

	modelName := "Auto (Best Available)"
	if len(f.models) > 0 {
		modelName = f.models[f.modelCursor].Name
	}
	modelSelector := "← " + modelName + " →"
	if f.focus == fieldModel {
		modelSelector = styleMenuSelected.Render(modelSelector)
	} else {
		modelSelector = styleText.Render(modelSelector)
	}

	telegramSection := ""
	if f.tools[api.ToolNotify] {
		telegramSection = lipgloss.JoinVertical(lipgloss.Left,
			label(fieldTelegram, "Telegram chat ID"), box(fieldTelegram, f.telegram.View()),
			"    "+styleMuted.Render("Send /start to the Sutra bot first, or it can't message you"),
		)
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		label(fieldName, "Name"), box(fieldName, f.name.View()),
		label(fieldDesc, "Description"), box(fieldDesc, f.desc.View()),
		label(fieldGoal, "Goal"), box(fieldGoal, f.goal.View()),
		label(fieldModel, "Model"),
		"    "+modelSelector,
		label(fieldTools, "Tools"),
		lipgloss.NewStyle().PaddingLeft(4).Width(max(f.width-8, 40)).Render(strings.Join(tools, "   ")),
		"    "+styleMuted.Render("› "+toolDesc),
		telegramSection,
		label(fieldSchedule, "Schedule (cron)"), box(fieldSchedule, f.cron.View()),
		"    "+styleMuted.Render("0 9 * * * daily 9:00 · 0 * * * * hourly · 0 9 * * 1 Mondays · empty = manual only"),
		label(fieldActive, "Status")+"  "+styleText.Render(activeMark+" Active now")+
			styleMuted.Render("  (unchecked = created paused)"),
		"",
		button,
	)

	status := ""
	switch {
	case f.saving:
		status = styleMuted.Render("Saving…")
	case f.err != "":
		status = lipgloss.NewStyle().Foreground(colorDanger).Render("✘ " + f.err)
	}

	keys := keyHints("↑/↓ or tab", "move", "enter", "next / create", "space", "toggle", "←/→", "pick tool", "esc", "cancel")
	return page{Title: "Create agent", Body: body, Status: status, Keys: keys}.render(f.width)
}
