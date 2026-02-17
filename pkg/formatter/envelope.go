// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package formatter

import "encoding/json"

// AgentEnvelope wraps API responses with metadata for agent consumption.
type AgentEnvelope struct {
	Status   string      `json:"status"`
	Data     interface{} `json:"data"`
	Metadata *Metadata   `json:"metadata,omitempty"`
}

// Metadata contains structured information about the response.
type Metadata struct {
	Count      *int     `json:"count,omitempty"`
	Truncated  bool     `json:"truncated,omitempty"`
	NextAction string   `json:"next_action,omitempty"`
	Command    string   `json:"command"`
	Warnings   []string `json:"warnings,omitempty"`
}

// AgentError is a structured error response for agent mode.
type AgentError struct {
	Status      string   `json:"status"`
	ErrorCode   int      `json:"error_code,omitempty"`
	Message     string   `json:"error_message"`
	Operation   string   `json:"operation"`
	Suggestions []string `json:"suggestions,omitempty"`
	APIResponse string   `json:"api_response,omitempty"`
}

// WrapForAgent wraps data in an AgentEnvelope and formats as JSON.
func WrapForAgent(data interface{}, meta *Metadata) (string, error) {
	env := AgentEnvelope{
		Status:   "success",
		Data:     data,
		Metadata: meta,
	}
	bytes, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FormatAgentError formats a structured error for agent mode.
func FormatAgentError(operation string, statusCode int, errMsg string, apiResponse string) (string, error) {
	ae := AgentError{
		Status:      "error",
		ErrorCode:   statusCode,
		Message:     errMsg,
		Operation:   operation,
		Suggestions: suggestionsForStatus(statusCode),
		APIResponse: apiResponse,
	}
	bytes, err := json.MarshalIndent(ae, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func suggestionsForStatus(code int) []string {
	switch {
	case code == 401:
		return []string{"Run 'pup auth login'", "Set DD_API_KEY and DD_APP_KEY"}
	case code == 403:
		return []string{"Verify your API/App keys have required permissions"}
	case code == 404:
		return []string{"Verify the resource ID", "Check if the resource was deleted"}
	case code == 429:
		return []string{"Wait and retry with backoff"}
	case code >= 500:
		return []string{"Retry after a short delay", "Check https://status.datadoghq.com/"}
	default:
		return nil
	}
}
