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
)

// DeleteHost creates a tool for deleting hosts from Zabbix
func DeleteHost(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_host",
			mcp.WithDescription("Delete one or more hosts from the Zabbix server."),
			mcp.WithString("hostids",
				mcp.Required(),
				mcp.Description("Comma-separated list of host IDs to delete"),
			),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteHostHandler(ctx, req, logger)
		},
	}
}

func deleteHostHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	logger.Debug("Handling delete_host request")

	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to get Zabbix client")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing or invalid arguments"), nil
	}

	hostidsStr, ok := args["hostids"].(string)
	if !ok || hostidsStr == "" {
		return mcp.NewToolResultError("hostids parameter is required"), nil
	}

	hostids := splitAndTrim(hostidsStr)
	if len(hostids) == 0 {
		return mcp.NewToolResultError("At least one hostid is required"), nil
	}

	result, err := zabbix.Call("host.delete", hostids)
	if err != nil {
		logger.WithError(err).Error("Failed to delete hosts")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete hosts: %v", err)), nil
	}

	var response struct {
		HostIDs []string `json:"hostids"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		logger.WithError(err).Error("Failed to parse delete response")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message": "Hosts deleted successfully",
		"hostids": response.HostIDs,
	}, "", "  ")

	logger.WithField("hostids", response.HostIDs).Info("Successfully deleted hosts")
	return mcp.NewToolResultText(string(jsonData)), nil
}
