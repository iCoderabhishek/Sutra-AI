package api

import (
	"encoding/json"
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
	Tools       []string        `json:"tools"`
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
	TotalCost   string          `json:"totalCost"`
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

// payloads -- --

type CreateAgentRequest struct {
	Name        string   `json:"name"`
	Prompt      any      `json:"prompt"`
	Instruction any      `json:"instruction,omitempty"`
	Tools       []string `json:"tools"`
	Schedule    any      `json:"schedule,omitempty"`
}
