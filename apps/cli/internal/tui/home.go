package tui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iCoderabhishek/Sutra-AI/internal/api"
)

const autoRefreshEvery = 30 * time.Second

const logo = `███████╗██╗   ██╗████████╗██████╗  █████╗
██╔════╝██║   ██║╚══██╔══╝██╔══██╗██╔══██╗
███████╗██║   ██║   ██║   ██████╔╝███████║
╚════██║██║   ██║   ██║   ██╔══██╗██╔══██║
███████║╚██████╔╝   ██║   ██║  ██║██║  ██║
╚══════╝ ╚═════╝    ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝`

// Screen identifies a destination the home screen can navigate to.
type Screen int

const (
	ScreenRunAgent Screen = iota
	ScreenCreateAgent
	ScreenAgents
	ScreenRuns
)

// NavigateMsg asks the root app to switch screens.
type NavigateMsg struct{ To Screen }

// Internal messages.
type (
	dashboardLoadedMsg struct{ stats *api.DashboardStats }
	dashboardErrMsg    struct{ err error }
	loginDoneMsg       struct{ err error }
	refreshTickMsg     struct{ gen int }
)

type homeState int

const (
	homeLoading homeState = iota
	homeReady
	homeNeedsLogin
	homeLoggingIn
	homeError
)

type menuItem struct {
	title   string
	desc    string
	hint    string
	screen  Screen
	refresh bool // true = refresh the dashboard instead of navigating
}

var homeMenu = []menuItem{
	{title: "Run an agent", desc: "Pick an agent and watch it work live", hint: "x", screen: ScreenRunAgent},
	{title: "Create agent", desc: "Describe a task, choose tools, set a schedule", hint: "n", screen: ScreenCreateAgent},
	{title: "My agents", desc: "Browse, edit, pause or delete agents", hint: "a", screen: ScreenAgents},
	{title: "Run history", desc: "Past runs, traces, cost and tokens", hint: "h", screen: ScreenRuns},
	{title: "Refresh", desc: "Reload stats from the server", hint: "r", refresh: true},
}

type homeKeys struct {
	Up, Down, Select           key.Binding
	Run, New, Agents, History  key.Binding
	Refresh, Login, Help, Quit key.Binding
}

func newHomeKeys() homeKeys {
	return homeKeys{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Select:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Run:     key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "run agent")),
		New:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new agent")),
		Agents:  key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "agents")),
		History: key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "history")),
		Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Login:   key.NewBinding(key.WithKeys("l", "enter"), key.WithHelp("enter", "sign in")),
		Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "more")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func (k homeKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Refresh, k.Help, k.Quit}
}

func (k homeKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Select},
		{k.Run, k.New, k.Agents, k.History},
		{k.Refresh, k.Help, k.Quit},
	}
}

type homeModel struct {
	client *api.Client
	keys   homeKeys
	help   help.Model
	spin   spinner.Model

	state       homeState
	stats       *api.DashboardStats
	err         error
	toast       string
	cursor      int
	dense       int  // layout density, set per render by View
	clip        bool // last resort: cut the body to fit under the logo
	refreshing  bool
	refreshGen  int
	lastUpdated time.Time

	width, height int
}

func newHome(client *api.Client) homeModel {
	sp := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(colorPrimary)),
	)
	return homeModel{
		client: client,
		keys:   newHomeKeys(),
		help:   help.New(),
		spin:   sp,
		state:  homeLoading,
	}
}

func (m homeModel) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick, fetchDashboard(m.client))
}

// ---- commands ----

func fetchDashboard(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		stats, err := c.GetDashboard()
		if err != nil {
			return dashboardErrMsg{err}
		}
		return dashboardLoadedMsg{stats}
	}
}

func login(c *api.Client) tea.Cmd {
	return func() tea.Msg { return loginDoneMsg{c.Login()} }
}

// scheduleRefresh bumps the generation so older pending ticks are ignored.
func (m *homeModel) scheduleRefresh() tea.Cmd {
	m.refreshGen++
	gen := m.refreshGen
	return tea.Tick(autoRefreshEvery, func(time.Time) tea.Msg { return refreshTickMsg{gen} })
}

func isUnauthorized(err error) bool {
	var apiErr *api.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 401
}

// ---- update ----

