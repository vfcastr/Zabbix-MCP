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

// DeleteProxyGroups creates a tool for deleting Zabbix proxy groups
func DeleteProxyGroups(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_delete_proxy_groups",
			mcp.WithDescription("Delete one or more proxy groups from the Zabbix server."),
			mcp.WithString("proxy_groupids",
				mcp.Description("Comma-separated list of proxy group IDs to delete"),
				mcp.Required(),
			),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteProxyGroupsHandler(ctx, req, logger)
		},
	}
}

func deleteProxyGroupsHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	logger.Debug("Handling zabbix_delete_proxy_groups request")

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

	paramsStr, ok := args["proxy_groupids"].(string)
	if !ok || paramsStr == "" {
		return mcp.NewToolResultError("proxy_groupids is required"), nil
	}

	groupIDs := utils.SplitAndTrim(paramsStr)

	result, err := zabbix.Call("proxygroup.delete", groupIDs)
	if err != nil {
		logger.WithError(err).Error("Failed to delete proxy groups")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete proxy groups: %v", err)), nil
	}

	var response struct {
		ProxyGroupIDs []string `json:"proxy_groupids"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		logger.WithError(err).Error("Failed to parse delete proxy groups response")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":        "Proxy groups deleted successfully",
		"proxy_groupids": response.ProxyGroupIDs,
	}, "", "  ")

	logger.WithField("group_count", len(response.ProxyGroupIDs)).Info("Successfully deleted proxy groups")
	return mcp.NewToolResultText(string(jsonData)), nil
}
