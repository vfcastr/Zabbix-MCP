// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package proxies

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

// GetProxies creates a tool for listing Zabbix proxies
func GetProxies(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_get_proxies",
			mcp.WithToolAnnotation(
				mcp.ToolAnnotation{
					IdempotentHint: utils.ToBoolPtr(true),
				},
			),
			mcp.WithDescription("Retrieve all configured proxies. Can filter by proxy IDs, proxy group IDs, or search term."),
			mcp.WithString("proxyids",
				mcp.Description("Comma-separated list of proxy IDs to filter by"),
			),
			mcp.WithString("proxy_groupids",
				mcp.Description("Comma-separated list of proxy group IDs to filter by"),
			),
			mcp.WithString("search",
				mcp.Description("Search proxies by name (partial match)"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of proxies to return (default: 100)"),
			),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getProxiesHandler(ctx, req, logger)
		},
	}
}

func getProxiesHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	logger.Debug("Handling zabbix_get_proxies request")

	// Get Zabbix client from context
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to get Zabbix client")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	// Build parameters
	params := client.ProxyGetParams{
		Output:      "extend",
		SelectHosts: "extend",
	}

	// Parse arguments with type assertion
	args, ok := req.Params.Arguments.(map[string]interface{})
	if ok && args != nil {
		if proxyids, ok := args["proxyids"].(string); ok && proxyids != "" {
			params.ProxyIDs = utils.SplitAndTrim(proxyids)
		}

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
	result, err := zabbix.Call("proxy.get", params)
	if err != nil {
		logger.WithError(err).Error("Failed to get proxies")
		// Zabbix might return an empty array if no proxies found but API call successful
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get proxies: %v", err)), nil
	}

	var proxies []client.Proxy
	if err := json.Unmarshal(result, &proxies); err != nil {
		logger.WithError(err).Error("Failed to parse proxies response")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse proxies: %v", err)), nil
	}

	jsonData, err := json.MarshalIndent(proxies, "", "  ")
	if err != nil {
		logger.WithError(err).Error("Failed to marshal proxies to JSON")
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling JSON: %v", err)), nil
	}

	logger.WithField("proxy_count", len(proxies)).Debug("Successfully listed proxies")
	return mcp.NewToolResultText(string(jsonData)), nil
}