func (m homeModel) Update(msg tea.Msg) (homeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.Width = msg.Width
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case dashboardLoadedMsg:
		m.stats, m.err = msg.stats, nil
		m.state, m.refreshing = homeReady, false
		m.lastUpdated = time.Now()
		return m, m.scheduleRefresh()

	case dashboardErrMsg:
		m.refreshing = false
		switch {
		case isUnauthorized(msg.err):
			m.state = homeNeedsLogin
			return m, nil
		case m.stats != nil:
			// Keep showing stale data, surface the error quietly.
			m.toast = "Couldn't refresh: server unreachable"
			return m, m.scheduleRefresh()
		default:
			m.state, m.err = homeError, msg.err
			return m, nil
		}

	case refreshTickMsg:
		if msg.gen != m.refreshGen || m.state != homeReady || m.refreshing {
			return m, nil
		}
		m.refreshing = true
		return m, fetchDashboard(m.client)

	case loginDoneMsg:
		// Login prints to stdout, so repaint the whole screen afterwards.
		if msg.err != nil {
			m.state, m.err = homeError, msg.err
			return m, tea.ClearScreen
		}
		m.state = homeLoading
		return m, tea.Batch(tea.ClearScreen, fetchDashboard(m.client))

	case toastMsg:
		m.toast = msg.text
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m homeModel) handleKey(msg tea.KeyMsg) (homeModel, tea.Cmd) {
	m.toast = ""

	if key.Matches(msg, m.keys.Help) {
		m.help.ShowAll = !m.help.ShowAll
		return m, nil
	}

	switch m.state {
	case homeNeedsLogin:
		if key.Matches(msg, m.keys.Login) {
			m.state = homeLoggingIn
			return m, login(m.client)
		}

	case homeError:
		if key.Matches(msg, m.keys.Refresh) {
			m.state, m.err = homeLoading, nil
			return m, fetchDashboard(m.client)
		}

	case homeReady:
		switch {
		case key.Matches(msg, m.keys.Up):
			m.cursor = (m.cursor - 1 + len(homeMenu)) % len(homeMenu)
		case key.Matches(msg, m.keys.Down):
			m.cursor = (m.cursor + 1) % len(homeMenu)
		case key.Matches(msg, m.keys.Select):
			return m.activate(homeMenu[m.cursor])
		default:
			for i, item := range homeMenu {
				if msg.String() == item.hint {
					m.cursor = i
					return m.activate(item)
				}
			}
		}
	}
	return m, nil
}

func (m homeModel) activate(item menuItem) (homeModel, tea.Cmd) {
	if item.refresh {
		if m.refreshing {
			return m, nil
		}
		m.refreshing = true
		return m, fetchDashboard(m.client)
	}
	screen := item.screen
	return m, func() tea.Msg { return NavigateMsg{To: screen} }
}

// ---- view ----

// View keeps the full logo on every screen. When the terminal is too short,
// it tightens the content below the logo step by step instead (Bubble Tea
// drops the top lines of an over-tall view, which would cut the logo off).
// The one-line header is only a last resort for very small terminals.
func (m homeModel) View() string {
	for m.dense = 0; m.dense <= maxDense; m.dense++ {
		out := m.render(m.headerView(m.contentWidth()))
		if m.height == 0 || lipgloss.Height(out) <= m.height {
			return out
		}
	}
	// Still too tall: keep the logo and footer, cut the dashboard from the bottom.
	m.dense = maxDense
	m.clip = true
	return m.render(m.headerView(m.contentWidth()))
}

// clipLines keeps the first n lines of s.
func clipLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if n < 0 {
		n = 0
	}
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n")
}

// Density levels used by View:
//
//	0 = roomy, 1 = no spacer lines or title margins,
//	2 = also hide menu descriptions, 3 recent runs, one-line banner.
const maxDense = 2

func (m homeModel) titleStyle() lipgloss.Style {
	if m.dense >= 1 {
		return stylePanelTitle.MarginBottom(0)
	}
	return stylePanelTitle
}

func (m homeModel) render(header string) string {
	w := m.contentWidth()

	var body string
	switch m.state {
	case homeLoading:
		body = m.centerNote(m.spin.View() + " Loading your workspace…")
	case homeNeedsLogin:
		body = m.loginView(w)
	case homeLoggingIn:
		body = m.centerNote(m.spin.View() + " Waiting for you to finish signing in in the browser…\n" +
			styleMuted.Render("A browser window should have opened. This screen updates automatically."))
	case homeError:
		body = lipgloss.JoinVertical(lipgloss.Left,
			styleError.Width(min(w-2, 70)).Render("Something went wrong\n\n"+m.err.Error()),
			"",
			styleMuted.Render("Is the backend running? Press ")+styleKeyHint.Render("r")+styleMuted.Render(" to retry."),
		)
	case homeReady:
		body = m.dashboardView(w)
	}

	footer := m.help.View(m.keys)
	if m.toast != "" {
		footer = styleToast.Render("● "+m.toast) + "\n" + footer
	}

	if m.clip && m.height > 0 {
		body = clipLines(body, m.height-lipgloss.Height(header)-lipgloss.Height(footer))
	}

	if m.dense >= 1 {
		return styleApp.Padding(0, 2).Render(lipgloss.JoinVertical(lipgloss.Left, header, body, footer))
	}
	return styleApp.Render(lipgloss.JoinVertical(lipgloss.Left, header, "", body, "", footer))
}

