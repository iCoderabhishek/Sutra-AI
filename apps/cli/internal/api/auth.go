package api

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"sutra-ai-agent/apps/cli/internal/config"
)


func (c *Client) InitAuth() error {
	cookie, err := config.LoadSession()
	if err != nil {
		return err
	}
	c.sessionCookie = cookie
	return nil
}

// Login opens the browser for OAuth and starts a local server to receive the callback.
func (c *Client) Login() error {

	cookieChan := make(chan string)
	errChan := make(chan error)

	env := config.GetEnvConfig()

	// local http server to listen callback
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    env.LocalServerAddr,
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// The backend redirects with ?token= for the CLI flow
		token := r.URL.Query().Get("token")
		
		if token == "" {
			http.Error(w, "No auth token received", http.StatusBadRequest)
			errChan <- fmt.Errorf("no token received in callback")
			return
		}

		fmt.Fprintf(w, "<html><body><h2>Authentication successful!</h2><p>You can close this window and return to the CLI.</p></body></html>")
		cookieChan <- token
	})
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to start local server: %w", err)
		}
	}()
	// open browser for oauth

	loginURL := fmt.Sprintf("%s%s?cli_redirect=%s", c.baseURL, env.OAuthRoute, env.OAuthCallback)
	fmt.Printf("Opening browser to authenticate: %s\n", loginURL)
	if err := openBrowser(loginURL); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	var sessionCookie string
	select {
	case sessionCookie = <-cookieChan:
		fmt.Println("Successfully logged in!")
	case err := <-errChan:
		return err
	case <-time.After(3 * time.Minute):
		return fmt.Errorf("authentication timed out")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Warning: failed to shutdown local server: %v\n", err)
	}

	if err := config.SaveSession(sessionCookie); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	c.sessionCookie = sessionCookie

	return nil
}






func openBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}