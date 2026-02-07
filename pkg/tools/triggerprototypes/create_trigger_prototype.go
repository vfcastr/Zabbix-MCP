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

type TriggerPrototypeCreateParams struct {
	Description        string              `json:"description"`
	Expression         string              `json:"expression"`
	RecoveryMode       int                 `json:"recovery_mode,omitempty"`
	RecoveryExpression string              `json:"recovery_expression,omitempty"`
	Priority           int                 `json:"priority,omitempty"`
	Comments           string              `json:"comments,omitempty"`
	Tags               []map[string]string `json:"tags,omitempty"`
}

func CreateTriggerPrototype(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_trigger_prototype",
			mcp.WithDescription("Create a new trigger prototype in Zabbix."),
			mcp.WithString("description", mcp.Description("Trigger prototype name/description"), mcp.Required()),
			mcp.WithString("expression", mcp.Description("Trigger expression"), mcp.Required()),
			mcp.WithNumber("priority", mcp.Description("Trigger severity: 0=not classified, 1=information, 2=warning, 3=average, 4=high, 5=disaster")),
			mcp.WithString("recovery_expression", mcp.Description("Recovery expression")),
			mcp.WithNumber("recovery_mode", mcp.Description("Recovery mode: 0=expression, 1=recovery expression, 2=none")),
			mcp.WithString("comments", mcp.Description("Comments/description")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return createTriggerPrototypeHandler(ctx, req, logger)
		},
	}
}

func createTriggerPrototypeHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	description, _ := args["description"].(string)
	expression, _ := args["expression"].(string)

	if description == "" || expression == "" {
		return mcp.NewToolResultError("description and expression are required"), nil
	}

	params := TriggerPrototypeCreateParams{
		Description: description,
		Expression:  expression,
	}

	if v, ok := args["priority"].(float64); ok {
		params.Priority = int(v)
	}
	if v, ok := args["recovery_expression"].(string); ok && v != "" {
		params.RecoveryExpression = v
	}
	if v, ok := args["recovery_mode"].(float64); ok {
		params.RecoveryMode = int(v)
	}
	if v, ok := args["comments"].(string); ok && v != "" {
		params.Comments = v
	}

	result, err := zabbix.Call("triggerprototype.create", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create trigger prototype: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":     "Trigger prototype created successfully",
		"description": description,
		"response":    response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
