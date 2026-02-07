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

type UserGroupUpdateParams struct {
	UserGroupID string `json:"usrgrpid"`
	Name        string `json:"name,omitempty"`
	GuiAccess   *int   `json:"gui_access,omitempty"`
	UsersStatus *int   `json:"users_status,omitempty"`
	DebugMode   *int   `json:"debug_mode,omitempty"`
}

func UpdateUserGroup(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("update_user_group",
			mcp.WithDescription("Update an existing user group in Zabbix."),
			mcp.WithString("usrgrpid", mcp.Description("User group ID to update"), mcp.Required()),
			mcp.WithString("name", mcp.Description("New name of the user group")),
			mcp.WithNumber("gui_access", mcp.Description("GUI access: 0=system default, 1=internal auth, 2=LDAP, 3=disabled")),
			mcp.WithNumber("users_status", mcp.Description("User status: 0=enabled, 1=disabled")),
			mcp.WithNumber("debug_mode", mcp.Description("Debug mode: 0=disabled, 1=enabled")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateUserGroupHandler(ctx, req, logger)
		},
	}
}

func updateUserGroupHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	usrgrpid, ok := args["usrgrpid"].(string)
	if !ok || usrgrpid == "" {
		return mcp.NewToolResultError("usrgrpid is required"), nil
	}

	params := UserGroupUpdateParams{UserGroupID: usrgrpid}

	if v, ok := args["name"].(string); ok && v != "" {
		params.Name = v
	}
	if v, ok := args["gui_access"].(float64); ok {
		val := int(v)
		params.GuiAccess = &val
	}
	if v, ok := args["users_status"].(float64); ok {
		val := int(v)
		params.UsersStatus = &val
	}
	if v, ok := args["debug_mode"].(float64); ok {
		val := int(v)
		params.DebugMode = &val
	}

	result, err := zabbix.Call("usergroup.update", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update user group: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "User group updated successfully",
		"usrgrpid": usrgrpid,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
