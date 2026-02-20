// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build js

package storage

import "fmt"

// GetStorage returns an error on WASM directing users to DD_ACCESS_TOKEN.
// Local token storage is not available in WASM runtimes.
func GetStorage(_ *StorageOptions) (Storage, error) {
	return nil, fmt.Errorf(
		"token storage is not available in WASM builds; set DD_ACCESS_TOKEN instead",
	)
}

// GetActiveBackend returns an empty backend type on WASM.
func GetActiveBackend() BackendType {
	return ""
}

// IsUsingSecureStorage always returns false on WASM.
func IsUsingSecureStorage() bool {
	return false
}

// GetStorageDescription returns a description indicating WASM mode.
func GetStorageDescription() string {
	return "not available (WASM)"
}

// ResetStorage is a no-op on WASM.
func ResetStorage() {}
