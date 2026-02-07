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

type UserRoleCreateParams struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

func CreateUserRole(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_user_role",
			mcp.WithDescription("Create a new user role in Zabbix."),
			mcp.WithString("name", mcp.Description("Name of the role"), mcp.Required()),
			mcp.WithNumber("type", mcp.Description("Role type: 1=User, 2=Admin, 3=Super admin"), mcp.Required()),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return createUserRoleHandler(ctx, req, logger)
		},
	}
}

func createUserRoleHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
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

	roleType, ok := args["type"].(float64)
	if !ok {
		return mcp.NewToolResultError("type is required"), nil
	}

	params := UserRoleCreateParams{
		Name: name,
		Type: int(roleType),
	}

	result, err := zabbix.Call("role.create", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create user role: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "User role created successfully",
		"name":     name,
		"type":     int(roleType),
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
