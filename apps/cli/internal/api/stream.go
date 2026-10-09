package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// StreamRunLogs connects to the sse endpoint for a run 
// and yields TraceEvent objects as they arrive from the backend.
func (c *Client) StreamRunLogs(ctx context.Context, runId string) (<-chan TraceEvent, <-chan error) {
	eventChan := make(chan TraceEvent)
	errChan := make(chan error, 1)

	go func() {
		defer close(eventChan)
		defer close(errChan)

		url := fmt.Sprintf("%s/api/v1/runs/%s/stream", c.baseURL, runId)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			errChan <- fmt.Errorf("failed to create stream request: %w", err)
			return
		}

		if c.sessionCookie != "" {
			req.Header.Set("Authorization", "Bearer "+c.sessionCookie)
		}
		
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")

		resp, err := c.http.Do(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to connect to stream: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			errChan <- fmt.Errorf("stream endpoint returned error status: %d", resp.StatusCode)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			// Extract the JSON payload
			data := strings.TrimPrefix(line, "data: ")
			data = strings.TrimSpace(data)

			var event TraceEvent
			// Skip malformed events silently: printing here would corrupt the TUI.
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			select {
			case eventChan <- event:
			case <-ctx.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil && err != context.Canceled {
			errChan <- fmt.Errorf("error reading stream: %w", err)
		}
	}()

	return eventChan, errChan
}
