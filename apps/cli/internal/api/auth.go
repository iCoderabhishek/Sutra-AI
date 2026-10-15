package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"

	"github.com/iCoderabhishek/Sutra-AI/apps/cli/internal/config"
)

func (c *Client) InitAuth() error {
	cookie, err := config.LoadSession()
	if err != nil {
		return err
	}
	c.sessionCookie = cookie
	return nil
}

// Logout forgets the session locally. The backend token simply expires.
func (c *Client) Logout() error {
	c.sessionCookie = ""
	return config.ClearSession()
}

// Login opens the browser for OAuth and starts a local server to receive the callback.
func (c *Client) Login() error {

	// Buffered, so a late browser callback can't block a handler forever.
	cookieChan := make(chan string, 1)
	errChan := make(chan error, 1)

	env := config.GetEnvConfig()

	callbackURL, err := url.Parse(env.OAuthCallback)
	if err != nil {
		return fmt.Errorf("invalid OAuth callback URL: %w", err)
	}

	// Bind before opening the browser. If the usual port is taken (another app,
	// or an older sutra still running), use a free port instead; the backend
	// accepts any localhost callback.
	ln, err := net.Listen("tcp", env.LocalServerAddr)
	if err != nil {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return fmt.Errorf("failed to start local login server: %w", err)
		}
		callbackURL.Host = fmt.Sprintf("localhost:%d", ln.Addr().(*net.TCPAddr).Port)
	}

	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}
	// Always release the port, including on timeout or error, so the next
	// login attempt in the same session can bind again.
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	// A random state travels through the backend and back. Without it, any web
	// page could hit the local callback and log the CLI into another account.
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return fmt.Errorf("failed to create login state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)
	q := callbackURL.Query()
	q.Set("state", state)
	callbackURL.RawQuery = q.Encode()

	mux.HandleFunc(callbackURL.Path, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "Login link expired or invalid. Start the login again from the CLI.", http.StatusBadRequest)
			return
		}
		// The backend redirects with ?token= for the CLI flow
		token := r.URL.Query().Get("token")

		if token == "" {
			http.Error(w, "No auth token received", http.StatusBadRequest)
			select {
			case errChan <- fmt.Errorf("no token received in callback"):
			default:
			}
			return
		}

		fmt.Fprintf(w, "<html><body><h2>Authentication successful!</h2><p>You can close this window and return to the CLI.</p></body></html>")
		select {
		case cookieChan <- token:
		default:
		}
	})
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			select {
			case errChan <- fmt.Errorf("local login server stopped: %w", err):
			default:
			}
		}
	}()
	// open browser for oauth

	loginURL := fmt.Sprintf("%s%s?cli_redirect=%s", c.baseURL, env.OAuthRoute, url.QueryEscape(callbackURL.String()))
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
