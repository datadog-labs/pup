// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var apmCmd = &cobra.Command{
	Use:   "apm",
	Short: "Manage APM services and entities",
	Long: `Manage Datadog APM services and entities.

APM (Application Performance Monitoring) tracks your services, operations, and dependencies
to provide performance insights. This command provides access to dynamic operational data
about traced services, datastores, queues, and other APM entities.

DISTINCTION FROM SERVICE CATALOG:
  • service-catalog: Static metadata registry (ownership, definitions, documentation)
  • apm: Dynamic operational data (performance stats, traces, actual runtime behavior)

  Service catalog shows "what services exist and who owns them"
  APM shows "what's running, how it's performing, and what it's calling"

CAPABILITIES:
  • List services with performance statistics (requests, errors, latency)
  • Query entities with rich metadata (services, datastores, queues, inferred services)
  • List operations and resources (endpoints) for services
  • View service dependencies and flow maps with performance metrics

COMMAND GROUPS:
  services       List and query APM services with performance data
  entities       Query APM entities (services, datastores, queues, etc.)
  dependencies   View service dependencies and call relationships
  flow-map       Visualize service flow with performance metrics

EXAMPLES:
  # List services with stats
  pup apm services stats --start $(date -d '1 hour ago' +%s) --end $(date +%s)

  # Query entities with filtering
  pup apm entities list --start $(date -d '1 hour ago' +%s) --end $(date +%s) --env prod

  # View service dependencies
  pup apm dependencies list --env prod --start $(date -d '1 hour ago' +%s) --end $(date +%s)

AUTHENTICATION:
  Requires either OAuth2 authentication (pup auth login) or API keys
  (DD_API_KEY and DD_APP_KEY environment variables).`,
}

var apmServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage APM services",
	Long: `List and query APM services with performance data.

Services are auto-discovered from traces and represent instrumented applications.
Performance statistics include request rates, error rates, and latency percentiles.

SUBCOMMANDS:
  list        List APM services (basic info)
  stats       List services with performance statistics
  operations  List operations for a service
  resources   List resources (endpoints) for a service operation

EXAMPLES:
  # List all services
  pup apm services list

  # Get services with performance stats
  pup apm services stats --start $(date -d '1 hour ago' +%s) --end $(date +%s) --env prod

  # List operations for a service
  pup apm services operations web-server --start $(date -d '1 hour ago' +%s) --end $(date +%s)

  # List resources for a service operation
  pup apm services resources web-server --operation "GET /api/users" --from $(date -d '1 hour ago' +%s) --to $(date +%s)`,
}

var apmServicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List APM services",
	Long: `List APM services with basic information.

This command retrieves services that have been auto-discovered from APM traces.
For performance statistics, use 'pup apm services stats' instead.

FLAGS:
  --env        Filter by environment (e.g., "prod", "staging")
  --start      Start time (Unix timestamp, default: 1 hour ago)
  --end        End time (Unix timestamp, default: now)

EXAMPLES:
  # List all services
  pup apm services list

  # Filter by environment
  pup apm services list --env prod

  # Custom time range
  pup apm services list --start $(date -d '2 hours ago' +%s) --end $(date +%s)

OUTPUT:
  Returns list of services with service names and basic metadata.`,
	RunE: runAPMServicesList,
}

var apmServicesStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "List services with performance statistics",
	Long: `List APM services with detailed performance statistics.

This command provides comprehensive performance metrics for each service including:
  • Request rate (hits per second)
  • Error rate (percentage and count)
  • Latency percentiles (p50, p75, p90, p95, p99)
  • Maximum latency

FLAGS:
  --env           Filter by environment (e.g., "prod", "staging")
  --primary-tag   Filter by primary tag (format: "group:value")
  --start         Start time (Unix timestamp) [REQUIRED]
  --end           End time (Unix timestamp) [REQUIRED]

EXAMPLES:
  # Get service stats for the last hour
  pup apm services stats --start $(date -d '1 hour ago' +%s) --end $(date +%s)

  # Filter by environment
  pup apm services stats --start $(date -d '1 hour ago' +%s) --end $(date +%s) --env prod

  # Filter by primary tag
  pup apm services stats --start $(date -d '1 hour ago' +%s) --end $(date +%s) --primary-tag "team:backend"

OUTPUT:
  Returns services with performance metrics:
  • Service name and environment
  • Hits per second
  • Error rate (percentage and count)
  • Latency percentiles (p50, p75, p90, p95, p99)
  • Max latency`,
	RunE: runAPMServicesStats,
}

