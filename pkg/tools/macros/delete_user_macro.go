// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package macros

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

func DeleteUserMacro(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_user_macro",
			mcp.WithDescription("Delete host-level user macros from Zabbix."),
			mcp.WithString("hostmacroids", mcp.Description("Comma-separated list of host macro IDs to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteUserMacroHandler(ctx, req, logger)
		},
	}
}

func deleteUserMacroHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	hostmacroidsStr, ok := args["hostmacroids"].(string)
	if !ok || hostmacroidsStr == "" {
		return mcp.NewToolResultError("hostmacroids is required"), nil
	}

	var hostmacroids []string
	for _, id := range strings.Split(hostmacroidsStr, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			hostmacroids = append(hostmacroids, trimmed)
		}
	}

	result, err := zabbix.Call("usermacro.delete", hostmacroids)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete user macros: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":      "User macros deleted successfully",
		"hostmacroids": hostmacroids,
		"response":     response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
