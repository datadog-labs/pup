// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	// FormatJSON outputs in JSON format
	FormatJSON OutputFormat = "json"
	// FormatTable outputs in table format
	FormatTable OutputFormat = "table"
	// FormatYAML outputs in YAML format
	FormatYAML OutputFormat = "yaml"
)

// FormatOutput formats the output based on the specified format
func FormatOutput(data interface{}, format OutputFormat) (string, error) {
	switch format {
	case FormatJSON:
		return ToJSON(data)
	case FormatTable:
		return ToTable(data)
	case FormatYAML:
		return ToYAML(data)
	default:
		return ToJSON(data)
	}
}

// ToJSON formats data as JSON
func ToJSON(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(bytes), nil
}

// ToTable formats data as a table
func ToTable(data interface{}) (string, error) {
	if data == nil {
		return "", nil
	}

	// Convert to JSON first to normalize the data structure
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	// Parse back to generic structure
	var normalized interface{}
	if err := json.Unmarshal(jsonBytes, &normalized); err != nil {
		return "", fmt.Errorf("failed to unmarshal data: %w", err)
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)

	// Handle different data types
	switch v := normalized.(type) {
	case []interface{}:
		if len(v) == 0 {
			// Empty slice - check if original data was also empty
			return "No results found", nil
		}
		// Format as table with rows
		if err := formatSliceAsTable(table, v); err != nil {
			return "", err
		}
	case map[string]interface{}:
		// Check if this is an API response wrapper with "data" field
		// Common pattern: {"data": [...], "meta": {...}}
		if dataField, hasData := v["data"]; hasData {
			// Check if data is an array
			if dataArray, isArray := dataField.([]interface{}); isArray {
				if len(dataArray) == 0 {
					return "No results found", nil
				}
				// Format the data array as table instead of the wrapper
				if err := formatSliceAsTable(table, dataArray); err != nil {
					return "", err
				}
			} else {
				// data is a single object - check if it's JSON:API format
				if dataMap, isMap := dataField.(map[string]interface{}); isMap {
					// Check if this is a JSON:API object with attributes
					if attrs, hasAttrs := dataMap["attributes"].(map[string]interface{}); hasAttrs {
						// Check if this is timeseries data (has times and values/series)
						if times, hasTimes := attrs["times"].([]interface{}); hasTimes {
							if err := formatTimeseriesAsTable(table, attrs, times); err != nil {
								return "", err
							}
						} else {
							// Has attributes but not timeseries - flatten and display
							flattened := flattenJSONAPIObject(dataMap)
							if err := formatMapAsTable(table, flattened); err != nil {
								return "", err
							}
						}
					} else {
						// No attributes - format as key-value pairs
						if err := formatMapAsTable(table, dataMap); err != nil {
							return "", err
						}
					}
				} else {
					// Single object - format as key-value pairs
					if err := formatMapAsTable(table, v); err != nil {
						return "", err
					}
				}
			}
		} else {
			// No "data" field - format as key-value pairs
			if err := formatMapAsTable(table, v); err != nil {
				return "", err
			}
		}
	default:
		// Fallback to JSON for unknown types
		return fmt.Sprintf("Unsupported data type for table format: %T\nUse JSON format instead:\n%s", normalized, string(jsonBytes)), nil
	}

	if err := table.Render(); err != nil {
		return "", fmt.Errorf("failed to render table: %w", err)
	}
	return buf.String(), nil
}

// flattenJSONAPIObject flattens JSON:API style objects with attributes and relationships
func flattenJSONAPIObject(obj map[string]interface{}) map[string]interface{} {
	flattened := make(map[string]interface{})

	// Copy top-level fields (id, type, etc.)
	for key, val := range obj {
		if key != "attributes" && key != "relationships" {
			flattened[key] = val
		}
	}

	// Flatten attributes into top level
	if attrs, hasAttrs := obj["attributes"].(map[string]interface{}); hasAttrs {
		for key, val := range attrs {
			flattened[key] = val
		}
	}

	// Extract useful data from relationships
	if rels, hasRels := obj["relationships"].(map[string]interface{}); hasRels {
		for relName, relVal := range rels {
			if relMap, ok := relVal.(map[string]interface{}); ok {
				// Extract relationship data if present
				if relData, hasData := relMap["data"]; hasData {
					// Handle single relationship
					if relDataMap, ok := relData.(map[string]interface{}); ok {
						if id, hasID := relDataMap["id"]; hasID {
							flattened[relName+"_id"] = id
						}
						if relType, hasType := relDataMap["type"]; hasType {
							flattened[relName+"_type"] = relType
						}
					}
					// Handle array of relationships
					if relDataArray, ok := relData.([]interface{}); ok && len(relDataArray) > 0 {
						ids := make([]string, 0, len(relDataArray))
						for _, rel := range relDataArray {
							if relMap, ok := rel.(map[string]interface{}); ok {
								if id, ok := relMap["id"].(string); ok {
									ids = append(ids, id)
								}
							}
						}
						if len(ids) > 0 {
							flattened[relName+"_ids"] = strings.Join(ids, ",")
						}
					}
				}
			}
		}
	}

	return flattened
}

