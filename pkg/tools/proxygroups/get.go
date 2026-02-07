// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package proxygroups

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

// GetProxyGroups creates a tool for listing Zabbix proxy groups
func GetProxyGroups(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_get_proxy_groups",
			mcp.WithToolAnnotation(
				mcp.ToolAnnotation{
					IdempotentHint: utils.ToBoolPtr(true),
				},
			),
			mcp.WithDescription("Retrieve all configured proxy groups. Can filter by proxy group IDs or search term."),
			mcp.WithString("proxy_groupids",
				mcp.Description("Comma-separated list of proxy group IDs to filter by"),
			),
			mcp.WithString("search",
				mcp.Description("Search proxy groups by name (partial match)"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of proxy groups to return (default: 100)"),
			),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getProxyGroupsHandler(ctx, req, logger)
		},
	}
}

func getProxyGroupsHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	logger.Debug("Handling zabbix_get_proxy_groups request")

	// Get Zabbix client from context
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to get Zabbix client")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	// Build parameters
	params := client.ProxyGroupGetParams{
		Output:        "extend",
		SelectProxies: "extend",
	}

	// Parse arguments with type assertion
	args, ok := req.Params.Arguments.(map[string]interface{})
	if ok && args != nil {
		if groupids, ok := args["proxy_groupids"].(string); ok && groupids != "" {
			params.ProxyGroupIDs = utils.SplitAndTrim(groupids)
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
	result, err := zabbix.Call("proxygroup.get", params)
	if err != nil {
		logger.WithError(err).Error("Failed to get proxy groups")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get proxy groups: %v", err)), nil
	}

	var groups []client.ProxyGroup
	if err := json.Unmarshal(result, &groups); err != nil {
		logger.WithError(err).Error("Failed to parse proxy groups response")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse proxy groups: %v", err)), nil
	}

	jsonData, err := json.MarshalIndent(groups, "", "  ")
	if err != nil {
		logger.WithError(err).Error("Failed to marshal proxy groups to JSON")
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling JSON: %v", err)), nil
	}

	logger.WithField("group_count", len(groups)).Debug("Successfully listed proxy groups")
	return mcp.NewToolResultText(string(jsonData)), nil
}
