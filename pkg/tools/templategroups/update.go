// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package templategroups

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

// UpdateTemplateGroup returns the tool definition and handler for updating a Zabbix template group
func UpdateTemplateGroup(logger *logrus.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_update_template_group",
			mcp.WithDescription("Update an existing template group in Zabbix."),
			mcp.WithString("groupid",
				mcp.Description("ID of the template group to update"),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("New name of the template group"),
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

			params := client.TemplateGroupUpdateParams{
				GroupID: groupID,
				Name:    name,
			}

			// Make API call (Zabbix 7.0 uses templategroup.update)
			result, err := zabbixClient.Call("templategroup.update", params)
			if err != nil {
				logger.WithError(err).Error("Failed to update template group")
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(string(result)), nil
		},
	}
}