// formatTimeseriesAsTable formats timeseries data (times + values) as a table
func formatTimeseriesAsTable(table *tablewriter.Table, attrs map[string]interface{}, times []interface{}) error {
	// Extract values array - typically a 2D array [[series1_values], [series2_values], ...]
	var valuesArray [][]interface{}
	if values, hasValues := attrs["values"].([]interface{}); hasValues {
		// Convert to 2D array
		for _, seriesVals := range values {
			if seriesArr, ok := seriesVals.([]interface{}); ok {
				valuesArray = append(valuesArray, seriesArr)
			}
		}
	}

	// If no values array, show times only
	if len(valuesArray) == 0 {
		table.Header("Timestamp")
		for _, t := range times {
			if err := table.Append(formatTableValue(t)); err != nil {
				return fmt.Errorf("failed to append row: %w", err)
			}
		}
		return nil
	}

	// Build headers - one column for timestamp, one for each series
	headers := []interface{}{"Timestamp"}
	for i := range valuesArray {
		headers = append(headers, fmt.Sprintf("Series %d", i))
	}
	table.Header(headers...)

	// Build rows - one row per timestamp
	for i, timestamp := range times {
		row := []interface{}{formatTableValue(timestamp)}

		// Add values from each series for this timestamp
		for _, seriesVals := range valuesArray {
			if i < len(seriesVals) {
				row = append(row, formatTableValue(seriesVals[i]))
			} else {
				row = append(row, "")
			}
		}

		if err := table.Append(row...); err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
	}

	return nil
}

// formatSliceAsTable formats a slice of objects as a table
func formatSliceAsTable(table *tablewriter.Table, data []interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Get headers from first object
	if _, ok := data[0].(map[string]interface{}); !ok {
		// If not a map, just display as a list
		for _, item := range data {
			if err := table.Append(fmt.Sprintf("%v", item)); err != nil {
				return fmt.Errorf("failed to append row: %w", err)
			}
		}
		return nil
	}

	// Flatten JSON:API style objects (with attributes and relationships)
	flattenedData := make([]map[string]interface{}, len(data))
	for i, item := range data {
		if itemMap, ok := item.(map[string]interface{}); ok {
			flattenedData[i] = flattenJSONAPIObject(itemMap)
		} else {
			flattenedData[i] = make(map[string]interface{})
		}
	}

	// Collect all unique keys across all flattened items
	headerSet := make(map[string]bool)
	var headers []string
	for _, itemMap := range flattenedData {
		for key := range itemMap {
			if !headerSet[key] {
				headerSet[key] = true
				headers = append(headers, key)
			}
		}
	}

	// Limit columns for readability - prioritize common fields
	priorityFields := []string{
		"id", "title", "name", "type", "status", "state", "severity",
		"created_at", "updated_at", "created", "modified",
	}
	finalHeaders := []string{}
	for _, field := range priorityFields {
		if headerSet[field] {
			finalHeaders = append(finalHeaders, field)
		}
	}
	// Add remaining fields (up to 12 total columns for attributes-heavy responses)
	for _, field := range headers {
		if len(finalHeaders) >= 12 {
			break
		}
		found := false
		for _, f := range finalHeaders {
			if f == field {
				found = true
				break
			}
		}
		if !found {
			finalHeaders = append(finalHeaders, field)
		}
	}

	// Convert headers to interface{} slice
	headerInts := make([]interface{}, len(finalHeaders))
	for i, h := range finalHeaders {
		headerInts[i] = h
	}
	table.Header(headerInts...)

	// Add rows from flattened data
	for _, itemMap := range flattenedData {
		row := make([]interface{}, len(finalHeaders))
		for i, header := range finalHeaders {
			val := itemMap[header]
			row[i] = formatTableValue(val)
		}
		if err := table.Append(row...); err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
	}

	return nil
}

// formatMapAsTable formats a map as key-value pairs
func formatMapAsTable(table *tablewriter.Table, data map[string]interface{}) error {
	table.Header("Field", "Value")

	for key, value := range data {
		if err := table.Append(key, formatTableValue(value)); err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
	}
	return nil
}

// formatTableValue formats a value for table display
func formatTableValue(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		// Truncate long strings
		if len(v) > 50 {
			return v[:47] + "..."
		}
		return v
	case []interface{}:
		// Format arrays compactly
		if len(v) == 0 {
			return "[]"
		}
		if len(v) <= 3 {
			parts := make([]string, len(v))
			for i, item := range v {
				parts[i] = fmt.Sprintf("%v", item)
			}
			return "[" + strings.Join(parts, ", ") + "]"
		}
		return fmt.Sprintf("[%d items]", len(v))
	case map[string]interface{}:
		// Format objects compactly
		return fmt.Sprintf("{%d fields}", len(v))
	case float64:
		// Format numbers cleanly
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%.2f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToYAML formats data as YAML
func ToYAML(data interface{}) (string, error) {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(bytes), nil
}

// FormatError formats an error message
func FormatError(err error) string {
	return fmt.Sprintf("Error: %v", err)
}

// FormatSuccess formats a success message with optional data
func FormatSuccess(message string, data interface{}) (string, error) {
	result := map[string]interface{}{
		"status":  "success",
		"message": message,
	}
	if data != nil {
		result["data"] = data
	}
	return ToJSON(result)
}
