// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package usergroups

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

type UserGroupCreateParams struct {
	Name        string `json:"name"`
	GuiAccess   int    `json:"gui_access,omitempty"`
	UsersStatus int    `json:"users_status,omitempty"`
	DebugMode   int    `json:"debug_mode,omitempty"`
}

func CreateUserGroup(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_user_group",
			mcp.WithDescription("Create a new user group in Zabbix."),
			mcp.WithString("name", mcp.Description("Name of the user group"), mcp.Required()),
			mcp.WithNumber("gui_access", mcp.Description("GUI access: 0=system default, 1=internal auth, 2=LDAP, 3=disabled")),
			mcp.WithNumber("users_status", mcp.Description("User status: 0=enabled, 1=disabled")),
			mcp.WithNumber("debug_mode", mcp.Description("Debug mode: 0=disabled, 1=enabled")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return createUserGroupHandler(ctx, req, logger)
		},
	}
}

func createUserGroupHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}

	params := UserGroupCreateParams{Name: name}

	if v, ok := args["gui_access"].(float64); ok {
		params.GuiAccess = int(v)
	}
	if v, ok := args["users_status"].(float64); ok {
		params.UsersStatus = int(v)
	}
	if v, ok := args["debug_mode"].(float64); ok {
		params.DebugMode = int(v)
	}

	result, err := zabbix.Call("usergroup.create", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create user group: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "User group created successfully",
		"name":     name,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
