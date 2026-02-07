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

type UserCreateParams struct {
	Username   string              `json:"username"`
	Name       string              `json:"name,omitempty"`
	Surname    string              `json:"surname,omitempty"`
	Passwd     string              `json:"passwd"`
	RoleID     string              `json:"roleid"`
	UserGroups []map[string]string `json:"usrgrps"`
}

func CreateUser(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_user",
			mcp.WithDescription("Create a new user in Zabbix."),
			mcp.WithString("username", mcp.Description("Username for login"), mcp.Required()),
			mcp.WithString("passwd", mcp.Description("User password"), mcp.Required()),
			mcp.WithString("roleid", mcp.Description("Role ID to assign to the user"), mcp.Required()),
			mcp.WithString("usrgrpids", mcp.Description("Comma-separated list of user group IDs to add the user to"), mcp.Required()),
			mcp.WithString("name", mcp.Description("First name of the user")),
			mcp.WithString("surname", mcp.Description("Last name of the user")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return createUserHandler(ctx, req, logger)
		},
	}
}

func createUserHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	username, _ := args["username"].(string)
	passwd, _ := args["passwd"].(string)
	roleid, _ := args["roleid"].(string)
	usrgrpidsStr, _ := args["usrgrpids"].(string)

	if username == "" || passwd == "" || roleid == "" || usrgrpidsStr == "" {
		return mcp.NewToolResultError("username, passwd, roleid, and usrgrpids are required"), nil
	}

	userGroups := []map[string]string{}
	for _, id := range strings.Split(usrgrpidsStr, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			userGroups = append(userGroups, map[string]string{"usrgrpid": trimmed})
		}
	}

	params := UserCreateParams{
		Username:   username,
		Passwd:     passwd,
		RoleID:     roleid,
		UserGroups: userGroups,
	}

	if v, ok := args["name"].(string); ok && v != "" {
		params.Name = v
	}
	if v, ok := args["surname"].(string); ok && v != "" {
		params.Surname = v
	}

	result, err := zabbix.Call("user.create", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create user: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "User created successfully",
		"username": username,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
