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

// CreateHostGroup returns the tool definition and handler for creating a new Zabbix host group
func CreateHostGroup(logger *logrus.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_create_host_group",
			mcp.WithDescription("Create a new host group in Zabbix."),
			mcp.WithString("name",
				mcp.Description("Name of the host group"),
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
			name, ok := args["name"].(string)
			if !ok || name == "" {
				return mcp.NewToolResultError("name is required"), nil
			}

			params := client.HostGroupCreateParams{
				Name: name,
			}

			// Make API call
			result, err := zabbixClient.Call("hostgroup.create", params)
			if err != nil {
				logger.WithError(err).Error("Failed to create host group")
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(string(result)), nil
		},
	}
}
