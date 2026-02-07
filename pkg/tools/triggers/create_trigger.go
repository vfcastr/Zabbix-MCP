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

func CreateTrigger(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_trigger",
			mcp.WithDescription("Create a new trigger in Zabbix."),
			mcp.WithString("description", mcp.Required(), mcp.Description("Trigger name")),
			mcp.WithString("expression", mcp.Required(), mcp.Description("Trigger expression")),
			mcp.WithNumber("priority", mcp.Description("Priority: 0-5 (default: 0)")),
			mcp.WithString("comments", mcp.Description("Comments")),
			mcp.WithString("tags", mcp.Description("Tags in format key:value,key2:value2")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return createTriggerHandler(ctx, req, logger)
		},
	}
}

func createTriggerHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing arguments"), nil
	}

	desc, _ := args["description"].(string)
	expr, _ := args["expression"].(string)
	if desc == "" || expr == "" {
		return mcp.NewToolResultError("description and expression required"), nil
	}

	params := client.TriggerCreateParams{Description: desc, Expression: expr}
	if v, ok := args["priority"].(float64); ok {
		params.Priority = int(v)
	}
	if v, ok := args["comments"].(string); ok {
		params.Comments = v
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

	result, err := zabbix.Call("trigger.create", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	var response struct {
		TriggerIDs []string `json:"triggerids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Trigger created", "triggerids": response.TriggerIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
