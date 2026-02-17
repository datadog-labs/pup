// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package formatter

import (
	"encoding/json"
	"testing"
)

func TestWrapForAgent(t *testing.T) {
	data := map[string]string{"key": "value"}
	count := 1
	meta := &Metadata{
		Count:   &count,
		Command: "test command",
	}

	result, err := WrapForAgent(data, meta)
	if err != nil {
		t.Fatalf("WrapForAgent() error = %v", err)
	}

	var envelope AgentEnvelope
	if err := json.Unmarshal([]byte(result), &envelope); err != nil {
		t.Fatalf("WrapForAgent() produced invalid JSON: %v", err)
	}

	if envelope.Status != "success" {
		t.Errorf("envelope.Status = %q, want %q", envelope.Status, "success")
	}
	if envelope.Metadata == nil {
		t.Fatal("envelope.Metadata should not be nil")
	}
	if *envelope.Metadata.Count != 1 {
		t.Errorf("envelope.Metadata.Count = %d, want 1", *envelope.Metadata.Count)
	}
	if envelope.Metadata.Command != "test command" {
		t.Errorf("envelope.Metadata.Command = %q, want %q", envelope.Metadata.Command, "test command")
	}
}

func TestWrapForAgent_NilMetadata(t *testing.T) {
	data := []int{1, 2, 3}

	result, err := WrapForAgent(data, nil)
	if err != nil {
		t.Fatalf("WrapForAgent() error = %v", err)
	}

	var envelope AgentEnvelope
	if err := json.Unmarshal([]byte(result), &envelope); err != nil {
		t.Fatalf("WrapForAgent() produced invalid JSON: %v", err)
	}

	if envelope.Status != "success" {
		t.Errorf("envelope.Status = %q, want %q", envelope.Status, "success")
	}
	if envelope.Metadata != nil {
		t.Error("envelope.Metadata should be nil when no metadata provided")
	}
}

func TestWrapForAgent_Truncated(t *testing.T) {
	data := "some data"
	count := 100
	meta := &Metadata{
		Count:      &count,
		Truncated:  true,
		NextAction: "Use --limit=500",
		Command:    "monitors list",
	}

	result, err := WrapForAgent(data, meta)
	if err != nil {
		t.Fatalf("WrapForAgent() error = %v", err)
	}

	var envelope AgentEnvelope
	if err := json.Unmarshal([]byte(result), &envelope); err != nil {
		t.Fatalf("WrapForAgent() produced invalid JSON: %v", err)
	}

	if !envelope.Metadata.Truncated {
		t.Error("envelope.Metadata.Truncated should be true")
	}
	if envelope.Metadata.NextAction != "Use --limit=500" {
		t.Errorf("envelope.Metadata.NextAction = %q, want %q", envelope.Metadata.NextAction, "Use --limit=500")
	}
}

func TestFormatAgentError(t *testing.T) {
	result, err := FormatAgentError("list monitors", 401, "authentication failed", "unauthorized")
	if err != nil {
		t.Fatalf("FormatAgentError() error = %v", err)
	}

	var agentErr AgentError
	if err := json.Unmarshal([]byte(result), &agentErr); err != nil {
		t.Fatalf("FormatAgentError() produced invalid JSON: %v", err)
	}

	if agentErr.Status != "error" {
		t.Errorf("agentErr.Status = %q, want %q", agentErr.Status, "error")
	}
	if agentErr.ErrorCode != 401 {
		t.Errorf("agentErr.ErrorCode = %d, want 401", agentErr.ErrorCode)
	}
	if agentErr.Operation != "list monitors" {
		t.Errorf("agentErr.Operation = %q, want %q", agentErr.Operation, "list monitors")
	}
	if len(agentErr.Suggestions) == 0 {
		t.Error("401 error should have suggestions")
	}
}

func TestFormatAgentError_AllStatusCodes(t *testing.T) {
	tests := []struct {
		code            int
		wantSuggestions bool
	}{
		{401, true},
		{403, true},
		{404, true},
		{429, true},
		{500, true},
		{200, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result, err := FormatAgentError("test", tt.code, "error", "")
			if err != nil {
				t.Fatalf("FormatAgentError() error = %v", err)
			}

			var agentErr AgentError
			if err := json.Unmarshal([]byte(result), &agentErr); err != nil {
				t.Fatalf("FormatAgentError() produced invalid JSON: %v", err)
			}

			hasSuggestions := len(agentErr.Suggestions) > 0
			if hasSuggestions != tt.wantSuggestions {
				t.Errorf("status %d: hasSuggestions = %v, want %v", tt.code, hasSuggestions, tt.wantSuggestions)
			}
		})
	}
}
