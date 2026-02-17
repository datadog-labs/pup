// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/spf13/cobra"
)

var productAnalyticsCmd = &cobra.Command{
	Use:   "product-analytics",
	Short: "Send product analytics events",
	Long: `Send server-side product analytics events to Datadog.

Product Analytics provides insights into user behavior and product usage
through server-side event tracking.

CAPABILITIES:
  • Send individual server-side events with custom properties

EXAMPLES:
  # Send a basic event
  pup product-analytics events send \
    --app-id=my-app \
    --event=button_clicked

  # Send event with properties and user context
  pup product-analytics events send \
    --app-id=my-app \
    --event=purchase_completed \
    --properties='{"amount":99.99,"currency":"USD"}' \
    --user-id=user-123

AUTHENTICATION:
  Requires OAuth2 (via 'pup auth login') or valid API + Application keys.`,
}

var productAnalyticsEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Send product analytics events",
}

var productAnalyticsEventsSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a product analytics event",
	Long: `Send a single server-side product analytics event.

REQUIRED FLAGS:
  --app-id         Application ID
  --event          Event name

OPTIONAL FLAGS:
  --properties     Event properties as JSON object (default: {})
  --user-id        User ID

EXAMPLES:
  # Basic event
  pup product-analytics events send \
    --app-id=my-app \
    --event=page_view

  # Event with properties and user
  pup product-analytics events send \
    --app-id=my-app \
    --event=purchase_completed \
    --properties='{"product_id":"abc-123","amount":99.99}' \
    --user-id=user-456

  # Event with JSON output
  pup product-analytics events send \
    --app-id=my-app \
    --event=signup \
    --user-id=user-789 \
    --output=json`,
	RunE: runProductAnalyticsEventsSend,
}

var (
	paAppID      string
	paEventName  string
	paEventProps string
	paUserID     string
)

func init() {
	// Send event flags
	productAnalyticsEventsSendCmd.Flags().StringVar(&paAppID, "app-id", "", "Application ID (required)")
	productAnalyticsEventsSendCmd.Flags().StringVar(&paEventName, "event", "", "Event name (required)")
	productAnalyticsEventsSendCmd.Flags().StringVar(&paEventProps, "properties", "{}", "Event properties (JSON object)")
	productAnalyticsEventsSendCmd.Flags().StringVar(&paUserID, "user-id", "", "User ID")

	_ = productAnalyticsEventsSendCmd.MarkFlagRequired("app-id")
	_ = productAnalyticsEventsSendCmd.MarkFlagRequired("event")

	// Command hierarchy
	productAnalyticsEventsCmd.AddCommand(productAnalyticsEventsSendCmd)
	productAnalyticsCmd.AddCommand(productAnalyticsEventsCmd)
}

func runProductAnalyticsEventsSend(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Parse event properties JSON
	var properties map[string]interface{}
	if paEventProps != "" {
		if err := json.Unmarshal([]byte(paEventProps), &properties); err != nil {
			return fmt.Errorf("invalid event properties JSON: %w", err)
		}
	}

	// Build application object
	app := datadogV2.NewProductAnalyticsServerSideEventItemApplication(paAppID)

	// Build event object with name
	event := datadogV2.NewProductAnalyticsServerSideEventItemEvent(paEventName)

	// Add properties as additional properties on the event
	if properties != nil && len(properties) > 0 {
		event.AdditionalProperties = properties
	}

	// Build event item
	eventType := datadogV2.PRODUCTANALYTICSSERVERSIDEEVENTITEMTYPE_SERVER
	eventItem := datadogV2.NewProductAnalyticsServerSideEventItem(*app, *event, eventType)

	// Add user if provided
	if paUserID != "" {
		usr := datadogV2.NewProductAnalyticsServerSideEventItemUsr(paUserID)
		eventItem.SetUsr(*usr)
	}

	api := datadogV2.NewProductAnalyticsApi(client.V2())

	resp, r, err := api.SubmitProductAnalyticsEvent(client.Context(), *eventItem)
	if err != nil {
		return formatAPIError("submit product analytics event", err, r)
	}

	return formatAndPrint(resp, nil)
}
