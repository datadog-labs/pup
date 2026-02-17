// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var investigationsCmd = &cobra.Command{
	Use:   "investigations",
	Short: "Manage Bits AI investigations",
	Long: `Manage Bits AI investigations.

Bits AI investigations allow you to trigger automated root cause analysis
for monitor alerts.

CAPABILITIES:
  • Trigger a new investigation (monitor alert)
  • Get investigation details by ID
  • List investigations with optional filters

EXAMPLES:
  # Trigger investigation from a monitor alert
  pup investigations trigger --type=monitor_alert --monitor-id=123456 --event-id="evt-abc" --event-ts=1706918956000

  # Get investigation details
  pup investigations get <investigation-id>

  # List investigations
  pup investigations list --page-limit=20

AUTHENTICATION:
  Requires OAuth2 (via 'pup auth login') or a valid API key + Application key.`,
}

var investigationsTriggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Trigger a new investigation",
	RunE:  runInvestigationsTrigger,
}

var investigationsGetCmd = &cobra.Command{
	Use:   "get [investigation-id]",
	Short: "Get investigation details",
	Args:  cobra.ExactArgs(1),
	RunE:  runInvestigationsGet,
}

var investigationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List investigations",
	RunE:  runInvestigationsList,
}

var (
	invTriggerType string
	invMonitorID   int64
	invEventID     string
	invEventTS     int64
	invPageOffset  int64
	invPageLimit   int64
	invFilterMonID int64
)

func init() {
	// trigger flags
	investigationsTriggerCmd.Flags().StringVar(&invTriggerType, "type", "", "Investigation type: monitor_alert (required)")
	investigationsTriggerCmd.Flags().Int64Var(&invMonitorID, "monitor-id", 0, "Monitor ID (required for monitor_alert)")
	investigationsTriggerCmd.Flags().StringVar(&invEventID, "event-id", "", "Event ID (required for monitor_alert)")
	investigationsTriggerCmd.Flags().Int64Var(&invEventTS, "event-ts", 0, "Event timestamp in milliseconds (required for monitor_alert)")
	if err := investigationsTriggerCmd.MarkFlagRequired("type"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// list flags
	investigationsListCmd.Flags().Int64Var(&invPageOffset, "page-offset", 0, "Pagination offset")
	investigationsListCmd.Flags().Int64Var(&invPageLimit, "page-limit", 10, "Page size")
	investigationsListCmd.Flags().Int64Var(&invFilterMonID, "monitor-id", 0, "Filter by monitor ID")

	investigationsCmd.AddCommand(investigationsTriggerCmd, investigationsGetCmd, investigationsListCmd)
}

func runInvestigationsTrigger(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	body, err := buildTriggerRequestBody()
	if err != nil {
		return err
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	resp, err := client.RawRequest("POST", "/api/v2/bits-ai/investigations", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to trigger investigation: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	result, err := readRawResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to trigger investigation: %w", err)
	}

	return formatAndPrint(result, nil)
}

func runInvestigationsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id := args[0]
	resp, err := client.RawRequest("GET", "/api/v2/bits-ai/investigations/"+id, nil)
	if err != nil {
		return fmt.Errorf("failed to get investigation: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	result, err := readRawResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to get investigation: %w", err)
	}

	return formatAndPrint(result, nil)
}

func runInvestigationsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/api/v2/bits-ai/investigations?page[offset]=%d&page[limit]=%d", invPageOffset, invPageLimit)
	if invFilterMonID != 0 {
		path += fmt.Sprintf("&filter[monitor_id]=%d", invFilterMonID)
	}

	resp, err := client.RawRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to list investigations: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	result, err := readRawResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to list investigations: %w", err)
	}

	return formatAndPrint(result, nil)
}

func buildTriggerRequestBody() (map[string]any, error) {
	var trigger map[string]any

	if invTriggerType != "monitor_alert" {
		return nil, fmt.Errorf("invalid investigation type %q: must be monitor_alert", invTriggerType)
	}
	if invMonitorID == 0 {
		return nil, fmt.Errorf("--monitor-id is required for monitor_alert investigations")
	}
	if invEventID == "" {
		return nil, fmt.Errorf("--event-id is required for monitor_alert investigations")
	}
	if invEventTS == 0 {
		return nil, fmt.Errorf("--event-ts is required for monitor_alert investigations")
	}
	trigger = map[string]any{
		"type": "monitor_alert_trigger",
		"monitor_alert_trigger": map[string]any{
			"monitor_id": invMonitorID,
			"event_id":   invEventID,
			"event_ts":   invEventTS,
		},
	}

	return map[string]any{
		"data": map[string]any{
			"type": "trigger_investigation_request",
			"attributes": map[string]any{
				"trigger": trigger,
			},
		},
	}, nil
}

func readRawResponse(resp *http.Response) (map[string]any, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		msg := string(bodyBytes)
		switch {
		case resp.StatusCode >= 500:
			return nil, fmt.Errorf("server error (status %d): %s\n\nThe Datadog API is experiencing issues. Please try again later.", resp.StatusCode, msg)
		case resp.StatusCode == 429:
			return nil, fmt.Errorf("rate limited (status 429): %s\n\nPlease wait a moment and try again.", msg)
		case resp.StatusCode == 403:
			return nil, fmt.Errorf("access denied (status 403): %s\n\nVerify your API/App keys have the required permissions.", msg)
		case resp.StatusCode == 401:
			return nil, fmt.Errorf("authentication failed (status 401): %s\n\nRun 'pup auth login' or verify your DD_API_KEY and DD_APP_KEY.", msg)
		case resp.StatusCode == 404:
			return nil, fmt.Errorf("not found (status 404): %s", msg)
		default:
			return nil, fmt.Errorf("request failed (status %d): %s", resp.StatusCode, msg)
		}
	}

	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("parsing response JSON: %w", err)
	}

	return result, nil
}
