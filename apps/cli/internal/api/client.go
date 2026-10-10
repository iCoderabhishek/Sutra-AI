package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Client struct {
	baseURL       string
	sessionCookie string
	http          *http.Client
}

func NewClient(baseURL string, sessionCookie string) *Client {
	return &Client{
		baseURL:       baseURL,
		sessionCookie: sessionCookie,
		http:          &http.Client{},
	}
}

func (c *Client) doRequest(method string, path string, reqBody interface{}, resBody interface{}) error {
	var bodyReader io.Reader
	if reqBody != nil {
		jsonBytes, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if c.sessionCookie != "" {
		req.Header.Set("Authorization", "Bearer "+c.sessionCookie)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Method:     resp.Request.Method,
			URL:        resp.Request.URL.String(),
		}

		// Attempt to parse the JSON error message from the backend
		bodyBytes, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(bodyBytes, apiErr); err != nil {
			// If not JSON, use the raw response as the error message
			apiErr.Message = string(bodyBytes)
		}

		if apiErr.Message == "" {
			apiErr.Message = "Unknown error"
		}

		return apiErr
	}

	if resBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(resBody); err != nil {
			return err
		}
	}

	return nil
}
