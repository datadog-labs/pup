// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage network monitoring",
	Long: `Query network monitoring data including flows and devices.

Network Performance Monitoring provides visibility into network traffic
flows between services, containers, and availability zones.

CAPABILITIES:
  • Query network flows
  • List network devices
  • View network metrics
  • Monitor network performance

EXAMPLES:
  # List network flows
  pup network flows list

  # List network devices
  pup network devices list

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var networkFlowsCmd = &cobra.Command{
	Use:   "flows",
	Short: "Query network flows",
}

var networkFlowsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List network flows",
	RunE:  runNetworkFlowsList,
}

var networkDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List network devices",
}

var networkDevicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List network devices",
	RunE:  runNetworkDevicesList,
}

func init() {
	networkFlowsCmd.AddCommand(networkFlowsListCmd)
	networkDevicesCmd.AddCommand(networkDevicesListCmd)
	networkCmd.AddCommand(networkFlowsCmd, networkDevicesCmd)
}

func runNetworkFlowsList(cmd *cobra.Command, args []string) error {
	result := map[string]interface{}{
		"data": []map[string]interface{}{},
		"meta": map[string]interface{}{
			"message": "Network flows list - API endpoint implementation pending",
		},
	}

	output, err := formatter.ToJSON(result)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runNetworkDevicesList(cmd *cobra.Command, args []string) error {
	result := map[string]interface{}{
		"data": []map[string]interface{}{},
		"meta": map[string]interface{}{
			"message": "Network devices list - API endpoint implementation pending",
		},
	}

	output, err := formatter.ToJSON(result)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
