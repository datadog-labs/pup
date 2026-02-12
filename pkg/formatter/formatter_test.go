// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package formatter

import (
	"errors"
	"strings"
	"testing"
)

func TestToJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		data         interface{}
		wantError    bool
		wantContains []string
	}{
		{
			name: "simple map",
			data: map[string]interface{}{
				"foo": "bar",
				"baz": 123,
			},
			wantError:    false,
			wantContains: []string{`"foo"`, `"bar"`, `"baz"`, `123`},
		},
		{
			name: "struct",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "Alice",
				Age:  30,
			},
			wantError:    false,
			wantContains: []string{`"name"`, `"Alice"`, `"age"`, `30`},
		},
		{
			name:         "array",
			data:         []string{"a", "b", "c"},
			wantError:    false,
			wantContains: []string{`"a"`, `"b"`, `"c"`},
		},
		{
			name:         "nil",
			data:         nil,
			wantError:    false,
			wantContains: []string{`null`},
		},
		{
			name:         "empty map",
			data:         map[string]interface{}{},
			wantError:    false,
			wantContains: []string{`{}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := ToJSON(tt.data)

			if tt.wantError {
				if err == nil {
					t.Error("ToJSON() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ToJSON() unexpected error: %v", err)
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("ToJSON() result missing %q. Got: %s", want, result)
				}
			}
		})
	}
}

func TestToTable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		data         interface{}
		wantError    bool
		wantContains []string
	}{
		{
			name: "map data",
			data: map[string]interface{}{
				"name":  "test",
				"value": 42,
			},
			wantError:    false,
			wantContains: []string{"name", "test", "value", "42"},
		},
		{
			name: "slice of maps",
			data: []interface{}{
				map[string]interface{}{
					"id":   1,
					"name": "test1",
				},
				map[string]interface{}{
					"id":   2,
					"name": "test2",
				},
			},
			wantError:    false,
			wantContains: []string{"ID", "NAME", "test1", "test2", "1", "2"},
		},
		{
			name:         "empty slice",
			data:         []interface{}{},
			wantError:    false,
			wantContains: []string{"No results found"},
		},
		{
			name: "API response wrapper with data array",
			data: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":     1,
						"title":  "Incident 1",
						"status": "active",
					},
					map[string]interface{}{
						"id":     2,
						"title":  "Incident 2",
						"status": "resolved",
					},
				},
				"meta": map[string]interface{}{
					"total": 2,
				},
			},
			wantError:    false,
			wantContains: []string{"ID", "TITLE", "STATUS", "Incident 1", "Incident 2", "active", "resolved"},
		},
		{
			name: "API response wrapper with single data object",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"id":     1,
					"title":  "Single Incident",
					"status": "active",
				},
				"meta": map[string]interface{}{
					"version": "1.0",
				},
			},
			wantError:    false,
			wantContains: []string{"id", "title", "status", "Single Incident", "active"},
		},
		{
			name: "JSON:API format with attributes",
			data: []interface{}{
				map[string]interface{}{
					"id":   "12345",
					"type": "incident",
					"attributes": map[string]interface{}{
						"title":      "Database timeout",
						"severity":   "SEV-2",
						"status":     "active",
						"created_at": "2024-01-15T10:30:00Z",
					},
				},
			},
			wantError: false,
			wantContains: []string{
				"12345", "incident", "Database timeout", "SEV-2", "active",
			},
		},
		{
			name: "JSON:API format with relationships",
			data: []interface{}{
				map[string]interface{}{
					"id":   "12345",
					"type": "incident",
					"attributes": map[string]interface{}{
						"title": "API Error",
					},
					"relationships": map[string]interface{}{
						"commander": map[string]interface{}{
							"data": map[string]interface{}{
								"id":   "user-123",
								"type": "user",
							},
						},
					},
				},
			},
			wantError: false,
			wantContains: []string{
				"12345", "incident", "API Error", "user-123",
			},
		},
		{
			name:         "nil data",
			data:         nil,
			wantError:    false,
			wantContains: []string{},
		},
		{
			name: "timeseries data with single series",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "0",
					"type": "timeseries_response",
					"attributes": map[string]interface{}{
						"times": []interface{}{
							float64(1704067200000),
							float64(1704067220000),
							float64(1704067240000),
						},
						"values": []interface{}{
							[]interface{}{22.5, 23.1, 22.8},
						},
						"series": []interface{}{
							map[string]interface{}{
								"query_index": 0,
								"group_tags":  []interface{}{},
							},
						},
					},
				},
			},
			wantError: false,
			wantContains: []string{
				"TIMESTAMP", "SERIES 0",
				"1704067200000", "22.5",
				"1704067220000", "23.1",
				"1704067240000", "22.8",
			},
		},
		{
			name: "timeseries data with multiple series",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "0",
					"type": "timeseries_response",
					"attributes": map[string]interface{}{
						"times": []interface{}{
							float64(1704067200000),
							float64(1704067220000),
						},
						"values": []interface{}{
							[]interface{}{10.5, 11.2},
							[]interface{}{20.3, 21.1},
						},
					},
				},
			},
			wantError: false,
			wantContains: []string{
				"TIMESTAMP", "SERIES 0", "SERIES 1",
				"1704067200000", "10.5", "20.3",
				"1704067220000", "11.2", "21.1",
			},
		},
		{
			name: "timeseries data with times only (no values)",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"times": []interface{}{
							float64(1704067200000),
							float64(1704067220000),
						},
					},
				},
			},
			wantError: false,
			wantContains: []string{
				"TIMESTAMP",
				"1704067200000",
				"1704067220000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := ToTable(tt.data)

			if tt.wantError {
				if err == nil {
					t.Error("ToTable() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ToTable() unexpected error: %v", err)
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("ToTable() result missing %q. Got: %s", want, result)
				}
			}
		})
	}
}

func TestToYAML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		data         interface{}
		wantError    bool
		wantContains []string
	}{
		{
			name: "map data",
			data: map[string]interface{}{
				"name":  "test",
				"value": 42,
			},
			wantError:    false,
			wantContains: []string{"name:", "test", "value:", "42"},
		},
		{
			name:         "slice data",
			data:         []string{"a", "b", "c"},
			wantError:    false,
			wantContains: []string{"a", "b", "c"},
		},
		{
			name:      "nil data",
			data:      nil,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := ToYAML(tt.data)

			if tt.wantError {
				if err == nil {
					t.Error("ToYAML() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ToYAML() unexpected error: %v", err)
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("ToYAML() result missing %q. Got: %s", want, result)
				}
			}
		})
	}
}

func TestFormatOutput(t *testing.T) {
	t.Parallel()
	data := map[string]string{"key": "value"}

	tests := []struct {
		name      string
		format    OutputFormat
		wantError bool
	}{
		{
			name:      "JSON format",
			format:    FormatJSON,
			wantError: false,
		},
		{
			name:      "Table format",
			format:    FormatTable,
			wantError: false,
		},
		{
			name:      "YAML format",
			format:    FormatYAML,
			wantError: false,
		},
		{
			name:      "unknown format defaults to JSON",
			format:    OutputFormat("unknown"),
			wantError: false,
		},
		{
			name:      "empty format defaults to JSON",
			format:    OutputFormat(""),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := FormatOutput(data, tt.format)

			if tt.wantError {
				if err == nil {
					t.Error("FormatOutput() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("FormatOutput() unexpected error: %v", err)
				return
			}

			if result == "" {
				t.Error("FormatOutput() returned empty string")
			}

			// All formats should contain the data
			if !strings.Contains(result, "key") {
				t.Error("FormatOutput() should contain data")
			}
		})
	}
}

func TestFormatError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "simple error",
			err:  errors.New("something went wrong"),
			want: "Error: something went wrong",
		},
		{
			name: "formatted error",
			err:  errors.New("failed to connect: connection refused"),
			want: "Error: failed to connect: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := FormatError(tt.err)
			if result != tt.want {
				t.Errorf("FormatError() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestFormatSuccess(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		message      string
		data         interface{}
		wantError    bool
		wantContains []string
	}{
		{
			name:      "success with data",
			message:   "Operation completed",
			data:      map[string]string{"result": "OK"},
			wantError: false,
			wantContains: []string{
				`"status"`,
				`"success"`,
				`"message"`,
				`"Operation completed"`,
				`"data"`,
				`"result"`,
				`"OK"`,
			},
		},
		{
			name:      "success without data",
			message:   "Done",
			data:      nil,
			wantError: false,
			wantContains: []string{
				`"status"`,
				`"success"`,
				`"message"`,
				`"Done"`,
			},
		},
		{
			name:      "success with array data",
			message:   "List retrieved",
			data:      []int{1, 2, 3},
			wantError: false,
			wantContains: []string{
				`"status"`,
				`"success"`,
				`"data"`,
				`1`,
				`2`,
				`3`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := FormatSuccess(tt.message, tt.data)

			if tt.wantError {
				if err == nil {
					t.Error("FormatSuccess() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("FormatSuccess() unexpected error: %v", err)
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("FormatSuccess() result missing %q. Got: %s", want, result)
				}
			}
		})
	}
}

func TestOutputFormat_Constants(t *testing.T) {
	t.Parallel()
	// Verify format constants are correctly defined
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want \"json\"", FormatJSON)
	}
	if FormatTable != "table" {
		t.Errorf("FormatTable = %q, want \"table\"", FormatTable)
	}
	if FormatYAML != "yaml" {
		t.Errorf("FormatYAML = %q, want \"yaml\"", FormatYAML)
	}
}

func TestFormatTableValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		val  interface{}
		want string
	}{
		{
			name: "nil value",
			val:  nil,
			want: "",
		},
		{
			name: "string short",
			val:  "hello",
			want: "hello",
		},
		{
			name: "string long truncated",
			val:  "this is a very long string that should be truncated because it exceeds fifty characters",
			want: "this is a very long string that should be trunc...",
		},
		{
			name: "string exactly 50 chars",
			val:  "12345678901234567890123456789012345678901234567890",
			want: "12345678901234567890123456789012345678901234567890",
		},
		{
			name: "empty array",
			val:  []interface{}{},
			want: "[]",
		},
		{
			name: "array with 1 item",
			val:  []interface{}{"item1"},
			want: "[item1]",
		},
		{
			name: "array with 3 items",
			val:  []interface{}{"a", "b", "c"},
			want: "[a, b, c]",
		},
		{
			name: "array with more than 3 items",
			val:  []interface{}{"a", "b", "c", "d", "e"},
			want: "[5 items]",
		},
		{
			name: "empty map",
			val:  map[string]interface{}{},
			want: "{0 fields}",
		},
		{
			name: "map with fields",
			val:  map[string]interface{}{"key1": "val1", "key2": "val2"},
			want: "{2 fields}",
		},
		{
			name: "float64 integer value",
			val:  float64(42),
			want: "42",
		},
		{
			name: "float64 decimal value",
			val:  float64(42.567),
			want: "42.57",
		},
		{
			name: "float64 negative integer",
			val:  float64(-10),
			want: "-10",
		},
		{
			name: "float64 negative decimal",
			val:  float64(-10.234),
			want: "-10.23",
		},
		{
			name: "bool true",
			val:  true,
			want: "true",
		},
		{
			name: "bool false",
			val:  false,
			want: "false",
		},
		{
			name: "int value",
			val:  123,
			want: "123",
		},
		{
			name: "int64 value",
			val:  int64(999),
			want: "999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatTableValue(tt.val)
			if result != tt.want {
				t.Errorf("formatTableValue(%v) = %q, want %q", tt.val, result, tt.want)
			}
		})
	}
}

func TestToJSON_Error(t *testing.T) {
	t.Parallel()
	// Test with unmarshalable data (channels can't be marshaled to JSON)
	ch := make(chan int)
	_, err := ToJSON(ch)
	if err == nil {
		t.Error("ToJSON() with channel should return error")
	}
	if !strings.Contains(err.Error(), "failed to marshal JSON") {
		t.Errorf("ToJSON() error should mention 'failed to marshal JSON', got: %v", err)
	}
}

// Note: ToYAML error path is difficult to test because yaml.v3 panics on unmarshalable
// types rather than returning errors. The error path in ToYAML would only be triggered
// by internal yaml library errors which are hard to simulate in a unit test.

func TestToTable_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		data         interface{}
		wantError    bool
		wantContains []string
	}{
		{
			name:         "slice of non-maps",
			data:         []interface{}{"string1", "string2", "string3"},
			wantError:    false,
			wantContains: []string{"string1", "string2", "string3"},
		},
		{
			name: "API response wrapper with empty data array",
			data: map[string]interface{}{
				"data": []interface{}{},
				"meta": map[string]interface{}{
					"total": 0,
				},
			},
			wantError:    false,
			wantContains: []string{"No results found"},
		},
		{
			name: "JSON:API single object with attributes but no timeseries",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "123",
					"type": "monitor",
					"attributes": map[string]interface{}{
						"name":   "CPU Monitor",
						"status": "OK",
					},
				},
			},
			wantError:    false,
			wantContains: []string{"id", "type", "name", "status", "123", "monitor", "CPU Monitor", "OK"},
		},
		{
			name: "JSON:API with array relationships",
			data: []interface{}{
				map[string]interface{}{
					"id":   "incident-1",
					"type": "incident",
					"attributes": map[string]interface{}{
						"title": "Outage",
					},
					"relationships": map[string]interface{}{
						"responders": map[string]interface{}{
							"data": []interface{}{
								map[string]interface{}{"id": "user-1", "type": "user"},
								map[string]interface{}{"id": "user-2", "type": "user"},
								map[string]interface{}{"id": "user-3", "type": "user"},
							},
						},
					},
				},
			},
			wantError:    false,
			wantContains: []string{"incident-1", "Outage", "user-1,user-2,user-3"},
		},
		{
			name:         "unsupported type (fails at marshal)",
			data:         func() {},
			wantError:    true, // Functions can't be marshaled to JSON
			wantContains: []string{},
		},
		{
			name: "API wrapper with single non-map data",
			data: map[string]interface{}{
				"data": "single value",
			},
			wantError:    false,
			wantContains: []string{"data", "single value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := ToTable(tt.data)

			if tt.wantError {
				if err == nil {
					t.Error("ToTable() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ToTable() unexpected error: %v", err)
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("ToTable() result missing %q. Got: %s", want, result)
				}
			}
		})
	}
}

func TestFlattenJSONAPIObject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		obj  map[string]interface{}
		want map[string]interface{}
	}{
		{
			name: "simple object with attributes",
			obj: map[string]interface{}{
				"id":   "123",
				"type": "incident",
				"attributes": map[string]interface{}{
					"title":  "Test Incident",
					"status": "active",
				},
			},
			want: map[string]interface{}{
				"id":     "123",
				"type":   "incident",
				"title":  "Test Incident",
				"status": "active",
			},
		},
		{
			name: "object with single relationship",
			obj: map[string]interface{}{
				"id":   "456",
				"type": "alert",
				"relationships": map[string]interface{}{
					"owner": map[string]interface{}{
						"data": map[string]interface{}{
							"id":   "user-123",
							"type": "user",
						},
					},
				},
			},
			want: map[string]interface{}{
				"id":         "456",
				"type":       "alert",
				"owner_id":   "user-123",
				"owner_type": "user",
			},
		},
		{
			name: "object with array relationship",
			obj: map[string]interface{}{
				"id":   "789",
				"type": "incident",
				"relationships": map[string]interface{}{
					"responders": map[string]interface{}{
						"data": []interface{}{
							map[string]interface{}{"id": "user-1"},
							map[string]interface{}{"id": "user-2"},
						},
					},
				},
			},
			want: map[string]interface{}{
				"id":             "789",
				"type":           "incident",
				"responders_ids": "user-1,user-2",
			},
		},
		{
			name: "object with empty relationship array",
			obj: map[string]interface{}{
				"id":   "999",
				"type": "monitor",
				"relationships": map[string]interface{}{
					"tags": map[string]interface{}{
						"data": []interface{}{},
					},
				},
			},
			want: map[string]interface{}{
				"id":   "999",
				"type": "monitor",
			},
		},
		{
			name: "object with attributes and relationships",
			obj: map[string]interface{}{
				"id":   "combo",
				"type": "service",
				"attributes": map[string]interface{}{
					"name":        "api-service",
					"environment": "production",
				},
				"relationships": map[string]interface{}{
					"team": map[string]interface{}{
						"data": map[string]interface{}{
							"id":   "team-42",
							"type": "team",
						},
					},
				},
			},
			want: map[string]interface{}{
				"id":          "combo",
				"type":        "service",
				"name":        "api-service",
				"environment": "production",
				"team_id":     "team-42",
				"team_type":   "team",
			},
		},
		{
			name: "relationship without data field",
			obj: map[string]interface{}{
				"id":   "rel-test",
				"type": "test",
				"relationships": map[string]interface{}{
					"link": map[string]interface{}{
						"href": "http://example.com",
					},
				},
			},
			want: map[string]interface{}{
				"id":   "rel-test",
				"type": "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := flattenJSONAPIObject(tt.obj)

			// Check all expected keys
			for key, expectedVal := range tt.want {
				if result[key] != expectedVal {
					t.Errorf("flattenJSONAPIObject() key %q = %v, want %v", key, result[key], expectedVal)
				}
			}

			// Check for unexpected keys
			for key := range result {
				if _, exists := tt.want[key]; !exists {
					t.Errorf("flattenJSONAPIObject() unexpected key %q with value %v", key, result[key])
				}
			}
		})
	}
}
