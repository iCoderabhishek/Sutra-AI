package api

import "fmt"

// GetAgent fetches a single agent by its ID
func (c *Client) GetAgent(id string) (*Agent, error) {
	var agent Agent

	// Format the URL path with the ID
	path := fmt.Sprintf("/api/v1/agents/%s", id)

	// doRequest(method, path, requestBody, responseBody)
	// We pass 'nil' for requestBody because GET requests don't have a body.
	// We pass '&agent' for responseBody so it automatically unmarshals the JSON into the struct.
	err := c.doRequest("GET", path, nil, &agent)
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (c *Client) ListAgents() ([]Agent, error) {
	var response struct {
		Data []Agent `json:"data"`
	}

	err := c.doRequest("GET", "/api/v1/agents?limit=100", nil, &response)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// CreateAgent creates a new agent and returns the created Agent object
func (c *Client) CreateAgent(payload CreateAgentRequest) (*Agent, error) {
	if err := ValidateTools(payload.Tools); err != nil {
		return nil, err
	}
	if payload.Tools == nil {
		payload.Tools = []ToolName{}
	}
	var agent Agent

	// Here we pass the 'payload' struct. The doRequest helper will automatically
	// json.Marshal it and send it as the request body!
	err := c.doRequest("POST", "/api/v1/agents", payload, &agent)
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

// UpdateAgent sends only the fields set on payload (PATCH semantics).
func (c *Client) UpdateAgent(id string, payload UpdateAgentRequest) (*Agent, error) {
	if err := ValidateTools(payload.Tools); err != nil {
		return nil, err
	}
	var agent Agent

	path := fmt.Sprintf("/api/v1/agents/%s", id)

	err := c.doRequest("PATCH", path, payload, &agent)
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

// SetAgentStatus pauses, resumes or deactivates an agent without touching other fields.
func (c *Client) SetAgentStatus(id string, status AgentStatus) (*Agent, error) {
	return c.UpdateAgent(id, UpdateAgentRequest{Status: &status})
}

func (c *Client) DeleteAgent(id string) error {
	path := fmt.Sprintf("/api/v1/agents/%s", id)

	return c.doRequest("DELETE", path, nil, nil)
}
