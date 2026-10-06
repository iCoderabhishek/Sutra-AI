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
	var agents []Agent

	err := c.doRequest("GET", "/api/v1/agents", nil, &agents)
	if err != nil {
		return nil, err
	}
	return agents, nil
}

// CreateAgent creates a new agent and returns the created Agent object
func (c *Client) CreateAgent(payload CreateAgentRequest) (*Agent, error) {
	var agent Agent

	// Here we pass the 'payload' struct. The doRequest helper will automatically
	// json.Marshal it and send it as the request body!
	err := c.doRequest("POST", "/api/v1/agents", payload, &agent)
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (c *Client) UpdateAgent(id string, payload CreateAgentRequest) (*Agent, error) {
	var agent Agent

	path := fmt.Sprintf("/api/v1/agents/:%s", id)

	err := c.doRequest("PATCH", path, payload, &agent)
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (c *Client) DeleteAgent(id string) error {
	path := fmt.Sprintf("/api/v1/agents/%s", id)

	return c.doRequest("DELETE", path, nil, nil)
}
