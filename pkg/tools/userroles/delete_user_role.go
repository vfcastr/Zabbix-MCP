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
)

func DeleteUserRole(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_user_role",
			mcp.WithDescription("Delete user roles from Zabbix."),
			mcp.WithString("roleids", mcp.Description("Comma-separated list of role IDs to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteUserRoleHandler(ctx, req, logger)
		},
	}
}

func deleteUserRoleHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	roleidsStr, ok := args["roleids"].(string)
	if !ok || roleidsStr == "" {
		return mcp.NewToolResultError("roleids is required"), nil
	}

	var roleids []string
	for _, id := range strings.Split(roleidsStr, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			roleids = append(roleids, trimmed)
		}
	}

	result, err := zabbix.Call("role.delete", roleids)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete user roles: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "User roles deleted successfully",
		"roleids":  roleids,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
