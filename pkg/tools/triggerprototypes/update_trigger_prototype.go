// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package triggerprototypes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

type TriggerPrototypeUpdateParams struct {
	TriggerID          string `json:"triggerid"`
	Description        string `json:"description,omitempty"`
	Expression         string `json:"expression,omitempty"`
	Priority           *int   `json:"priority,omitempty"`
	Comments           string `json:"comments,omitempty"`
	Status             *int   `json:"status,omitempty"`
	RecoveryExpression string `json:"recovery_expression,omitempty"`
}

func UpdateTriggerPrototype(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("update_trigger_prototype",
			mcp.WithDescription("Update an existing trigger prototype in Zabbix."),
			mcp.WithString("triggerid", mcp.Description("Trigger prototype ID to update"), mcp.Required()),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithString("expression", mcp.Description("New expression")),
			mcp.WithNumber("priority", mcp.Description("New severity (0-5)")),
			mcp.WithString("comments", mcp.Description("New comments")),
			mcp.WithNumber("status", mcp.Description("Status: 0=enabled, 1=disabled")),
			mcp.WithString("recovery_expression", mcp.Description("New recovery expression")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateTriggerPrototypeHandler(ctx, req, logger)
		},
	}
}

func updateTriggerPrototypeHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	triggerid, ok := args["triggerid"].(string)
	if !ok || triggerid == "" {
		return mcp.NewToolResultError("triggerid is required"), nil
	}

	params := TriggerPrototypeUpdateParams{TriggerID: triggerid}

	if v, ok := args["description"].(string); ok && v != "" {
		params.Description = v
	}
	if v, ok := args["expression"].(string); ok && v != "" {
		params.Expression = v
	}
	if v, ok := args["priority"].(float64); ok {
		p := int(v)
		params.Priority = &p
	}
	if v, ok := args["comments"].(string); ok && v != "" {
		params.Comments = v
	}
	if v, ok := args["status"].(float64); ok {
		s := int(v)
		params.Status = &s
	}
	if v, ok := args["recovery_expression"].(string); ok && v != "" {
		params.RecoveryExpression = v
	}

	result, err := zabbix.Call("triggerprototype.update", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update trigger prototype: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":   "Trigger prototype updated successfully",
		"triggerid": triggerid,
		"response":  response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
