// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package callback

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"time"
)

// CallbackResult represents the result of the OAuth callback
type CallbackResult struct {
	Code             string
	State            string
	Error            string
	ErrorDescription string
}

// Server handles the OAuth callback
type Server struct {
	port     int
	server   *http.Server
	resultCh chan CallbackResult
}

// DCRRedirectPorts are the specific ports to try for OAuth callbacks
// Must match TypeScript PR #84 and DCR registration
var DCRRedirectPorts = []int{8000, 8080, 8888, 9000}

// NewServer creates a new callback server
// Tries to bind to one of the pre-registered DCR ports (8000, 8080, 8888, 9000)
// Matches TypeScript PR #84 behavior for compatibility
func NewServer() (*Server, error) {
	// Try each DCR port in order
	var listener net.Listener
	var port int
	var lastErr error

	for _, p := range DCRRedirectPorts {
		addr := fmt.Sprintf("127.0.0.1:%d", p)
		l, err := net.Listen("tcp", addr)
		if err == nil {
			listener = l
			port = p
			break
		}
		lastErr = err
	}

	if listener == nil {
		return nil, fmt.Errorf("failed to bind to any DCR port (tried %v): %w", DCRRedirectPorts, lastErr)
	}
	if err := listener.Close(); err != nil {
		return nil, fmt.Errorf("failed to close test listener: %w", err)
	}

	return &Server{
		port:     port,
		resultCh: make(chan CallbackResult, 1),
	}, nil
}

// Port returns the server port
func (s *Server) Port() int {
	return s.port
}

// RedirectURI returns the full redirect URI
// Uses /oauth/callback path to match TypeScript PR #84
func (s *Server) RedirectURI() string {
	return fmt.Sprintf("http://127.0.0.1:%d/oauth/callback", s.port)
}

// Start starts the callback server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/callback", s.handleCallback)

	s.server = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler: mux,
	}

	// Start server in background
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Callback server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the callback server
func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// WaitForCallback waits for the OAuth callback
func (s *Server) WaitForCallback(timeout time.Duration) (*CallbackResult, error) {
	select {
	case result := <-s.resultCh:
		return &result, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for OAuth callback")
	}
}

// handleCallback handles the OAuth callback request
func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	code := query.Get("code")
	state := query.Get("state")
	errorCode := query.Get("error")
	errorDesc := query.Get("error_description")

	// Send result to channel
	s.resultCh <- CallbackResult{
		Code:             code,
		State:            state,
		Error:            errorCode,
		ErrorDescription: errorDesc,
	}

	// Render success or error page
	if errorCode != "" {
		s.renderError(w, errorCode, errorDesc)
	} else {
		s.renderSuccess(w)
	}
}

// renderSuccess renders the success page
func (s *Server) renderSuccess(w http.ResponseWriter) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Authentication Successful</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .container {
            background: white;
            padding: 3rem;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            text-align: center;
            max-width: 400px;
        }
        .success-icon {
            font-size: 4rem;
            color: #48bb78;
            margin-bottom: 1rem;
        }
        h1 {
            color: #2d3748;
            margin-bottom: 1rem;
        }
        p {
            color: #718096;
            line-height: 1.6;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">✓</div>
        <h1>Authentication Successful!</h1>
        <p>You have successfully authenticated with Datadog.</p>
        <p><strong>You can close this window and return to your terminal.</strong></p>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := template.Must(template.New("success").Parse(tmpl)).Execute(w, nil); err != nil {
		// Log error but don't fail - response is already started
		fmt.Fprintf(os.Stderr, "Failed to render success template: %v\n", err)
	}
}

// renderError renders the error page
func (s *Server) renderError(w http.ResponseWriter, errorCode, errorDesc string) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Authentication Failed</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
        }
        .container {
            background: white;
            padding: 3rem;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            text-align: center;
            max-width: 400px;
        }
        .error-icon {
            font-size: 4rem;
            color: #f56565;
            margin-bottom: 1rem;
        }
        h1 {
            color: #2d3748;
            margin-bottom: 1rem;
        }
        p {
            color: #718096;
            line-height: 1.6;
        }
        .error-details {
            background: #fed7d7;
            border: 1px solid #fc8181;
            border-radius: 6px;
            padding: 1rem;
            margin-top: 1rem;
            font-size: 0.875rem;
            color: #742a2a;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="error-icon">✗</div>
        <h1>Authentication Failed</h1>
        <p>There was an error during authentication.</p>
        {{if .ErrorDesc}}
        <div class="error-details">
            <strong>{{.ErrorCode}}:</strong> {{.ErrorDesc}}
        </div>
        {{else}}
        <div class="error-details">
            <strong>Error:</strong> {{.ErrorCode}}
        </div>
        {{end}}
        <p style="margin-top: 1.5rem;"><strong>You can close this window and try again.</strong></p>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)

	data := struct {
		ErrorCode string
		ErrorDesc string
	}{
		ErrorCode: errorCode,
		ErrorDesc: errorDesc,
	}

	if err := template.Must(template.New("error").Parse(tmpl)).Execute(w, data); err != nil {
		// Log error but don't fail - response is already started
		fmt.Fprintf(os.Stderr, "Failed to render error template: %v\n", err)
	}
}
