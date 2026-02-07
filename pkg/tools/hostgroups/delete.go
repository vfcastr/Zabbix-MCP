// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package hostgroups

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

// DeleteHostGroup returns the tool definition and handler for deleting Zabbix host groups
func DeleteHostGroup(logger *logrus.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_delete_host_group",
			mcp.WithDescription("Delete host groups from Zabbix server."),
			mcp.WithString("groupids",
				mcp.Description("Comma-separated list of host group IDs to delete"),
				mcp.Required(),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			zabbixClient, err := client.GetZabbixClientFromContext(ctx, logger)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			args, ok := request.Params.Arguments.(map[string]interface{})
			if !ok {
				return mcp.NewToolResultError("invalid arguments"), nil
			}

			// Parse parameters
			groupIDsStr, ok := args["groupids"].(string)
			if !ok || groupIDsStr == "" {
				return mcp.NewToolResultError("groupids is required"), nil
			}

			groupIDs := strings.Split(groupIDsStr, ",")

			// Make API call
			result, err := zabbixClient.Call("hostgroup.delete", groupIDs)
			if err != nil {
				logger.WithError(err).Error("Failed to delete host groups")
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(string(result)), nil
		},
	}
}
