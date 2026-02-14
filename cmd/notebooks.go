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
	"os"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var notebooksCmd = &cobra.Command{
	Use:   "notebooks",
	Short: "Manage notebooks",
	Long: `Manage Datadog notebooks for investigation and documentation.

Notebooks combine graphs, logs, and narrative text to document
investigations, share findings, and create runbooks.

CAPABILITIES:
  • List notebooks
  • Get notebook details
  • Create new notebooks
  • Update notebooks
  • Delete notebooks

EXAMPLES:
  # List all notebooks
  pup notebooks list

  # Get notebook details
  pup notebooks get notebook-id

  # Create a notebook from file
  pup notebooks create --body @notebook.json

  # Create from stdin
  cat notebook.json | pup notebooks create --body -

  # Update a notebook
  pup notebooks update 12345 --body @updated.json

  # Delete a notebook
  pup notebooks delete 12345

AUTHENTICATION:
  Requires API key authentication (DD_API_KEY + DD_APP_KEY).
  OAuth2 is not supported for this endpoint.`,
}

var notebooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notebooks",
	RunE:  runNotebooksList,
}

var notebooksGetCmd = &cobra.Command{
	Use:   "get [notebook-id]",
	Short: "Get notebook details",
	Args:  cobra.ExactArgs(1),
	RunE:  runNotebooksGet,
}

var notebooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new notebook",
	RunE:  runNotebooksCreate,
}

var notebooksUpdateCmd = &cobra.Command{
	Use:   "update [notebook-id]",
	Short: "Update a notebook",
	Args:  cobra.ExactArgs(1),
	RunE:  runNotebooksUpdate,
}

var notebooksDeleteCmd = &cobra.Command{
	Use:   "delete [notebook-id]",
	Short: "Delete a notebook",
	Args:  cobra.ExactArgs(1),
	RunE:  runNotebooksDelete,
}

var notebooksCellsCmd = &cobra.Command{
	Use:   "cells",
	Short: "Manage notebook cells",
	Long: `Manage individual cells within a Datadog notebook.

Cells are the building blocks of notebooks. Each cell can contain
markdown text, metric graphs, or log streams.`,
}

var notebooksCellsAppendCmd = &cobra.Command{
	Use:   "append [notebook-id]",
	Short: "Append cells to a notebook",
	Long: `Append one or more cells to an existing notebook.

Uses a simplified cell format for easy creation. Each cell has a type
(markdown, metric, or logs) and data content.

ARGUMENTS:
  notebook-id    The numeric notebook ID

FLAGS:
  --body         Cell definitions in simplified JSON (@filepath or - for stdin)

CELL TYPES:
  markdown   Text content with Markdown formatting
  metric     Timeseries graph using Datadog metrics query language
  logs       Log stream widget using Datadog logs query language

CELL FORMAT:
  {
    "cells": [
      {"type": "markdown", "data": "# Heading\nSome text..."},
      {"type": "metric", "data": "avg:system.cpu.user{*}", "title": "CPU Usage"},
      {"type": "logs", "data": "service:api status:error", "title": "API Errors"}
    ]
  }

  For metric and logs cells, optional time windows can be specified:
    "start": "2025-01-01T00:00:00Z"
    "end":   "2025-01-01T01:00:00Z"

  If omitted, metric/logs cells default to the last hour.

EXAMPLES:
  # Append a markdown cell from file
  pup notebooks cells append 12345 --body @cells.json

  # Append from stdin
  echo '{"cells":[{"type":"markdown","data":"# New Section"}]}' | \
    pup notebooks cells append 12345 --body -

  # Append a metric graph
  echo '{"cells":[{"type":"metric","data":"avg:system.load.1{*}","title":"Load"}]}' | \
    pup notebooks cells append 12345 --body -

AUTHENTICATION:
  Requires API key authentication (DD_API_KEY + DD_APP_KEY).`,
	Args: cobra.ExactArgs(1),
	RunE: runNotebooksCellsAppend,
}

// simpleCellType represents the type of a notebook cell.
type simpleCellType string

const (
	simpleCellMarkdown simpleCellType = "markdown"
	simpleCellMetric   simpleCellType = "metric"
	simpleCellLogs     simpleCellType = "logs"
)

