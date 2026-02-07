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

func DeleteGlobalMacro(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_global_macro",
			mcp.WithDescription("Delete global macros from Zabbix."),
			mcp.WithString("globalmacroids", mcp.Description("Comma-separated list of global macro IDs to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteGlobalMacroHandler(ctx, req, logger)
		},
	}
}

func deleteGlobalMacroHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	globalmacroidsStr, ok := args["globalmacroids"].(string)
	if !ok || globalmacroidsStr == "" {
		return mcp.NewToolResultError("globalmacroids is required"), nil
	}

	var globalmacroids []string
	for _, id := range strings.Split(globalmacroidsStr, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			globalmacroids = append(globalmacroids, trimmed)
		}
	}

	result, err := zabbix.Call("usermacro.deleteglobal", globalmacroids)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete global macros: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":        "Global macros deleted successfully",
		"globalmacroids": globalmacroids,
		"response":       response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
