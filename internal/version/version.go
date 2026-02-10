// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package version

// Version information set by build flags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// BuildInfo returns formatted build information
func BuildInfo() string {
	return "Pup " + Version + " (" + GitCommit + ") built on " + BuildDate
}
