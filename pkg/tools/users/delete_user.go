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

func DeleteUser(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_user",
			mcp.WithDescription("Delete users from Zabbix."),
			mcp.WithString("userids", mcp.Description("Comma-separated list of user IDs to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteUserHandler(ctx, req, logger)
		},
	}
}

func deleteUserHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	useridsStr, ok := args["userids"].(string)
	if !ok || useridsStr == "" {
		return mcp.NewToolResultError("userids is required"), nil
	}

	var userids []string
	for _, id := range strings.Split(useridsStr, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			userids = append(userids, trimmed)
		}
	}

	result, err := zabbix.Call("user.delete", userids)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete users: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "Users deleted successfully",
		"userids":  userids,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
