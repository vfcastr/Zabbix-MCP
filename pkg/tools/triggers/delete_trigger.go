// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package triggers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

func DeleteTrigger(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_trigger",
			mcp.WithDescription("Delete triggers from Zabbix."),
			mcp.WithString("triggerids", mcp.Required(), mcp.Description("Comma-separated trigger IDs")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteTriggerHandler(ctx, req, logger)
		},
	}
}

func deleteTriggerHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing arguments"), nil
	}

	triggeridsStr, _ := args["triggerids"].(string)
	if triggeridsStr == "" {
		return mcp.NewToolResultError("triggerids required"), nil
	}

	result, err := zabbix.Call("trigger.delete", splitAndTrim(triggeridsStr))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	var response struct {
		TriggerIDs []string `json:"triggerids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Triggers deleted", "triggerids": response.TriggerIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
