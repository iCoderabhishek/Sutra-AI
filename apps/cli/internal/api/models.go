package api

import (
	"encoding/json"
	"fmt"
	"time"
)

type AgentStatus string

const (
	StatusActive   AgentStatus = "ACTIVE"
	StatusInactive AgentStatus = "INACTIVE"
	StatusPaused   AgentStatus = "PAUSED"
)

type RunStatus string

const (
	RunQueued              RunStatus = "QUEUED"
	RunRunning             RunStatus = "RUNNING"
	RunSucceeded           RunStatus = "SUCCEEDED"
	RunFailed              RunStatus = "FAILED"
	RunInsufficientCredits RunStatus = "INSUFFICIENT_CREDITS"
)

// Agent represents the Agent model from the database.
type Agent struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Desc        *string         `json:"desc"`
	UserID      *string         `json:"userId"`
	Prompt      json.RawMessage `json:"prompt"`
	Template    *string         `json:"template"`
	Instruction json.RawMessage `json:"instruction"`
	Model       *string         `json:"model"`
	Tools       []ToolName      `json:"tools"` // may contain names this CLI doesn't know; check Valid()
	Schedule    json.RawMessage `json:"schedule"`
	Status      AgentStatus     `json:"status"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type JobRun struct {
	ID          string          `json:"id"`
	AgentID     string          `json:"agentId"`
	Status      RunStatus       `json:"status"`
	Trace       json.RawMessage `json:"trace"` // Stores the array of event traces
	TotalCost   float64         `json:"totalCost"`
	TotalTokens int             `json:"totalTokens"`
	StartedAt   *time.Time      `json:"startedAt"`
	FinishedAt  *time.Time      `json:"finishedAt"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type Credits struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Balance   string    `json:"balance"` // Decimal as string to preserve precision
	Plan      string    `json:"plan"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// DashboardStats represents the response from /api/v1/dashboard
// actually this is what returns from backend -
//
//	agents: {
//	                total: agentsByStatus.reduce((sum, row) => sum + row._count._all, 0),
//	                byStatus: countOf(agentsByStatus as never),
//	            },
//	            runs: {
//	                total: totals._count._all,
//	                byStatus: countOf(runsByStatus as never),
//	            },
//	            usage: {
//	                totalCostUsd: Number(totals._sum.totalCost ?? 0),
//	                totalTokens: totals._sum.totalTokens ?? 0,
//	            },
//	            credits: {
//	                balance: credits ? Number(credits.balance) : 0,
//	                plan: credits?.plan ?? null,
//	            },
//	            recentRuns: recentRuns.map((run) => ({ ...run, totalCost: Number(run.totalCost) })),

type DashboardStats struct {
	Agents struct {
		Total    int                 `json:"total"`
		ByStatus map[AgentStatus]int `json:"byStatus"`
	} `json:"agents"`
	Runs struct {
		Total    int               `json:"total"`
		ByStatus map[RunStatus]int `json:"byStatus"`
	} `json:"runs"`
	Usage struct {
		TotalCostUsd float64 `json:"totalCostUsd"`
		TotalTokens  int     `json:"totalTokens"`
	} `json:"usage"`
	Credits struct {
		Balance float64 `json:"balance"`
		Plan    *string `json:"plan"`
	} `json:"credits"`
	RecentRuns []RecentRun `json:"recentRuns"`
}

type RecentRun struct {
	ID          string     `json:"id"`
	Status      RunStatus  `json:"status"`
	TotalCost   float64    `json:"totalCost"`
	TotalTokens int        `json:"totalTokens"`
	StartedAt   *time.Time `json:"startedAt"`
	FinishedAt  *time.Time `json:"finishedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	Agent       struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		Template *string `json:"template"`
	} `json:"agent"`
}

type TraceEventStatus string

const (
	TraceEventRunning TraceEventStatus = "running"
	TraceEventDone    TraceEventStatus = "done"
	TraceEventError   TraceEventStatus = "error"
)

// Mirrors models/events.py TraceEvent.
// const TraceEventSchema = z.object({
//     step: z.string(),
//     status: z.enum(["running", "done", "error"]),
//     iteration: z.number().nullable().optional(),
//     content: z.string().nullable().optional(),
//     args: z.record(z.string(), z.unknown()).nullable().optional(),
//     result_preview: z.string().nullable().optional(),
//     cost: z.object({
//         input_tokens: z.number(),
//         output_tokens: z.number(),
//         total_tokens: z.number(),
//         cost_usd: z.number(),
//         model: z.string(),
//     }).nullable().optional(),

