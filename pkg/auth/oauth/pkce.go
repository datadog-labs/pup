// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// PKCEChallenge represents a PKCE code challenge
type PKCEChallenge struct {
	Verifier  string
	Challenge string
	Method    string
}

// GeneratePKCEChallenge generates a PKCE code verifier and challenge
// Uses S256 method as specified in RFC 7636
func GeneratePKCEChallenge() (*PKCEChallenge, error) {
	// Generate code verifier (43-128 characters)
	verifier, err := generateRandomString(128)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	// Generate code challenge using S256 (SHA256)
	h := sha256.New()
	h.Write([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return &PKCEChallenge{
		Verifier:  verifier,
		Challenge: challenge,
		Method:    "S256",
	}, nil
}

// generateRandomString generates a cryptographically secure random string
func generateRandomString(length int) (string, error) {
	// Calculate number of bytes needed
	numBytes := (length * 3) / 4

	// Generate random bytes
	bytes := make([]byte, numBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to base64 URL encoding without padding
	encoded := base64.RawURLEncoding.EncodeToString(bytes)

	// Trim to requested length
	if len(encoded) > length {
		encoded = encoded[:length]
	}

	return encoded, nil
}

// GenerateState generates a random state parameter for CSRF protection
func GenerateState() (string, error) {
	return generateRandomString(32)
}
