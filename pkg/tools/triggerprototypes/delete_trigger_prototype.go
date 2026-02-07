// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package triggerprototypes

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

func DeleteTriggerPrototype(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_trigger_prototype",
			mcp.WithDescription("Delete trigger prototypes from Zabbix."),
			mcp.WithString("triggerids", mcp.Description("Comma-separated list of trigger prototype IDs to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteTriggerPrototypeHandler(ctx, req, logger)
		},
	}
}

func deleteTriggerPrototypeHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	triggeridsStr, ok := args["triggerids"].(string)
	if !ok || triggeridsStr == "" {
		return mcp.NewToolResultError("triggerids is required"), nil
	}

	var triggerids []string
	for _, id := range strings.Split(triggeridsStr, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			triggerids = append(triggerids, trimmed)
		}
	}

	result, err := zabbix.Call("triggerprototype.delete", triggerids)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete trigger prototypes: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":    "Trigger prototypes deleted successfully",
		"triggerids": triggerids,
		"response":   response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
