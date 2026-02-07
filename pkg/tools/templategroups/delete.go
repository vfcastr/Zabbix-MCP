// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package templategroups

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

// DeleteTemplateGroup returns the tool definition and handler for deleting Zabbix template groups
func DeleteTemplateGroup(logger *logrus.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_delete_template_group",
			mcp.WithDescription("Delete template groups from Zabbix server."),
			mcp.WithString("groupids",
				mcp.Description("Comma-separated list of template group IDs to delete"),
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

			// Make API call (Zabbix 7.0 uses templategroup.delete)
			result, err := zabbixClient.Call("templategroup.delete", groupIDs)
			if err != nil {
				logger.WithError(err).Error("Failed to delete template groups")
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(string(result)), nil
		},
	}
}
