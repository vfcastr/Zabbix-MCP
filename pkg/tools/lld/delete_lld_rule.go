// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package lld

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

func DeleteLLDRule(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_lld_rule",
			mcp.WithDescription("Delete low-level discovery rules from Zabbix."),
			mcp.WithString("itemids", mcp.Description("Comma-separated list of LLD rule IDs to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteLLDRuleHandler(ctx, req, logger)
		},
	}
}

func deleteLLDRuleHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	itemidsStr, ok := args["itemids"].(string)
	if !ok || itemidsStr == "" {
		return mcp.NewToolResultError("itemids is required"), nil
	}

	var itemids []string
	for _, id := range strings.Split(itemidsStr, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			itemids = append(itemids, trimmed)
		}
	}

	result, err := zabbix.Call("discoveryrule.delete", itemids)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete LLD rules: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "LLD rules deleted successfully",
		"itemids":  itemids,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