type TraceEventCost struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	TotalTokens  int     `json:"total_tokens"`
	CostUsd      float64 `json:"cost_usd"`
	Model        string  `json:"model"`
}

// TraceEvent represents a single event inside the SSE stream or the `trace` JSON array
type TraceEvent struct {
	Step          string           `json:"step"`
	Status        TraceEventStatus `json:"status"`
	Iteration     *int             `json:"iteration,omitempty"`
	Content       *string          `json:"content,omitempty"`
	Args          map[string]any   `json:"args,omitempty"`
	ResultPreview *string          `json:"result_preview,omitempty"`
	Cost          *TraceEventCost  `json:"cost,omitempty"`
}

//  -- STRICT TYPING OF TOOLS USED GLOBALLY in cli--

type ToolName string

const (
	ToolWebSearch ToolName = "web_search"
	ToolScrapper  ToolName = "scrapper"
	ToolReadFeed  ToolName = "read_feed"
	ToolSendEmail ToolName = "send_email"
)

type ToolInfo struct {
	Name ToolName
	Desc string
}

var AllTools = []ToolInfo{
	{ToolWebSearch, "Search the web"},
	{ToolScrapper, "Read a web page as markdown"},
	{ToolReadFeed, "Latest items from an RSS/Atom feed"},
	{ToolSendEmail, "Email the result"},
}

// Valid reports whether t is a tool the runtime knows.
func (t ToolName) Valid() bool {
	for _, info := range AllTools {
		if info.Name == t {
			return true
		}
	}
	return false
}

// ValidateTools returns an error naming the first unknown tool, if any.
func ValidateTools(tools []ToolName) error {
	for _, t := range tools {
		if !t.Valid() {
			return fmt.Errorf("unknown tool %q", t)
		}
	}
	return nil
}

// payloads -- --

// CreateAgentRequest mirrors the backend AgentSchema. Name, Prompt, Tools and
// Status are required by the backend; Status must be ACTIVE for runs to work.
type CreateAgentRequest struct {
	Name        string      `json:"name"`
	Desc        string      `json:"desc,omitempty"`
	Prompt      any         `json:"prompt"`
	Template    string      `json:"template,omitempty"`
	Instruction any         `json:"instruction,omitempty"`
	Model       string      `json:"model,omitempty"`
	Tools       []ToolName  `json:"tools"`
	Schedule    any         `json:"schedule,omitempty"`
	Status      AgentStatus `json:"status"`
}

type UpdateAgentRequest struct {
	Name        *string          `json:"name,omitempty"`
	Desc        *string          `json:"desc,omitempty"`
	Prompt      any              `json:"prompt,omitempty"`
	Template    *string          `json:"template,omitempty"`
	Instruction any              `json:"instruction,omitempty"`
	Model       *string          `json:"model,omitempty"`
	Tools       []ToolName       `json:"tools,omitempty"`
	Schedule    *json.RawMessage `json:"schedule,omitempty"`
	Status      *AgentStatus     `json:"status,omitempty"`
}

var ClearSchedule = func() *json.RawMessage { r := json.RawMessage("null"); return &r }()

type TriggerRunResponse struct {
	Message string `json:"message"`
	RunID   string `json:"runId"`
}

// Report is the fixed shape of every final answer (agent-runtime
// models/report.py). It arrives as JSON in the "Final Answer" event's content.
type Report struct {
	Outcome    string          `json:"outcome"` // completed | partial | declined
	Headline   string          `json:"headline"`
	Summary    string          `json:"summary"`
	Findings   []ReportFinding `json:"findings"`
	Sources    []ReportSource  `json:"sources"`
	Takeaway   string          `json:"takeaway"`
	Confidence string          `json:"confidence"` // high | medium | low
}

type ReportFinding struct {
	Title  string   `json:"title"`
	Points []string `json:"points"`
}

type ReportSource struct {
	Name string  `json:"name"`
	Date *string `json:"date"`
	URL  *string `json:"url"`
}

// ParseReport decodes a final answer. ok is false for older runs that stored
// free text, so callers can fall back to plain rendering.
func ParseReport(content string) (*Report, bool) {
	var r Report
	if err := json.Unmarshal([]byte(content), &r); err != nil || r.Headline == "" || r.Summary == "" {
		return nil, false
	}
	return &r, true
}
