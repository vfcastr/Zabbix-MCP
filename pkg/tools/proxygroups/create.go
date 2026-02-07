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
)

// CreateProxyGroup creates a tool for creating a new Zabbix proxy group
func CreateProxyGroup(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_create_proxy_group",
			mcp.WithDescription("Create a new proxy group in Zabbix."),
			mcp.WithString("name",
				mcp.Description("Name of the proxy group"),
				mcp.Required(),
			),
			mcp.WithString("failover_delay",
				mcp.Description("Failover delay (e.g. 1m)"),
			),
			mcp.WithString("min_online",
				mcp.Description("Minimum number of online proxies"),
			),
			mcp.WithString("description",
				mcp.Description("Description of the proxy group"),
			),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return createProxyGroupHandler(ctx, req, logger)
		},
	}
}

func createProxyGroupHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	logger.Debug("Handling zabbix_create_proxy_group request")

	// Get Zabbix client from context
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to get Zabbix client")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}

	params := client.ProxyGroupCreateParams{
		Name: name,
	}

	if v, ok := args["failover_delay"].(string); ok {
		params.FailoverDelay = v
	}
	if v, ok := args["min_online"].(string); ok {
		params.MinOnline = v
	}
	if v, ok := args["description"].(string); ok {
		params.Description = v
	}

	result, err := zabbix.Call("proxygroup.create", params)
	if err != nil {
		logger.WithError(err).Error("Failed to create proxy group")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create proxy group: %v", err)), nil
	}

	var response struct {
		ProxyGroupIDs []string `json:"proxy_groupids"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		logger.WithError(err).Error("Failed to parse create proxy group response")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":        "Proxy group created successfully",
		"proxy_groupids": response.ProxyGroupIDs,
	}, "", "  ")

	logger.WithField("proxy_groupids", response.ProxyGroupIDs).Info("Successfully created proxy group")
	return mcp.NewToolResultText(string(jsonData)), nil
}