func (m homeModel) contentWidth() int {
	if m.width == 0 {
		return 100
	}
	return max(m.width-4, 40) // styleApp horizontal padding
}

func (m homeModel) centerNote(s string) string {
	return lipgloss.NewStyle().Padding(1, 0).Render(s)
}

func (m homeModel) headerView(w int) string {
	if w < 46 {
		return m.compactHeaderView()
	}

	lines := strings.Split(logo, "\n")
	for i, l := range lines {
		lines[i] = lipgloss.NewStyle().Foreground(logoGradient[i%len(logoGradient)]).Bold(true).Render(l)
	}
	logoBlock := strings.Join(lines, "\n")
	// Right-align the signature under the logo.
	byline := lipgloss.PlaceHorizontal(lipgloss.Width(logoBlock), lipgloss.Right, styleSignature.Render("by Abhishek"))
	brand := lipgloss.JoinVertical(lipgloss.Left, logoBlock, byline)

	const tagline = "your autonomous agents, one keystroke away"

	// Wide terminals: greeting sits beside the logo, saving three rows.
	info := lipgloss.JoinVertical(lipgloss.Left,
		styleGreet.Render(greeting()),
		styleTagline.Render(tagline),
		"",
		m.statusLine(),
	)
	if gap := 4; w >= lipgloss.Width(brand)+gap+lipgloss.Width(info) {
		return lipgloss.JoinHorizontal(lipgloss.Center, brand, strings.Repeat(" ", gap), info)
	}

	greet := styleGreet.Render(greeting()) + styleMuted.Render(" · ") + styleTagline.Render(tagline)
	if m.dense >= 1 {
		return lipgloss.JoinVertical(lipgloss.Left, brand, greet, m.statusLine())
	}
	return lipgloss.JoinVertical(lipgloss.Left, brand, "", greet, m.statusLine())
}

// compactHeaderView is the one-line brand used when vertical space is tight.
func (m homeModel) compactHeaderView() string {
	line := brandLine() + styleMuted.Render("  ·  ") + styleGreet.Render(greeting())
	if s := m.statusLine(); s != "" {
		line += styleMuted.Render("  ·  ") + s
	}
	return line
}

func (m homeModel) statusLine() string {
	if m.state != homeReady {
		return ""
	}
	if m.refreshing {
		return m.spin.View() + styleMuted.Render(" syncing")
	}
	return styleMuted.Render("● live · updated " + relTime(m.lastUpdated))
}

func (m homeModel) loginView(w int) string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		stylePanelTitle.Render("Welcome to Sutra"),
		styleText.Render("Build agents that search, scrape and email for you,\nthen run them on demand or on a schedule."),
		"",
		styleMuted.Render("Press ")+styleKeyHint.Render("enter")+styleMuted.Render(" to sign in with Google."),
	)
	return stylePanel.Width(min(w-2, 64)).Render(content)
}