// simpleCell is a simplified representation of a notebook cell,
// matching the format used by the Datadog MCP server.
type simpleCell struct {
	Type  simpleCellType `json:"type"`
	Data  string         `json:"data"`
	Start string         `json:"start,omitempty"`
	End   string         `json:"end,omitempty"`
	Title string         `json:"title,omitempty"`
}

// simpleCellsBody is the top-level JSON structure for cell input.
type simpleCellsBody struct {
	Cells []simpleCell `json:"cells"`
}

func init() {
	notebooksCreateCmd.Flags().String("body", "", "JSON body (@filepath or - for stdin) (required)")
	if err := notebooksCreateCmd.MarkFlagRequired("body"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	notebooksUpdateCmd.Flags().String("body", "", "JSON body (@filepath or - for stdin) (required)")
	if err := notebooksUpdateCmd.MarkFlagRequired("body"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	notebooksCellsAppendCmd.Flags().String("body", "", "Simplified cell JSON (@filepath or - for stdin) (required)")
	if err := notebooksCellsAppendCmd.MarkFlagRequired("body"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	notebooksCellsCmd.AddCommand(notebooksCellsAppendCmd)
	notebooksCmd.AddCommand(notebooksListCmd, notebooksGetCmd, notebooksCreateCmd, notebooksUpdateCmd, notebooksDeleteCmd, notebooksCellsCmd)
}

// readBody reads JSON body content from a file (@path) or stdin (-).
func readBody(value string) ([]byte, error) {
	var data []byte
	var err error

	switch {
	case value == "-":
		data, err = io.ReadAll(inputReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read body from stdin: %w", err)
		}
	case strings.HasPrefix(value, "@"):
		path := strings.TrimPrefix(value, "@")
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read body file: %w", err)
		}
	default:
		return nil, fmt.Errorf("body must be @filepath or - for stdin")
	}

	if !json.Valid(data) {
		return nil, fmt.Errorf("invalid JSON in body")
	}

	return data, nil
}

func runNotebooksCreate(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("POST", "/api/v1/notebooks")
	if err != nil {
		return err
	}

	bodyFlag, _ := cmd.Flags().GetString("body")
	data, err := readBody(bodyFlag)
	if err != nil {
		return err
	}

	var body datadogV1.NotebookCreateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse notebook create request: %w", err)
	}

	api := datadogV1.NewNotebooksApi(client.V1())
	resp, r, err := api.CreateNotebook(client.Context(), body)
	if err != nil {
		return formatAPIError("create notebook", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runNotebooksUpdate(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("PUT", "/api/v1/notebooks/")
	if err != nil {
		return err
	}

	notebookID := parseInt64(args[0])

	bodyFlag, _ := cmd.Flags().GetString("body")
	data, err := readBody(bodyFlag)
	if err != nil {
		return err
	}

	var body datadogV1.NotebookUpdateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse notebook update request: %w", err)
	}

	api := datadogV1.NewNotebooksApi(client.V1())
	resp, r, err := api.UpdateNotebook(client.Context(), notebookID, body)
	if err != nil {
		return formatAPIError("update notebook", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runNotebooksList(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("GET", "/api/v1/notebooks")
	if err != nil {
		return err
	}

	api := datadogV1.NewNotebooksApi(client.V1())
	resp, r, err := api.ListNotebooks(client.Context())
	if err != nil {
		return formatAPIError("list notebooks", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runNotebooksGet(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("GET", "/api/v1/notebooks/")
	if err != nil {
		return err
	}

	notebookID := parseInt64(args[0])
	api := datadogV1.NewNotebooksApi(client.V1())
	resp, r, err := api.GetNotebook(client.Context(), notebookID)
	if err != nil {
		return formatAPIError("get notebook", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runNotebooksDelete(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("DELETE", "/api/v1/notebooks/")
	if err != nil {
		return err
	}

	notebookID := parseInt64(args[0])
	if !cfg.AutoApprove {
		printOutput("⚠️  WARNING: This will permanently delete notebook %d\n", notebookID)
		printOutput("Are you sure you want to continue? (y/N): ")

		response, err := readConfirmation()
		if err != nil {
			printOutput("\nOperation cancelled\n")
			return nil
		}
		if response != "y" && response != "Y" {
			printOutput("Operation cancelled\n")
			return nil
		}
	}

	api := datadogV1.NewNotebooksApi(client.V1())
	r, err := api.DeleteNotebook(client.Context(), notebookID)
	if err != nil {
		return formatAPIError("delete notebook", err, r)
	}

	printOutput("Successfully deleted notebook %d\n", notebookID)
	return nil
}

func runNotebooksCellsAppend(cmd *cobra.Command, args []string) error {
	notebookID := args[0]

	bodyFlag, _ := cmd.Flags().GetString("body")
	data, err := readBody(bodyFlag)
	if err != nil {
		return err
	}

	var cellsBody simpleCellsBody
	if err := json.Unmarshal(data, &cellsBody); err != nil {
		return fmt.Errorf("failed to parse cells: %w", err)
	}
	if len(cellsBody.Cells) == 0 {
		return fmt.Errorf("no cells provided")
	}

	// Validate and build all cell requests before making any API calls
	cellReqs := make([]datadogV1.NotebookUpdateCell, len(cellsBody.Cells))
	for i, cell := range cellsBody.Cells {
		req, err := buildCellRequest(cell)
		if err != nil {
			return fmt.Errorf("cell %d: %w", i, err)
		}
		cellReqs[i] = req
	}

	client, err := getClientForEndpoint("POST", "/api/v1/notebooks/")
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/api/v1/notebooks/%s/cells", notebookID)

	for i, cellReq := range cellReqs {

		reqBody := struct {
			Data datadogV1.NotebookUpdateCell `json:"data"`
		}{Data: cellReq}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("cell %d: failed to marshal request: %w", i, err)
		}

		resp, err := client.RawRequest("POST", path, bytes.NewReader(bodyBytes))
		if err != nil {
			if i > 0 {
				return fmt.Errorf("partially appended %d of %d cells; cell %d failed: %w", i, len(cellsBody.Cells), i, err)
			}
			return fmt.Errorf("failed to append cell: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			respBody, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if i > 0 {
				return fmt.Errorf("partially appended %d of %d cells; cell %d returned status %d: %s", i, len(cellsBody.Cells), i, resp.StatusCode, string(respBody))
			}
			return fmt.Errorf("failed to append cell (status %d): %s", resp.StatusCode, string(respBody))
		}
		_ = resp.Body.Close()
	}

	printOutput("Successfully appended %d cell(s) to notebook %s\n", len(cellsBody.Cells), notebookID)
	return nil
}

// buildCellRequest converts a simpleCell into a NotebookUpdateCell for the append API.
func buildCellRequest(cell simpleCell) (datadogV1.NotebookUpdateCell, error) {
	switch cell.Type {
	case simpleCellMarkdown:
		return buildMarkdownCell(cell), nil
	case simpleCellMetric:
		return buildMetricCell(cell), nil
	case simpleCellLogs:
		return buildLogsCell(cell), nil
	default:
		return datadogV1.NotebookUpdateCell{}, fmt.Errorf("unknown cell type %q (expected: markdown, metric, logs)", cell.Type)
	}
}

func buildMarkdownCell(cell simpleCell) datadogV1.NotebookUpdateCell {
	createReq := datadogV1.NotebookCellCreateRequest{
		Attributes: datadogV1.NotebookCellCreateRequestAttributes{
			NotebookMarkdownCellAttributes: &datadogV1.NotebookMarkdownCellAttributes{
				Definition: datadogV1.NotebookMarkdownCellDefinition{
					Text: cell.Data,
					Type: datadogV1.NOTEBOOKMARKDOWNCELLDEFINITIONTYPE_MARKDOWN,
				},
			},
		},
		Type: datadogV1.NOTEBOOKCELLRESOURCETYPE_NOTEBOOK_CELLS,
	}
	return datadogV1.NotebookUpdateCell{NotebookCellCreateRequest: &createReq}
}

func buildMetricCell(cell simpleCell) datadogV1.NotebookUpdateCell {
	graphSize := datadogV1.NOTEBOOKGRAPHSIZE_MEDIUM
	showLegend := true
	graphType := datadogV1.TIMESERIESWIDGETDEFINITIONTYPE_TIMESERIES
	linearScale := "linear"
	line := datadogV1.WIDGETDISPLAYTYPE_LINE
	lineType := datadogV1.WIDGETLINETYPE_SOLID
	lineWidth := datadogV1.WIDGETLINEWIDTH_NORMAL
	palette := "dog_classic"
	title := cell.Title

	startTime, endTime := parseCellTimes(cell.Start, cell.End)
	createReq := datadogV1.NotebookCellCreateRequest{
		Attributes: datadogV1.NotebookCellCreateRequestAttributes{
			NotebookTimeseriesCellAttributes: &datadogV1.NotebookTimeseriesCellAttributes{
				Definition: datadogV1.TimeseriesWidgetDefinition{
					Requests: []datadogV1.TimeseriesWidgetRequest{{
						DisplayType: &line,
						Q:           &cell.Data,
						Style: &datadogV1.WidgetRequestStyle{
							LineType:  &lineType,
							LineWidth: &lineWidth,
							Palette:   &palette,
						},
					}},
					ShowLegend: &showLegend,
					Type:       graphType,
					Yaxis: &datadogV1.WidgetAxis{
						Scale: &linearScale,
					},
					Title: &title,
				},
				GraphSize: &graphSize,
				Time: *datadogV1.NewNullableNotebookCellTime(&datadogV1.NotebookCellTime{
					NotebookAbsoluteTime: datadogV1.NewNotebookAbsoluteTime(endTime, startTime),
				}),
			},
		},
		Type: datadogV1.NOTEBOOKCELLRESOURCETYPE_NOTEBOOK_CELLS,
	}
	return datadogV1.NotebookUpdateCell{NotebookCellCreateRequest: &createReq}
}

func buildLogsCell(cell simpleCell) datadogV1.NotebookUpdateCell {
	graphSize := datadogV1.NOTEBOOKGRAPHSIZE_MEDIUM
	messageDisplay := datadogV1.WIDGETMESSAGEDISPLAY_INLINE
	showDate := true
	showMessage := true
	title := "logs"
	if cell.Title != "" {
		title = cell.Title
	}
	textAlign := datadogV1.WIDGETTEXTALIGN_LEFT

	startTime, endTime := parseCellTimes(cell.Start, cell.End)
	createReq := datadogV1.NotebookCellCreateRequest{
		Attributes: datadogV1.NotebookCellCreateRequestAttributes{
			NotebookLogStreamCellAttributes: &datadogV1.NotebookLogStreamCellAttributes{
				Definition: datadogV1.LogStreamWidgetDefinition{
					Columns:           []string{"timestamp", "host", "service", "message"},
					MessageDisplay:    &messageDisplay,
					Query:             &cell.Data,
					ShowDateColumn:    &showDate,
					ShowMessageColumn: &showMessage,
					Sort: &datadogV1.WidgetFieldSort{
						Column: "timestamp",
						Order:  datadogV1.WIDGETSORT_ASCENDING,
					},
					Title:      &title,
					TitleAlign: &textAlign,
					Type:       datadogV1.LOGSTREAMWIDGETDEFINITIONTYPE_LOG_STREAM,
				},
				GraphSize: &graphSize,
				Time: *datadogV1.NewNullableNotebookCellTime(&datadogV1.NotebookCellTime{
					NotebookAbsoluteTime: datadogV1.NewNotebookAbsoluteTime(endTime, startTime),
				}),
			},
		},
		Type: datadogV1.NOTEBOOKCELLRESOURCETYPE_NOTEBOOK_CELLS,
	}
	return datadogV1.NotebookUpdateCell{NotebookCellCreateRequest: &createReq}
}

// parseCellTimes parses optional ISO 8601 start/end times for cells.
// Defaults to (now - 1h, now) if not specified.
func parseCellTimes(startStr, endStr string) (time.Time, time.Time) {
	now := time.Now()
	endTime := now
	startTime := now.Add(-time.Hour)

	if endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}
	if startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}

	return startTime, endTime
}