var apmServicesOperationsCmd = &cobra.Command{
	Use:   "operations <service>",
	Short: "List operations for a service",
	Long: `List operations (spans) for a specific APM service.

Operations represent different types of work performed by a service (e.g., HTTP requests,
database queries, cache operations). Each operation has a name, service, span kind, and type.

ARGUMENTS:
  service    Service name (required)

FLAGS:
  --env           Filter by environment
  --primary-tag   Filter by primary tag
  --primary-only  Only return primary operations (default: false)
  --start         Start time (Unix timestamp) [REQUIRED]
  --end           End time (Unix timestamp) [REQUIRED]

EXAMPLES:
  # List operations for a service
  pup apm services operations web-server --start $(date -d '1 hour ago' +%s) --end $(date +%s)

  # Filter by environment
  pup apm services operations web-server --start $(date -d '1 hour ago' +%s) --end $(date +%s) --env prod

  # Only primary operations
  pup apm services operations web-server --start $(date -d '1 hour ago' +%s) --end $(date +%s) --primary-only

OUTPUT:
  Returns list of operations with:
  • Operation name
  • Service name
  • Span kind (server, client, producer, consumer, internal)
  • Type (web, db, cache, custom)`,
	Args: cobra.ExactArgs(1),
	RunE: runAPMServicesOperations,
}

var apmServicesResourcesCmd = &cobra.Command{
	Use:   "resources <service>",
	Short: "List resources (endpoints) for a service operation",
	Long: `List resources (endpoints) for a specific service and operation.

Resources represent specific endpoints or queries within an operation, such as:
  • HTTP endpoints: "GET /api/users", "POST /orders"
  • Database queries: "SELECT FROM users WHERE id = ?"
  • Cache operations: "redis.get user:123"

ARGUMENTS:
  service    Service name (required)

FLAGS:
  --operation     Operation name [REQUIRED]
  --env           Filter by environment
  --primary-tag   Filter by primary tag
  --peer-service  Filter by peer service
  --from, -f      Start time (Unix timestamp) [REQUIRED]
  --to, -t        End time (Unix timestamp) [REQUIRED]

EXAMPLES:
  # List resources for a service operation
  pup apm services resources web-server --operation "GET /api/users" --from $(date -d '1 hour ago' +%s) --to $(date +%s)

  # Filter by environment
  pup apm services resources web-server --operation "GET /api/users" --from $(date -d '1 hour ago' +%s) --to $(date +%s) --env prod

  # Filter by peer service
  pup apm services resources web-server --operation "GET /api/users" --from $(date -d '1 hour ago' +%s) --to $(date +%s) --peer-service "database"

OUTPUT:
  Returns list of resources with:
  • Resource name (endpoint/query)
  • Resource hash
  • Service name
  • Top-level operation name`,
	Args: cobra.ExactArgs(1),
	RunE: runAPMServicesResources,
}

var apmEntitiesCmd = &cobra.Command{
	Use:   "entities",
	Short: "Manage APM entities",
	Long: `Query APM entities with rich metadata.

Entities represent all types of APM resources including:
  • Services (instrumented applications)
  • Datastores (databases, caches)
  • Queues (message queues, event streams)
  • Inferred services (auto-detected external dependencies)

SUBCOMMANDS:
  list    Query entities with filtering and metadata

EXAMPLES:
  # List all entities
  pup apm entities list --start $(date -d '1 hour ago' +%s) --end $(date +%s)

  # Filter by type and environment
  pup apm entities list --start $(date -d '1 hour ago' +%s) --end $(date +%s) --env prod --types service

  # Include additional metadata
  pup apm entities list --start $(date -d '1 hour ago' +%s) --end $(date +%s) --include stats,health`,
}

var apmEntitiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Query APM entities",
	Long: `Query APM entities with filtering and rich metadata.

⚠️  WARNING: This command uses an UNSTABLE API endpoint that may change.

Entities can be filtered by type, environment, primary tags, and more. Additional
metadata can be included such as performance statistics, health status, and incidents.

FLAGS:
  --env           Filter by environment
  --primary-tag   Filter by primary tag
  --types         Comma-separated entity types (service, datastore, queue, inferred)
  --include       Comma-separated fields to include (stats, incidents, health)
  --limit         Maximum results (default: 50)
  --offset        Pagination offset (default: 0)
  --start         Start time (Unix timestamp) [REQUIRED]
  --end           End time (Unix timestamp) [REQUIRED]

ENTITY TYPES:
  • service: Instrumented applications
  • datastore: Databases, caches, storage systems
  • queue: Message queues, event streams
  • inferred: Auto-detected external dependencies

INCLUDE FIELDS:
  • stats: Performance statistics (request rate, error rate, latency)
  • health: Health status and scores
  • incidents: Related incidents and alerts

EXAMPLES:
  # List all entities
  pup apm entities list --start $(date -d '1 hour ago' +%s) --end $(date +%s)

  # Filter by environment and type
  pup apm entities list --start $(date -d '1 hour ago' +%s) --end $(date +%s) --env prod --types service

  # Include performance stats
  pup apm entities list --start $(date -d '1 hour ago' +%s) --end $(date +%s) --include stats,health

  # Query datastores only
  pup apm entities list --start $(date -d '1 hour ago' +%s) --end $(date +%s) --types datastore --limit 10

OUTPUT:
  Returns entities with:
  • Entity name and type
  • Environment and tags
  • Optional: performance stats, health status, incidents (based on --include)`,
	RunE: runAPMEntitiesList,
}

var apmDependenciesCmd = &cobra.Command{
	Use:   "dependencies",
	Short: "Manage service dependencies",
	Long: `View service dependencies and call relationships.

Dependencies show which services call other services, based on actual trace data.
This provides a real-time view of service communication patterns.

SUBCOMMANDS:
  list    List service dependencies (all or specific service)

EXAMPLES:
  # List all service dependencies
  pup apm dependencies list --env prod --start $(date -d '1 hour ago' +%s) --end $(date +%s)

  # List dependencies for a specific service
  pup apm dependencies list web-server --env prod --start $(date -d '1 hour ago' +%s) --end $(date +%s)`,
}

var apmDependenciesListCmd = &cobra.Command{
	Use:   "list [service]",
	Short: "List service dependencies",
	Long: `List service dependencies showing call relationships.

Without arguments, lists all service dependencies across the environment.
With a service argument, shows what that service calls and what calls it.

ARGUMENTS:
  service    Optional service name (if omitted, lists all dependencies)

FLAGS:
  --env           Environment filter [REQUIRED]
  --primary-tag   Filter by primary tag
  --start         Start time (Unix timestamp) [REQUIRED]
  --end           End time (Unix timestamp) [REQUIRED]

EXAMPLES:
  # List all service dependencies
  pup apm dependencies list --env prod --start $(date -d '1 hour ago' +%s) --end $(date +%s)

  # List dependencies for a specific service
  pup apm dependencies list web-server --env prod --start $(date -d '1 hour ago' +%s) --end $(date +%s)

  # Filter by primary tag
  pup apm dependencies list --env prod --primary-tag "team:backend" --start $(date -d '1 hour ago' +%s) --end $(date +%s)

OUTPUT (All dependencies):
  Returns map of service -> {calls: [...], called_by: [...]}

OUTPUT (Specific service):
  Returns {name: "service", calls: [...], called_by: [...]}`,
	RunE: runAPMDependenciesList,
}

var apmFlowMapCmd = &cobra.Command{
	Use:   "flow-map",
	Short: "View service flow map",
	Long: `Visualize service flow with performance metrics.

The flow map shows how services communicate with each other, including:
  • Nodes: Services with their performance metrics
  • Edges: Service calls with request rate, error rate, and latency

This provides a visual representation of service architecture and data flow patterns.

FLAGS:
  --query         Query filter [REQUIRED] (e.g., "env:prod", "service:web-server")
  --limit         Maximum nodes to return (default: 100)
  --from, -f      Start time (Unix timestamp) [REQUIRED]
  --to, -t        End time (Unix timestamp) [REQUIRED]

EXAMPLES:
  # Get flow map for production environment
  pup apm flow-map --query "env:prod" --from $(date -d '1 hour ago' +%s) --to $(date +%s)

  # Focus on a specific service
  pup apm flow-map --query "env:prod service:web-server" --from $(date -d '1 hour ago' +%s) --to $(date +%s)

  # Limit number of nodes
  pup apm flow-map --query "env:prod" --from $(date -d '1 hour ago' +%s) --to $(date +%s) --limit 50

OUTPUT:
  Returns nodes (services) and edges (calls) with metrics:
  • Nodes: Service name with performance metrics
  • Edges: Source and target services with:
    - Hits per second
    - Error rate
    - Latency percentiles (p50, p75, p90, p95, p99)
    - Max latency`,
	RunE: runAPMFlowMap,
}

