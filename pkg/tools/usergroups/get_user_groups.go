// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package usergroups

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
	"github.com/vfcastr/Zabbix-MCP/pkg/utils"
)

type UserGroup struct {
	UserGroupID string `json:"usrgrpid"`
	Name        string `json:"name"`
	GuiAccess   string `json:"gui_access"`
	UsersStatus string `json:"users_status"`
	DebugMode   string `json:"debug_mode"`
}

type UserGroupGetParams struct {
	Output       interface{} `json:"output,omitempty"`
	UserGroupIDs []string    `json:"usrgrpids,omitempty"`
	UserIDs      []string    `json:"userids,omitempty"`
	Search       interface{} `json:"search,omitempty"`
	SelectUsers  interface{} `json:"selectUsers,omitempty"`
	Limit        int         `json:"limit,omitempty"`
}

func GetUserGroups(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_user_groups",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("Retrieve user groups from Zabbix."),
			mcp.WithString("usrgrpids", mcp.Description("Comma-separated list of user group IDs to filter by")),
			mcp.WithString("userids", mcp.Description("Comma-separated list of user IDs to filter by")),
			mcp.WithString("search", mcp.Description("Search user groups by name")),
			mcp.WithNumber("limit", mcp.Description("Max user groups to return (default: 100)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getUserGroupsHandler(ctx, req, logger)
		},
	}
}

func getUserGroupsHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	params := UserGroupGetParams{
		Output:      "extend",
		SelectUsers: "extend",
		Limit:       100,
	}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		if v, ok := args["usrgrpids"].(string); ok && v != "" {
			params.UserGroupIDs = splitAndTrim(v)
		}
		if v, ok := args["userids"].(string); ok && v != "" {
			params.UserIDs = splitAndTrim(v)
		}
		if v, ok := args["search"].(string); ok && v != "" {
			params.Search = map[string]string{"name": v}
		}
		if v, ok := args["limit"].(float64); ok && v > 0 {
			params.Limit = int(v)
		}
	}

	result, err := zabbix.Call("usergroup.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get user groups: %v", err)), nil
	}

	var groups []UserGroup
	json.Unmarshal(result, &groups)
	jsonData, _ := json.MarshalIndent(groups, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}

func splitAndTrim(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
