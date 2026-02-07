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
)

func DeleteUserGroup(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_user_group",
			mcp.WithDescription("Delete user groups from Zabbix."),
			mcp.WithString("usrgrpids", mcp.Description("Comma-separated list of user group IDs to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteUserGroupHandler(ctx, req, logger)
		},
	}
}

func deleteUserGroupHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	usrgrpidsStr, ok := args["usrgrpids"].(string)
	if !ok || usrgrpidsStr == "" {
		return mcp.NewToolResultError("usrgrpids is required"), nil
	}

	var usrgrpids []string
	for _, id := range strings.Split(usrgrpidsStr, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			usrgrpids = append(usrgrpids, trimmed)
		}
	}

	result, err := zabbix.Call("usergroup.delete", usrgrpids)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete user groups: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":   "User groups deleted successfully",
		"usrgrpids": usrgrpids,
		"response":  response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
