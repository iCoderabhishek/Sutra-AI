package api

func (c *Client) GetCredits() (*Credits, error) {
	var credits Credits
	err := c.doRequest("GET", "/api/v1/credits", nil, &credits)
	if err != nil {
		return nil, err
	}
	return &credits, nil
}

func (c *Client) GetDashboard() (*DashboardStats, error) {
	var dashboard DashboardStats
	err := c.doRequest("GET", "/api/v1/dashboard", nil, &dashboard)
	if err != nil {
		return nil, err
	}
	return &dashboard, nil
}
