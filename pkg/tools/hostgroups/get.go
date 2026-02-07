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

// GetHostGroups returns the tool definition and handler for retrieving Zabbix host groups
func GetHostGroups(logger *logrus.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_get_host_groups",
			mcp.WithDescription("List host groups from Zabbix server. Can filter by group IDs, host IDs, or search term."),
			mcp.WithString("groupids",
				mcp.Description("Comma-separated list of host group IDs to filter by"),
			),
			mcp.WithString("hostids",
				mcp.Description("Comma-separated list of host IDs to filter by"),
			),
			mcp.WithString("search",
				mcp.Description("Search host groups by name"),
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
			params := client.HostGroupGetParams{
				Output:      "extend",
				SelectHosts: []string{"hostid", "name"},
				Limit:       100,
			}

			if groupIDs, ok := args["groupids"].(string); ok && groupIDs != "" {
				params.GroupIDs = strings.Split(groupIDs, ",")
			}

			if hostIDs, ok := args["hostids"].(string); ok && hostIDs != "" {
				params.HostIDs = strings.Split(hostIDs, ",")
			}

			if search, ok := args["search"].(string); ok && search != "" {
				params.Search = map[string]string{"name": search}
			}

			if limit, ok := args["limit"].(float64); ok {
				params.Limit = int(limit)
			}

			// Make API call
			result, err := zabbixClient.Call("hostgroup.get", params)
			if err != nil {
				logger.WithError(err).Error("Failed to get host groups")
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(string(result)), nil
		},
	}
}
