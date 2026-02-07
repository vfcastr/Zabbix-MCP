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

// GetTemplateGroups returns the tool definition and handler for retrieving Zabbix template groups
func GetTemplateGroups(logger *logrus.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_get_template_groups",
			mcp.WithDescription("List template groups from Zabbix server. Can filter by group IDs, template IDs, or search term."),
			mcp.WithString("groupids",
				mcp.Description("Comma-separated list of template group IDs to filter by"),
			),
			mcp.WithString("templateids",
				mcp.Description("Comma-separated list of template IDs to filter by"),
			),
			mcp.WithString("search",
				mcp.Description("Search template groups by name"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of groups to return (default: 100)"),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			zabbixClient, err := client.GetZabbixClientFromContext(ctx, logger)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			args, ok := request.Params.Arguments.(map[string]interface{})
			if !ok {
				args = make(map[string]interface{})
			}

			// Parse parameters
			params := client.TemplateGroupGetParams{
				Output:          "extend",
				SelectTemplates: []string{"templateid", "name"},
				Limit:           100,
			}

			if groupIDs, ok := args["groupids"].(string); ok && groupIDs != "" {
				params.GroupIDs = strings.Split(groupIDs, ",")
			}

			if templateIDs, ok := args["templateids"].(string); ok && templateIDs != "" {
				params.TemplateIDs = strings.Split(templateIDs, ",")
			}

			if search, ok := args["search"].(string); ok && search != "" {
				params.Search = map[string]string{"name": search}
			}

			if limit, ok := args["limit"].(float64); ok {
				params.Limit = int(limit)
			}

			// Make API call (Zabbix 7.0 uses templategroup.get)
			result, err := zabbixClient.Call("templategroup.get", params)
			if err != nil {
				logger.WithError(err).Error("Failed to get template groups")
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(string(result)), nil
		},
	}
}
