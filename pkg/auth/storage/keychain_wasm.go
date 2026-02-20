// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build js

package storage

// IsKeychainAvailable always returns false on WASM — the OS keychain is not
// accessible from a browser/WASI runtime.
func IsKeychainAvailable() bool {
	return false
}
