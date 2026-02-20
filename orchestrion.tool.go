// This file was created by `orchestrion pin`, and is used to ensure the
// `go.mod` file contains the necessary entries to ensure repeatable builds when
// using `orchestrion`. It is also used to set up which integrations are enabled.

//go:build tools

//go:generate go run github.com/DataDog/orchestrion pin -generate

package tools

// Imports in this file determine which tracer integrations are enabled in
// orchestrion. New integrations can be automatically discovered by running
// `orchestrion pin` again. You can also manually add new imports here to
// enable additional integrations. When doing so, you can run `orchestrion pin`
// to make sure manually added integrations are valid (i.e, the imported package
// includes a valid `orchestrion.yml` file).
import (
	// Ensures `orchestrion` is present in `go.mod` so that builds are repeatable.
	// Do not remove.
	_ "github.com/DataDog/orchestrion" // integration

	// Only include integrations pup actually uses (net/http for API calls and OAuth flows).
	// Do not run `orchestrion pin` without reviewing - it will re-add all integrations.
	_ "github.com/DataDog/dd-trace-go/contrib/net/http/v2" // integration
)
