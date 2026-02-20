// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build js

package client

import (
	"context"

	"github.com/datadog-labs/pup/pkg/config"
)

// tryOAuthFromStorage is a no-op on WASM — local token storage is not
// available. Use DD_ACCESS_TOKEN instead.
func tryOAuthFromStorage(_ *config.Config) context.Context {
	return nil
}
