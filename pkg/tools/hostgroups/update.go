// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package hostgroups

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

// UpdateHostGroup returns the tool definition and handler for updating a Zabbix host group
func UpdateHostGroup(logger *logrus.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_update_host_group",
			mcp.WithDescription("Update an existing host group in Zabbix."),
			mcp.WithString("groupid",
				mcp.Description("ID of the host group to update"),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("New name of the host group"),
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
			groupID, ok := args["groupid"].(string)
			if !ok || groupID == "" {
				return mcp.NewToolResultError("groupid is required"), nil
			}

			name, ok := args["name"].(string)
			if !ok || name == "" {
				return mcp.NewToolResultError("name is required"), nil
			}

			params := client.HostGroupUpdateParams{
				GroupID: groupID,
				Name:    name,
			}

			// Make API call
			result, err := zabbixClient.Call("hostgroup.update", params)
			if err != nil {
				logger.WithError(err).Error("Failed to update host group")
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(string(result)), nil
		},
	}
}
