// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage host tags",
	Long: `Manage tags for hosts in your infrastructure.

Tags provide metadata about your hosts and help organize and filter
your infrastructure.

CAPABILITIES:
  • List all host tags
  • Get tags for a specific host
  • Add tags to a host
  • Update host tags
  • Remove tags from a host

EXAMPLES:
  # List all host tags
  pup tags list

  # Get tags for a host
  pup tags get my-host

  # Add tags to a host
  pup tags add my-host env:prod team:backend

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all host tags",
	RunE:  runTagsList,
}

var tagsGetCmd = &cobra.Command{
	Use:   "get [hostname]",
	Short: "Get tags for a host",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagsGet,
}

var tagsAddCmd = &cobra.Command{
	Use:   "add [hostname] [tags...]",
	Short: "Add tags to a host",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runTagsAdd,
}

var tagsUpdateCmd = &cobra.Command{
	Use:   "update [hostname] [tags...]",
	Short: "Update host tags",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runTagsUpdate,
}

var tagsDeleteCmd = &cobra.Command{
	Use:   "delete [hostname]",
	Short: "Delete all tags from a host",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagsDelete,
}

func init() {
	tagsCmd.AddCommand(tagsListCmd, tagsGetCmd, tagsAddCmd, tagsUpdateCmd, tagsDeleteCmd)
}

func runTagsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewTagsApi(client.V1())
	resp, r, err := api.ListHostTags(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list host tags: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list host tags: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runTagsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	hostname := args[0]
	api := datadogV1.NewTagsApi(client.V1())
	resp, r, err := api.GetHostTags(client.Context(), hostname)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get host tags: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get host tags: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runTagsAdd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	hostname := args[0]
	tags := args[1:]

	api := datadogV1.NewTagsApi(client.V1())
	body := datadogV1.HostTags{
		Tags: &tags,
	}

	resp, r, err := api.CreateHostTags(client.Context(), hostname, body)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to add host tags: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to add host tags: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runTagsUpdate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	hostname := args[0]
	tags := args[1:]

	api := datadogV1.NewTagsApi(client.V1())
	body := datadogV1.HostTags{
		Tags: &tags,
	}

	resp, r, err := api.UpdateHostTags(client.Context(), hostname, body)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to update host tags: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to update host tags: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runTagsDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	hostname := args[0]
	if !cfg.AutoApprove {
		fmt.Printf("⚠️  WARNING: This will delete all tags from host %s\n", hostname)
		fmt.Print("Are you sure you want to continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	api := datadogV1.NewTagsApi(client.V1())
	r, err := api.DeleteHostTags(client.Context(), hostname)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to delete host tags: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to delete host tags: %w", err)
	}

	fmt.Printf("Successfully deleted all tags from host %s\n", hostname)
	return nil
}
