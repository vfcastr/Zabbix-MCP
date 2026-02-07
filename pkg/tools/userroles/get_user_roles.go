// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package userroles

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

type UserRole struct {
	RoleID   string `json:"roleid"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Readonly string `json:"readonly"`
}

type UserRoleGetParams struct {
	Output      interface{} `json:"output,omitempty"`
	RoleIDs     []string    `json:"roleids,omitempty"`
	Search      interface{} `json:"search,omitempty"`
	SelectRules interface{} `json:"selectRules,omitempty"`
	SelectUsers interface{} `json:"selectUsers,omitempty"`
	Limit       int         `json:"limit,omitempty"`
}

func GetUserRoles(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_user_roles",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("Retrieve user roles from Zabbix."),
			mcp.WithString("roleids", mcp.Description("Comma-separated list of role IDs to filter by")),
			mcp.WithString("search", mcp.Description("Search roles by name")),
			mcp.WithNumber("limit", mcp.Description("Max roles to return (default: 100)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getUserRolesHandler(ctx, req, logger)
		},
	}
}

func getUserRolesHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	params := UserRoleGetParams{
		Output:      "extend",
		SelectRules: "extend",
		Limit:       100,
	}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		if v, ok := args["roleids"].(string); ok && v != "" {
			params.RoleIDs = splitAndTrim(v)
		}
		if v, ok := args["search"].(string); ok && v != "" {
			params.Search = map[string]string{"name": v}
		}
		if v, ok := args["limit"].(float64); ok && v > 0 {
			params.Limit = int(v)
		}
	}

	result, err := zabbix.Call("role.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get user roles: %v", err)), nil
	}

	var roles []UserRole
	json.Unmarshal(result, &roles)
	jsonData, _ := json.MarshalIndent(roles, "", "  ")
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
