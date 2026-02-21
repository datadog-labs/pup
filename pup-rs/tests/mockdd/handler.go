// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// RequestLog represents a single logged request in JSONL format.
type RequestLog struct {
	Timestamp string            `json:"timestamp"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Query     string            `json:"query,omitempty"`
	AuthType  string            `json:"auth_type"`
	HasBody   bool              `json:"has_body"`
	Headers   map[string]string `json:"headers,omitempty"`
}

// Handler serves mock Datadog API responses and logs requests.
type Handler struct {
	routes  []Route
	logFile *os.File
	mu      sync.Mutex
}

// NewHandler creates a new Handler with the given log file.
func NewHandler(logFile *os.File) *Handler {
	return &Handler{
		routes:  buildRoutes(),
		logFile: logFile,
	}
}

// ServeHTTP matches the request against known routes and returns fixture data.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Determine auth type
	authType := "none"
	if r.Header.Get("Authorization") != "" {
		authType = "bearer"
	} else if r.Header.Get("DD-API-KEY") != "" {
		authType = "api-key"
	}

	// Log the request
	entry := RequestLog{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Method:    r.Method,
		Path:      r.URL.Path,
		Query:     r.URL.RawQuery,
		AuthType:  authType,
		HasBody:   r.ContentLength > 0,
	}

	h.mu.Lock()
	if data, err := json.Marshal(entry); err == nil {
		fmt.Fprintf(h.logFile, "%s\n", data)
	}
	h.mu.Unlock()

	// Drain request body
	if r.Body != nil {
		io.Copy(io.Discard, r.Body)
		r.Body.Close()
	}

	// Match route
	for _, route := range h.routes {
		if route.Method == r.Method && route.Match(r.URL.Path) {
			w.Header().Set("Content-Type", "application/json")
			if route.Status != 0 {
				w.WriteHeader(route.Status)
			} else {
				w.WriteHeader(200)
			}
			w.Write(route.Fixture)
			return
		}
	}

	// 404 fallback - DD API often returns 200 even for "empty" results
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(`{"data":[]}`))
}