func (m homeModel) dashboardView(w int) string {
	s := m.stats
	sections := []string{m.cardsView(w)}

	if s.Agents.Total == 0 {
		text := styleKeyHint.Render("✦ Get started  ") +
			styleText.Render("You don't have any agents yet. Press ") +
			styleKeyHint.Render("n") + styleText.Render(" to create your first one.")
		if m.dense >= 2 {
			sections = append(sections, " "+text)
		} else {
			sections = append(sections, stylePanel.BorderForeground(colorAccent).Width(w-2).Render(text))
		}
	}

	const menuOuter = 42
	if w >= 92 {
		runsOuter := w - menuOuter - 1
		sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
			m.menuView(menuOuter-2), " ", m.recentRunsView(runsOuter-2)))
	} else {
		sections = append(sections, m.menuView(w-2), m.recentRunsView(w-2))
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m homeModel) cardsView(w int) string {
	s := m.stats

	plan := "free plan"
	if s.Credits.Plan != nil && *s.Credits.Plan != "" {
		plan = strings.ToLower(*s.Credits.Plan) + " plan"
	}
	creditColor := colorSuccess
	if s.Credits.Balance < 5 {
		creditColor = colorWarning
	}
	if s.Credits.Balance <= 0 {
		creditColor = colorDanger
	}

	succeeded := s.Runs.ByStatus[api.RunSucceeded]
	failed := s.Runs.ByStatus[api.RunFailed]
	runsSub := "no runs yet"
	if succeeded+failed > 0 {
		runsSub = fmt.Sprintf("%d%% success rate", succeeded*100/(succeeded+failed))
	}
	if running := s.Runs.ByStatus[api.RunRunning]; running > 0 {
		runsSub = fmt.Sprintf("%d running now", running)
	}

	type cardData struct {
		label, value, sub string
		color             lipgloss.TerminalColor
	}
	cards := []cardData{
		{"CREDITS", fmt.Sprintf("%.2f", s.Credits.Balance), plan, creditColor},
		{"AGENTS", fmt.Sprintf("%d", s.Agents.Total),
			fmt.Sprintf("%d active · %d paused", s.Agents.ByStatus[api.StatusActive], s.Agents.ByStatus[api.StatusPaused]), colorPrimary},
		{"RUNS", fmt.Sprintf("%d", s.Runs.Total), runsSub, colorAccent},
		{"SPEND", fmt.Sprintf("$%.2f", s.Usage.TotalCostUsd), humanize(s.Usage.TotalTokens) + " tokens", colorText},
	}

	perRow := 4
	if w < 80 {
		perRow = 2
	}
	// Each card adds 2 columns of border; cards are separated by 1 space.
	inner := (w-(perRow-1))/perRow - 2

	var rows, row []string
	for i, c := range cards {
		body := lipgloss.JoinVertical(lipgloss.Left,
			styleCardLabel.Render(c.label),
			styleCardValue.Foreground(c.color).Render(c.value),
			styleCardSub.Render(truncate(c.sub, inner-2)),
		)
		if len(row) > 0 {
			row = append(row, " ")
		}
		row = append(row, stylePanel.Width(inner).Render(body))
		if (i+1)%perRow == 0 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, row...))
			row = nil
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m homeModel) menuView(inner int) string {
	lines := []string{m.titleStyle().Render("Quick actions")}
	textW := inner - 2 // stylePanel padding

	for i, item := range homeMenu {
		hint := styleKeyHint.Render(item.hint)
		var left string
		if i == m.cursor {
			left = styleMenuSelected.Render("▸ " + item.title)
		} else {
			left = styleMenuItem.Render(item.title)
		}
		gap := max(textW-lipgloss.Width(left)-lipgloss.Width(hint), 1)
		lines = append(lines, left+strings.Repeat(" ", gap)+hint)
		if i == m.cursor && m.dense < 2 {
			lines = append(lines, styleMenuDesc.Render(truncate(item.desc, textW-4)))
		}
	}
	return stylePanel.Width(inner).Render(strings.Join(lines, "\n"))
}

func (m homeModel) recentRunsView(inner int) string {
	lines := []string{m.titleStyle().Render("Recent runs")}
	limit := 6
	if m.dense >= 2 {
		limit = 3
	}
	textW := inner - 2

	runs := m.stats.RecentRuns
	if len(runs) == 0 {
		lines = append(lines, styleMuted.Render("Nothing here yet. Runs show up the moment an agent starts."))
	}
	for i, r := range runs {
		if i == limit {
			break
		}
		icon := runIcon(r.Status)
		meta := styleMuted.Render(fmt.Sprintf("%s · $%.3f", relTime(r.CreatedAt), r.TotalCost))
		nameW := max(textW-lipgloss.Width(icon)-lipgloss.Width(meta)-2, 4)
		name := styleText.Render(truncate(r.Agent.Name, nameW))
		gap := max(textW-lipgloss.Width(icon)-1-lipgloss.Width(name)-lipgloss.Width(meta), 1)
		lines = append(lines, icon+" "+name+strings.Repeat(" ", gap)+meta)
	}
	return stylePanel.Width(inner).Render(strings.Join(lines, "\n"))
}

// ---- helpers ----

func runIcon(s api.RunStatus) string {
	switch s {
	case api.RunSucceeded:
		return lipgloss.NewStyle().Foreground(colorSuccess).Render("✔")
	case api.RunFailed:
		return lipgloss.NewStyle().Foreground(colorDanger).Render("✘")
	case api.RunRunning:
		return lipgloss.NewStyle().Foreground(colorPrimary).Render("●")
	case api.RunInsufficientCredits:
		return lipgloss.NewStyle().Foreground(colorWarning).Render("!")
	default:
		return lipgloss.NewStyle().Foreground(colorMuted).Render("◌")
	}
}

func greeting() string {
	switch h := time.Now().Hour(); {
	case h < 12:
		return "Good morning"
	case h < 17:
		return "Good afternoon"
	default:
		return "Good evening"
	}
}

func relTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < 10*time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func humanize(n int) string {
	switch {
	case n < 1_000:
		return fmt.Sprintf("%d", n)
	case n < 1_000_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
}

func truncate(s string, n int) string {
	r := []rune(s)
	if n <= 1 || len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
