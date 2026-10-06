package api

import (
	"fmt"
	"net/url"
)

func (c *Client) TriggerRun(agentId string) (*JobRun, error) {
	var jobRun JobRun
	err := c.doRequest("POST", fmt.Sprintf("/api/v1/agents/%s/run", agentId), nil, &jobRun)
	if err != nil {
		return nil, err
	}
	return &jobRun, nil
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

