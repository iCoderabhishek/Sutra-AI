package api

type LLMModel struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ContextLength int    `json:"context_length"`
}

func (c *Client) GetLLMModels() ([]LLMModel, error) {
	var models []LLMModel
	err := c.doRequest("GET", "/api/v1/models", nil, &models)
	return models, err
}
