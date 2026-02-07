// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package triggers

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

func UpdateTrigger(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("update_trigger",
			mcp.WithDescription("Update an existing trigger in Zabbix."),
			mcp.WithString("triggerid", mcp.Required(), mcp.Description("Trigger ID")),
			mcp.WithString("description", mcp.Description("New name")),
			mcp.WithNumber("priority", mcp.Description("New priority (0-5)")),
			mcp.WithNumber("status", mcp.Description("Status: 0=enabled, 1=disabled")),
			mcp.WithString("tags", mcp.Description("Tags in format key:value,key2:value2")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateTriggerHandler(ctx, req, logger)
		},
	}
}

func updateTriggerHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing arguments"), nil
	}

	triggerid, _ := args["triggerid"].(string)
	if triggerid == "" {
		return mcp.NewToolResultError("triggerid required"), nil
	}

	params := client.TriggerUpdateParams{TriggerID: triggerid}
	if v, ok := args["description"].(string); ok && v != "" {
		params.Description = v
	}
	if v, ok := args["priority"].(float64); ok {
		p := int(v)
		params.Priority = &p
	}
	if v, ok := args["status"].(float64); ok {
		s := int(v)
		params.Status = &s
	}

	if tagsStr, ok := args["tags"].(string); ok && tagsStr != "" {
		var tags []client.Tag
		for _, tagPart := range strings.Split(tagsStr, ",") {
			parts := strings.SplitN(strings.TrimSpace(tagPart), ":", 2)
			if len(parts) == 2 {
				tags = append(tags, client.Tag{Tag: strings.TrimSpace(parts[0]), Value: strings.TrimSpace(parts[1])})
			} else if len(parts) == 1 {
				tags = append(tags, client.Tag{Tag: strings.TrimSpace(parts[0]), Value: ""})
			}
		}
		params.Tags = tags
	}

	result, err := zabbix.Call("trigger.update", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	var response struct {
		TriggerIDs []string `json:"triggerids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Trigger updated", "triggerids": response.TriggerIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
