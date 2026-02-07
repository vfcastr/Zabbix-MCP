// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package hosts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
	"github.com/vfcastr/Zabbix-MCP/pkg/utils"
)

// Host represents a Zabbix host
type Host struct {
	HostID      string                 `json:"hostid"`
	Host        string                 `json:"host"`
	Name        string                 `json:"name"`
	Status      string                 `json:"status"`
	Description string                 `json:"description,omitempty"`
	Tags        []client.Tag           `json:"tags"`
	Interfaces  []client.HostInterface `json:"interfaces"`
	Macros      []client.Macro         `json:"macros,omitempty"`
	Inventory   interface{}            `json:"inventory,omitempty"`
}

// GetHosts creates a tool for listing Zabbix hosts
func GetHosts(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_hosts",
			mcp.WithToolAnnotation(
				mcp.ToolAnnotation{
					IdempotentHint: utils.ToBoolPtr(true),
				},
			),
			mcp.WithDescription("List hosts from the Zabbix server. Can filter by host IDs, group IDs, or search term."),
			mcp.WithString("hostids",
				mcp.Description("Comma-separated list of host IDs to filter by"),
			),
			mcp.WithString("groupids",
				mcp.Description("Comma-separated list of host group IDs to filter by"),
			),
			mcp.WithString("search",
				mcp.Description("Search hosts by name (partial match)"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of hosts to return (default: 100)"),
			),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getHostsHandler(ctx, req, logger)
		},
	}
}

func getHostsHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	logger.Debug("Handling get_hosts request")

	// Get Zabbix client from context
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to get Zabbix client")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	// Build parameters
	params := client.HostGetParams{
		Output:           "extend",
		SelectGroups:     "extend",
		SelectTemplates:  "extend",
		SelectTags:       "extend",
		SelectInterfaces: "extend",
		SelectMacros:     "extend",
		SelectInventory:  "extend",
	}

	// Parse arguments with type assertion
	args, ok := req.Params.Arguments.(map[string]interface{})
	if ok && args != nil {
		if hostids, ok := args["hostids"].(string); ok && hostids != "" {
			params.HostIDs = splitAndTrim(hostids)
		}

		if groupids, ok := args["groupids"].(string); ok && groupids != "" {
			params.GroupIDs = splitAndTrim(groupids)
		}

		if search, ok := args["search"].(string); ok && search != "" {
			params.Search = map[string]string{"name": search}
		}

		if limit, ok := args["limit"].(float64); ok && limit > 0 {
			params.Limit = int(limit)
		} else {
			params.Limit = 100
		}
	} else {
		params.Limit = 100
	}

	// Make API call
	result, err := zabbix.Call("host.get", params)
	if err != nil {
		logger.WithError(err).Error("Failed to get hosts")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get hosts: %v", err)), nil
	}

	var hosts []Host
	if err := json.Unmarshal(result, &hosts); err != nil {
		logger.WithError(err).Error("Failed to parse hosts response")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse hosts: %v", err)), nil
	}

	jsonData, err := json.MarshalIndent(hosts, "", "  ")
	if err != nil {
		logger.WithError(err).Error("Failed to marshal hosts to JSON")
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling JSON: %v", err)), nil
	}

	logger.WithField("host_count", len(hosts)).Debug("Successfully listed hosts")
	return mcp.NewToolResultText(string(jsonData)), nil
}

// Helper function to split comma-separated strings
func splitAndTrim(s string) []string {
	var result []string
	for _, part := range splitString(s, ",") {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
