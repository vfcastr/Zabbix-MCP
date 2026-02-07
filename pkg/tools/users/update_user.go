// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package users

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

type UserUpdateParams struct {
	UserID     string              `json:"userid"`
	Username   string              `json:"username,omitempty"`
	Name       string              `json:"name,omitempty"`
	Surname    string              `json:"surname,omitempty"`
	Passwd     string              `json:"passwd,omitempty"`
	RoleID     string              `json:"roleid,omitempty"`
	UserGroups []map[string]string `json:"usrgrps,omitempty"`
}

func UpdateUser(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("update_user",
			mcp.WithDescription("Update an existing user in Zabbix."),
			mcp.WithString("userid", mcp.Description("User ID to update"), mcp.Required()),
			mcp.WithString("username", mcp.Description("New username")),
			mcp.WithString("name", mcp.Description("New first name")),
			mcp.WithString("surname", mcp.Description("New last name")),
			mcp.WithString("passwd", mcp.Description("New password")),
			mcp.WithString("roleid", mcp.Description("New role ID")),
			mcp.WithString("usrgrpids", mcp.Description("Comma-separated list of user group IDs")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateUserHandler(ctx, req, logger)
		},
	}
}

func updateUserHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	userid, ok := args["userid"].(string)
	if !ok || userid == "" {
		return mcp.NewToolResultError("userid is required"), nil
	}

	params := UserUpdateParams{UserID: userid}

	if v, ok := args["username"].(string); ok && v != "" {
		params.Username = v
	}
	if v, ok := args["name"].(string); ok && v != "" {
		params.Name = v
	}
	if v, ok := args["surname"].(string); ok && v != "" {
		params.Surname = v
	}
	if v, ok := args["passwd"].(string); ok && v != "" {
		params.Passwd = v
	}
	if v, ok := args["roleid"].(string); ok && v != "" {
		params.RoleID = v
	}
	if v, ok := args["usrgrpids"].(string); ok && v != "" {
		groups := []map[string]string{}
		for _, id := range strings.Split(v, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				groups = append(groups, map[string]string{"usrgrpid": trimmed})
			}
		}
		params.UserGroups = groups
	}

	result, err := zabbix.Call("user.update", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update user: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "User updated successfully",
		"userid":   userid,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
