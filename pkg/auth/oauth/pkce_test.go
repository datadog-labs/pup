// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"regexp"
	"testing"
)

func TestGeneratePKCEChallenge(t *testing.T) {
	t.Parallel()
	challenge, err := GeneratePKCEChallenge()
	if err != nil {
		t.Fatalf("GeneratePKCEChallenge() error = %v", err)
	}

	if challenge == nil {
		t.Fatal("GeneratePKCEChallenge() returned nil")
	}

	// Verify verifier length (should be 128 characters, matching our implementation)
	if len(challenge.Verifier) != 128 {
		t.Errorf("Verifier length = %d, want 128", len(challenge.Verifier))
	}

	// Verify verifier is URL-safe base64 (no padding, only a-zA-Z0-9-_)
	matched, _ := regexp.MatchString(`^[A-Za-z0-9_-]+$`, challenge.Verifier)
	if !matched {
		t.Errorf("Verifier contains invalid characters: %s", challenge.Verifier)
	}

	// Verify challenge is derived from verifier using S256
	h := sha256.New()
	h.Write([]byte(challenge.Verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if challenge.Challenge != expectedChallenge {
		t.Errorf("Challenge = %s, want %s", challenge.Challenge, expectedChallenge)
	}

	// Verify method is S256
	if challenge.Method != "S256" {
		t.Errorf("Method = %s, want S256", challenge.Method)
	}

	// Verify challenge is URL-safe base64
	matched, _ = regexp.MatchString(`^[A-Za-z0-9_-]+$`, challenge.Challenge)
	if !matched {
		t.Errorf("Challenge contains invalid characters: %s", challenge.Challenge)
	}
}

func TestGeneratePKCEChallenge_Uniqueness(t *testing.T) {
	t.Parallel()
	// Generate multiple challenges and ensure they're unique
	challenges := make(map[string]bool)

	for i := 0; i < 10; i++ {
		challenge, err := GeneratePKCEChallenge()
		if err != nil {
			t.Fatalf("GeneratePKCEChallenge() error = %v", err)
		}

		if challenges[challenge.Verifier] {
			t.Errorf("Duplicate verifier generated: %s", challenge.Verifier)
		}
		challenges[challenge.Verifier] = true
	}

	if len(challenges) != 10 {
		t.Errorf("Expected 10 unique challenges, got %d", len(challenges))
	}
}

func TestGenerateState(t *testing.T) {
	t.Parallel()
	state, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() error = %v", err)
	}

	// Verify state length (should be 32 characters)
	if len(state) != 32 {
		t.Errorf("State length = %d, want 32", len(state))
	}

	// Verify state is URL-safe base64 (no padding, only a-zA-Z0-9-_)
	matched, _ := regexp.MatchString(`^[A-Za-z0-9_-]+$`, state)
	if !matched {
		t.Errorf("State contains invalid characters: %s", state)
	}
}

func TestGenerateState_Uniqueness(t *testing.T) {
	t.Parallel()
	// Generate multiple states and ensure they're unique
	states := make(map[string]bool)

	for i := 0; i < 10; i++ {
		state, err := GenerateState()
		if err != nil {
			t.Fatalf("GenerateState() error = %v", err)
		}

		if states[state] {
			t.Errorf("Duplicate state generated: %s", state)
		}
		states[state] = true
	}

	if len(states) != 10 {
		t.Errorf("Expected 10 unique states, got %d", len(states))
	}
}

func TestPKCEChallenge_RFC7636Compliance(t *testing.T) {
	t.Parallel()
	// Test that PKCE implementation complies with RFC 7636
	challenge, err := GeneratePKCEChallenge()
	if err != nil {
		t.Fatalf("GeneratePKCEChallenge() error = %v", err)
	}

	// RFC 7636 Section 4.1: code_verifier
	// code_verifier = high-entropy cryptographic random STRING using the
	// unreserved characters [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
	// with a minimum length of 43 characters and a maximum length of 128
	if len(challenge.Verifier) < 43 || len(challenge.Verifier) > 128 {
		t.Errorf("Verifier length %d not in RFC 7636 range [43, 128]", len(challenge.Verifier))
	}

	// RFC 7636 Section 4.2: code_challenge
	// code_challenge = BASE64URL(SHA256(ASCII(code_verifier)))
	if challenge.Method != "S256" {
		t.Errorf("Method = %s, RFC 7636 requires S256", challenge.Method)
	}

	// Verify the challenge is correctly computed
	h := sha256.New()
	h.Write([]byte(challenge.Verifier))
	hash := h.Sum(nil)
	expectedChallenge := base64.RawURLEncoding.EncodeToString(hash)

	if challenge.Challenge != expectedChallenge {
		t.Error("Challenge does not match SHA256(verifier) encoded as base64url")
	}
}

func TestGenerateRandomString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		length int
	}{
		{"minimum RFC length", 43},
		{"default length", 64},
		{"maximum RFC length", 128},
		{"small length", 16},
		{"medium length", 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			str, err := generateRandomString(tt.length)
			if err != nil {
				t.Fatalf("generateRandomString(%d) error = %v", tt.length, err)
			}

			if len(str) != tt.length {
				t.Errorf("String length = %d, want %d", len(str), tt.length)
			}

			// Verify it's URL-safe
			matched, _ := regexp.MatchString(`^[A-Za-z0-9_-]+$`, str)
			if !matched {
				t.Errorf("String contains invalid characters: %s", str)
			}
		})
	}
}
