// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package userroles

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

type UserRoleUpdateParams struct {
	RoleID string `json:"roleid"`
	Name   string `json:"name,omitempty"`
}

func UpdateUserRole(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("update_user_role",
			mcp.WithDescription("Update an existing user role in Zabbix."),
			mcp.WithString("roleid", mcp.Description("Role ID to update"), mcp.Required()),
			mcp.WithString("name", mcp.Description("New name for the role")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateUserRoleHandler(ctx, req, logger)
		},
	}
}

func updateUserRoleHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	roleid, ok := args["roleid"].(string)
	if !ok || roleid == "" {
		return mcp.NewToolResultError("roleid is required"), nil
	}

	params := UserRoleUpdateParams{RoleID: roleid}

	if v, ok := args["name"].(string); ok && v != "" {
		params.Name = v
	}

	result, err := zabbix.Call("role.update", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update user role: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "User role updated successfully",
		"roleid":   roleid,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
