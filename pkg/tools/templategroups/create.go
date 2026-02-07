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

// CreateTemplateGroup returns the tool definition and handler for creating a new Zabbix template group
func CreateTemplateGroup(logger *logrus.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_create_template_group",
			mcp.WithDescription("Create a new template group in Zabbix."),
			mcp.WithString("name",
				mcp.Description("Name of the template group"),
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

			params := client.TemplateGroupCreateParams{
				Name: name,
			}

			// Make API call (Zabbix 7.0 uses templategroup.create)
			result, err := zabbixClient.Call("templategroup.create", params)
			if err != nil {
				logger.WithError(err).Error("Failed to create template group")
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(string(result)), nil
		},
	}
}
