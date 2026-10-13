package tui

import (
	"encoding/json"
	"testing"

	"github.com/iCoderabhishek/Sutra-AI/internal/api"
)

// A client pointing nowhere: GetLLMModels fails and the form falls back to "auto".
func offlineClient() *api.Client { return api.NewClient("http://127.0.0.1:1", "") }

func strPtr(s string) *string { return &s }

func patchBody(t *testing.T, f *agentForm) map[string]any {
	t.Helper()
	req, problem, _ := f.validate()
	if problem != "" {
		t.Fatalf("validate: %s", problem)
	}
	raw, err := json.Marshal(updateFrom(req))
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	return body
}

func TestEditFormPrefillsAndKeepsAgent(t *testing.T) {
	a := api.Agent{
		ID:       "agent-1",
		Name:     "Bike updates",
		Desc:     strPtr("New bikes in India"),
		Prompt:   json.RawMessage(`{"goal":"Find new bikes","telegram_chat_id":"123456789","extra":"keep me"}`),
		Model:    strPtr("some/unlisted-model:free"),
		Tools:    []api.ToolName{api.ToolWebSearch, api.ToolNotify},
		Schedule: json.RawMessage(`{"cron":"0 9 * * *"}`),
		Status:   api.StatusPaused,
	}
	f := newEditAgentForm(offlineClient(), a)

	if f.name.Value() != "Bike updates" || f.goal.Value() != "Find new bikes" ||
		f.cron.Value() != "0 9 * * *" || f.telegram.Value() != "123456789" || f.active {
		t.Fatalf("form not pre-filled: name=%q goal=%q cron=%q chat=%q active=%v",
			f.name.Value(), f.goal.Value(), f.cron.Value(), f.telegram.Value(), f.active)
	}

	body := patchBody(t, f)
	if body["model"] != "some/unlisted-model:free" {
		t.Errorf("model changed to %v", body["model"])
	}
	if body["status"] != "PAUSED" {
		t.Errorf("status = %v", body["status"])
	}
	prompt := body["prompt"].(map[string]any)
	if prompt["goal"] != "Find new bikes" || prompt["telegram_chat_id"] != "123456789" || prompt["extra"] != "keep me" {
		t.Errorf("prompt = %v", prompt)
	}
	if body["schedule"].(map[string]any)["cron"] != "0 9 * * *" {
		t.Errorf("schedule = %v", body["schedule"])
	}
}

func TestEditFormClearsWhatTheUserRemoved(t *testing.T) {
	a := api.Agent{
		ID:       "agent-1",
		Name:     "Bike updates",
		Prompt:   json.RawMessage(`{"goal":"Find new bikes","telegram_chat_id":"123456789"}`),
		Tools:    []api.ToolName{api.ToolNotify},
		Schedule: json.RawMessage(`{"cron":"0 9 * * *"}`),
		Status:   api.StatusActive,
	}
	f := newEditAgentForm(offlineClient(), a)
	f.tools = map[api.ToolName]bool{} // untick every tool
	f.cron.SetValue("")               // manual only

	body := patchBody(t, f)
	if tools, ok := body["tools"].([]any); !ok || len(tools) != 0 {
		t.Errorf("tools should be sent as an empty list, got %#v", body["tools"])
	}
	if v, ok := body["schedule"]; !ok || v != nil {
		t.Errorf("schedule should be sent as null, got %#v (present=%v)", v, ok)
	}
	if _, ok := body["prompt"].(map[string]any)["telegram_chat_id"]; ok {
		t.Error("chat ID should be dropped when notify is off")
	}
}
