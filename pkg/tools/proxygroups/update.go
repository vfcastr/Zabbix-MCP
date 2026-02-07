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

// UpdateProxyGroup creates a tool for updating a Zabbix proxy group
func UpdateProxyGroup(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_update_proxy_group",
			mcp.WithDescription("Update an existing proxy group in Zabbix."),
			mcp.WithString("proxy_groupid",
				mcp.Description("ID of the proxy group to update"),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Name of the proxy group"),
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
			return updateProxyGroupHandler(ctx, req, logger)
		},
	}
}

func updateProxyGroupHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	logger.Debug("Handling zabbix_update_proxy_group request")

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

	proxyGroupID, ok := args["proxy_groupid"].(string)
	if !ok || proxyGroupID == "" {
		return mcp.NewToolResultError("proxy_groupid is required"), nil
	}

	params := client.ProxyGroupUpdateParams{
		ProxyGroupID: proxyGroupID,
	}

	if v, ok := args["name"].(string); ok {
		params.Name = v
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

	result, err := zabbix.Call("proxygroup.update", params)
	if err != nil {
		logger.WithError(err).Error("Failed to update proxy group")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update proxy group: %v", err)), nil
	}

	var response struct {
		ProxyGroupIDs []string `json:"proxy_groupids"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		logger.WithError(err).Error("Failed to parse update proxy group response")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":        "Proxy group updated successfully",
		"proxy_groupids": response.ProxyGroupIDs,
	}, "", "  ")

	logger.WithField("proxy_groupid", proxyGroupID).Info("Successfully updated proxy group")
	return mcp.NewToolResultText(string(jsonData)), nil
}
