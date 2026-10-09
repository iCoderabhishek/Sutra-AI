package api

import (
	"fmt"
	"net/url"
)

// TriggerRun queues a run. The backend only returns the run ID (202), so the											
// returned JobRun has ID, AgentID and Status set; call GetRun for full details.
func (c *Client) TriggerRun(agentId string) (*JobRun, error) {
	var res TriggerRunResponse
	err := c.doRequest("POST", fmt.Sprintf("/api/v1/agents/%s/run", agentId), nil, &res)
	if err != nil {
		return nil, err
	}
	if res.RunID == "" {
		return nil, fmt.Errorf("trigger run: backend returned no runId")
	}
	return &JobRun{ID: res.RunID, AgentID: agentId, Status: RunQueued}, nil
}

func (c *Client) ListRuns(agentId string, status string) ([]RecentRun, error) {
	query := url.Values{}
	if agentId != "" {
		query.Add("agentId", agentId)
	}
	if status != "" {
		query.Add("status", status)
	}

	path := "/api/v1/runs"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var response struct {
		Data []RecentRun `json:"data"`
	}

	err := c.doRequest("GET", path, nil, &response)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (c *Client) GetRun(id string) (*JobRun, error) {
	var jobRun JobRun
	err := c.doRequest("GET", fmt.Sprintf("/api/v1/runs/%s", id), nil, &jobRun)
	if err != nil {
		return nil, err
	}
	return &jobRun, nil
}

