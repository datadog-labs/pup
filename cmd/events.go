// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Manage Datadog events",
	Long: `Query and search Datadog events.

Events represent important occurrences in your infrastructure such as
deployments, configuration changes, alerts, and custom events.

CAPABILITIES:
  • List recent events
  • Search events with queries
  • Get event details

EXAMPLES:
  # List recent events
  pup events list

  # Search for deployment events
  pup events search --query="tags:deployment"

  # Get specific event
  pup events get 1234567890

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var eventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent events",
	RunE:  runEventsList,
}

var eventsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search events",
	RunE:  runEventsSearch,
}

var eventsGetCmd = &cobra.Command{
	Use:   "get [event-id]",
	Short: "Get event details",
	Args:  cobra.ExactArgs(1),
	RunE:  runEventsGet,
}

var (
	eventsStart  int64
	eventsEnd    int64
	eventsTags   string
	eventsQuery  string
	eventsFilter string
	eventsFrom   string
	eventsTo     string
	eventsLimit  int32
)

func init() {
	eventsListCmd.Flags().Int64Var(&eventsStart, "start", 0, "Start timestamp")
	eventsListCmd.Flags().Int64Var(&eventsEnd, "end", 0, "End timestamp")
	eventsListCmd.Flags().StringVar(&eventsTags, "tags", "", "Filter by tags")

	eventsSearchCmd.Flags().StringVar(&eventsQuery, "query", "", "Search query")
	eventsSearchCmd.Flags().StringVar(&eventsFilter, "filter", "", "Filter query")
	eventsSearchCmd.Flags().StringVar(&eventsFrom, "from", "", "Start time")
	eventsSearchCmd.Flags().StringVar(&eventsTo, "to", "", "End time")
	eventsSearchCmd.Flags().Int32Var(&eventsLimit, "limit", 100, "Maximum results")

	eventsCmd.AddCommand(eventsListCmd, eventsSearchCmd, eventsGetCmd)
}

func runEventsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewEventsApi(client.V1())
	opts := datadogV1.ListEventsOptionalParameters{}
	if eventsStart > 0 {
		opts.WithStart(eventsStart)
	}
	if eventsEnd > 0 {
		opts.WithEnd(eventsEnd)
	}
	if eventsTags != "" {
		opts.WithTags(eventsTags)
	}

	resp, r, err := api.ListEvents(client.Context(), opts)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list events: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list events: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runEventsSearch(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewEventsApi(client.V2())
	opts := datadogV2.SearchEventsOptionalParameters{}

	if eventsQuery != "" || eventsFilter != "" {
		body := datadogV2.EventsListRequest{}
		filter := datadogV2.EventsQueryFilter{}

		if eventsQuery != "" {
			filter.SetQuery(eventsQuery)
		}
		if eventsFilter != "" {
			filter.SetQuery(eventsFilter)
		}
		if eventsFrom != "" {
			filter.SetFrom(eventsFrom)
		}
		if eventsTo != "" {
			filter.SetTo(eventsTo)
		}

		body.SetFilter(filter)

		if eventsLimit > 0 {
			page := datadogV2.EventsRequestPage{}
			page.SetLimit(eventsLimit)
			body.SetPage(page)
		}

		opts.WithBody(body)
	}

	resp, r, err := api.SearchEvents(client.Context(), opts)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to search events: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to search events: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runEventsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	eventID := parseInt64(args[0])
	api := datadogV1.NewEventsApi(client.V1())
	resp, r, err := api.GetEvent(client.Context(), eventID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get event: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get event: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