var (
	// Time range flags
	startTime int64
	endTime   int64

	// Filter flags
	envFilter   string
	primaryTag  string
	entityTypes string

	// Pagination flags
	pageLimit  int
	pageOffset int

	// Include flags (for entities)
	includeFields string

	// Operation/resource specific flags
	operationName string
	peerService   string
	primaryOnly   bool

	// Flow map specific flags
	flowMapQuery string
	flowMapLimit int
)

func init() {
	// Services list flags
	apmServicesListCmd.Flags().StringVar(&envFilter, "env", "prod", "Environment filter")
	apmServicesListCmd.Flags().Int64Var(&startTime, "start", time.Now().Add(-1*time.Hour).Unix(), "Start time (Unix timestamp)")
	apmServicesListCmd.Flags().Int64Var(&endTime, "end", time.Now().Unix(), "End time (Unix timestamp)")

	// Services stats flags
	apmServicesStatsCmd.Flags().StringVar(&envFilter, "env", "", "Environment filter")
	apmServicesStatsCmd.Flags().StringVar(&primaryTag, "primary-tag", "", "Primary tag (group:value)")
	apmServicesStatsCmd.Flags().Int64Var(&startTime, "start", 0, "Start time (Unix timestamp)")
	apmServicesStatsCmd.Flags().Int64Var(&endTime, "end", 0, "End time (Unix timestamp)")
	if err := apmServicesStatsCmd.MarkFlagRequired("start"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := apmServicesStatsCmd.MarkFlagRequired("end"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// Services operations flags
	apmServicesOperationsCmd.Flags().StringVar(&envFilter, "env", "", "Environment filter")
	apmServicesOperationsCmd.Flags().StringVar(&primaryTag, "primary-tag", "", "Primary tag")
	apmServicesOperationsCmd.Flags().BoolVar(&primaryOnly, "primary-only", false, "Only primary operations")
	apmServicesOperationsCmd.Flags().Int64Var(&startTime, "start", 0, "Start time (Unix timestamp)")
	apmServicesOperationsCmd.Flags().Int64Var(&endTime, "end", 0, "End time (Unix timestamp)")
	if err := apmServicesOperationsCmd.MarkFlagRequired("start"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := apmServicesOperationsCmd.MarkFlagRequired("end"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// Services resources flags
	apmServicesResourcesCmd.Flags().StringVar(&operationName, "operation", "", "Operation name (required)")
	apmServicesResourcesCmd.Flags().StringVar(&envFilter, "env", "", "Environment filter")
	apmServicesResourcesCmd.Flags().StringVar(&primaryTag, "primary-tag", "", "Primary tag")
	apmServicesResourcesCmd.Flags().StringVar(&peerService, "peer-service", "", "Peer service filter")
	apmServicesResourcesCmd.Flags().Int64VarP(&startTime, "from", "f", 0, "Start time (Unix timestamp)")
	apmServicesResourcesCmd.Flags().Int64VarP(&endTime, "to", "t", 0, "End time (Unix timestamp)")
	if err := apmServicesResourcesCmd.MarkFlagRequired("operation"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := apmServicesResourcesCmd.MarkFlagRequired("from"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := apmServicesResourcesCmd.MarkFlagRequired("to"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// Entities list flags
	apmEntitiesListCmd.Flags().StringVar(&envFilter, "env", "", "Environment filter")
	apmEntitiesListCmd.Flags().StringVar(&primaryTag, "primary-tag", "", "Primary tag")
	apmEntitiesListCmd.Flags().StringVar(&entityTypes, "types", "", "Entity types (comma-separated)")
	apmEntitiesListCmd.Flags().StringVar(&includeFields, "include", "", "Fields to include (comma-separated)")
	apmEntitiesListCmd.Flags().IntVar(&pageLimit, "limit", 50, "Max results")
	apmEntitiesListCmd.Flags().IntVar(&pageOffset, "offset", 0, "Page offset")
	apmEntitiesListCmd.Flags().Int64Var(&startTime, "start", 0, "Start time (Unix timestamp)")
	apmEntitiesListCmd.Flags().Int64Var(&endTime, "end", 0, "End time (Unix timestamp)")
	if err := apmEntitiesListCmd.MarkFlagRequired("start"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := apmEntitiesListCmd.MarkFlagRequired("end"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// Dependencies list flags
	apmDependenciesListCmd.Flags().StringVar(&envFilter, "env", "", "Environment filter (required)")
	apmDependenciesListCmd.Flags().StringVar(&primaryTag, "primary-tag", "", "Primary tag")
	apmDependenciesListCmd.Flags().Int64Var(&startTime, "start", 0, "Start time (Unix timestamp)")
	apmDependenciesListCmd.Flags().Int64Var(&endTime, "end", 0, "End time (Unix timestamp)")
	if err := apmDependenciesListCmd.MarkFlagRequired("env"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := apmDependenciesListCmd.MarkFlagRequired("start"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := apmDependenciesListCmd.MarkFlagRequired("end"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// Flow map flags
	apmFlowMapCmd.Flags().StringVar(&flowMapQuery, "query", "", "Query filter (required)")
	apmFlowMapCmd.Flags().IntVar(&flowMapLimit, "limit", 100, "Max nodes")
	apmFlowMapCmd.Flags().Int64VarP(&startTime, "from", "f", 0, "Start time (Unix timestamp)")
	apmFlowMapCmd.Flags().Int64VarP(&endTime, "to", "t", 0, "End time (Unix timestamp)")
	if err := apmFlowMapCmd.MarkFlagRequired("query"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := apmFlowMapCmd.MarkFlagRequired("from"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := apmFlowMapCmd.MarkFlagRequired("to"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// Register subcommands
	apmServicesCmd.AddCommand(
		apmServicesListCmd,
		apmServicesStatsCmd,
		apmServicesOperationsCmd,
		apmServicesResourcesCmd,
	)
	apmEntitiesCmd.AddCommand(apmEntitiesListCmd)
	apmDependenciesCmd.AddCommand(apmDependenciesListCmd)

	apmCmd.AddCommand(
		apmServicesCmd,
		apmEntitiesCmd,
		apmDependenciesCmd,
		apmFlowMapCmd,
	)
}

func runAPMServicesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Build query parameters
	params := url.Values{}
	params.Add("start", strconv.FormatInt(startTime, 10))
	params.Add("end", strconv.FormatInt(endTime, 10))
	if envFilter != "" {
		params.Add("filter[env]", envFilter)
	}

	path := fmt.Sprintf("/api/v2/apm/services?%s", params.Encode())
	resp, err := client.RawRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to list APM services: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	result, err := readRawResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to list APM services: %w", err)
	}

	output, err := formatter.FormatOutput(result, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runAPMServicesStats(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Validate time range
	if startTime >= endTime {
		return fmt.Errorf("start time must be before end time")
	}

	// Build query parameters
	params := url.Values{}
	params.Add("start", strconv.FormatInt(startTime, 10))
	params.Add("end", strconv.FormatInt(endTime, 10))
	if envFilter != "" {
		params.Add("filter[env]", envFilter)
	}
	if primaryTag != "" {
		params.Add("filter[primary_tag]", primaryTag)
	}

	path := fmt.Sprintf("/api/v2/apm/services/stats?%s", params.Encode())
	resp, err := client.RawRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to get service stats: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	result, err := readRawResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to get service stats: %w", err)
	}

	output, err := formatter.FormatOutput(result, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runAPMServicesOperations(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	serviceName := args[0]

	// Validate time range
	if startTime >= endTime {
		return fmt.Errorf("start time must be before end time")
	}

	// Build query parameters
	params := url.Values{}
	params.Add("start", strconv.FormatInt(startTime, 10))
	params.Add("end", strconv.FormatInt(endTime, 10))
	params.Add("service", serviceName)
	if envFilter != "" {
		params.Add("env", envFilter)
	}
	if primaryTag != "" {
		params.Add("primary_tag", primaryTag)
	}
	if primaryOnly {
		params.Add("primary_only", "true")
	}

	path := fmt.Sprintf("/api/v1/trace/operation_names/%s?%s", url.PathEscape(serviceName), params.Encode())
	resp, err := client.RawRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to get service operations: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	result, err := readRawResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to get service operations: %w", err)
	}

	output, err := formatter.FormatOutput(result, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runAPMServicesResources(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	serviceName := args[0]

	// Validate time range
	if startTime >= endTime {
		return fmt.Errorf("start time must be before end time")
	}

	// Build query parameters
	params := url.Values{}
	params.Add("from", strconv.FormatInt(startTime, 10))
	params.Add("to", strconv.FormatInt(endTime, 10))
	params.Add("service", serviceName)
	params.Add("name", operationName)
	if envFilter != "" {
		params.Add("env", envFilter)
	}
	if primaryTag != "" {
		params.Add("primary_tag", primaryTag)
	}
	if peerService != "" {
		params.Add("peer.service", peerService)
	}

	path := fmt.Sprintf("/api/ui/apm/resources?%s", params.Encode())
	resp, err := client.RawRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to get service resources: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	result, err := readRawResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to get service resources: %w", err)
	}

	output, err := formatter.FormatOutput(result, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runAPMEntitiesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Validate time range
	if startTime >= endTime {
		return fmt.Errorf("start time must be before end time")
	}

	// Build query parameters
	params := url.Values{}
	params.Add("start", strconv.FormatInt(startTime, 10))
	params.Add("end", strconv.FormatInt(endTime, 10))
	if envFilter != "" {
		params.Add("filter[env]", envFilter)
	}
	if primaryTag != "" {
		params.Add("filter[primary_tag]", primaryTag)
	}
	if entityTypes != "" {
		params.Add("filter[entity.type.catalog.kind]", entityTypes)
	}
	if includeFields != "" {
		params.Add("include", includeFields)
	}
	params.Add("page[limit]", strconv.Itoa(pageLimit))
	params.Add("page[offset]", strconv.Itoa(pageOffset))

	path := fmt.Sprintf("/api/unstable/apm/entities?%s", params.Encode())
	resp, err := client.RawRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to list APM entities: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	result, err := readRawResponse(resp)
	if err != nil {
		// Add special message for unstable API 403 errors
		if resp.StatusCode == 403 {
			return fmt.Errorf("failed to list APM entities: %w\n\n⚠️  This endpoint uses an unstable API that may require feature flag enablement", err)
		}
		return fmt.Errorf("failed to list APM entities: %w", err)
	}

	output, err := formatter.FormatOutput(result, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runAPMDependenciesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Validate time range
	if startTime >= endTime {
		return fmt.Errorf("start time must be before end time")
	}

	// Build query parameters
	params := url.Values{}
	params.Add("start", strconv.FormatInt(startTime, 10))
	params.Add("end", strconv.FormatInt(endTime, 10))
	params.Add("env", envFilter)
	if primaryTag != "" {
		params.Add("primary_tag", primaryTag)
	}

	var path string
	if len(args) == 0 {
		// All dependencies
		path = fmt.Sprintf("/api/v1/service_dependencies?%s", params.Encode())
	} else {
		// Specific service
		serviceName := args[0]
		path = fmt.Sprintf("/api/v1/service_dependencies/%s?%s", url.PathEscape(serviceName), params.Encode())
	}

	resp, err := client.RawRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to list service dependencies: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	result, err := readRawResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to list service dependencies: %w", err)
	}

	output, err := formatter.FormatOutput(result, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runAPMFlowMap(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Validate time range
	if startTime >= endTime {
		return fmt.Errorf("start time must be before end time")
	}

	// Build query parameters
	params := url.Values{}
	params.Add("from", strconv.FormatInt(startTime, 10))
	params.Add("to", strconv.FormatInt(endTime, 10))
	params.Add("query", flowMapQuery)
	params.Add("limit", strconv.Itoa(flowMapLimit))

	path := fmt.Sprintf("/api/ui/apm/flow-map?%s", params.Encode())
	resp, err := client.RawRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to get flow map: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	result, err := readRawResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to get flow map: %w", err)
	}

	output, err := formatter.FormatOutput(result, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}
