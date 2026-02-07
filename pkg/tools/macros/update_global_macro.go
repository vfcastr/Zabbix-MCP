// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package macros

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

type GlobalMacroUpdateParams struct {
	GlobalMacroID string `json:"globalmacroid"`
	Macro         string `json:"macro,omitempty"`
	Value         string `json:"value,omitempty"`
	Description   string `json:"description,omitempty"`
	Type          *int   `json:"type,omitempty"`
}

func UpdateGlobalMacro(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("update_global_macro",
			mcp.WithDescription("Update an existing global macro in Zabbix."),
			mcp.WithString("globalmacroid", mcp.Description("Global macro ID to update"), mcp.Required()),
			mcp.WithString("macro", mcp.Description("New macro name")),
			mcp.WithString("value", mcp.Description("New macro value")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithNumber("type", mcp.Description("Macro type: 0=text, 1=secret, 2=vault secret")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateGlobalMacroHandler(ctx, req, logger)
		},
	}
}

func updateGlobalMacroHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	globalmacroid, ok := args["globalmacroid"].(string)
	if !ok || globalmacroid == "" {
		return mcp.NewToolResultError("globalmacroid is required"), nil
	}

	params := GlobalMacroUpdateParams{GlobalMacroID: globalmacroid}

	if v, ok := args["macro"].(string); ok && v != "" {
		params.Macro = v
	}
	if v, ok := args["value"].(string); ok && v != "" {
		params.Value = v
	}
	if v, ok := args["description"].(string); ok && v != "" {
		params.Description = v
	}
	if v, ok := args["type"].(float64); ok {
		t := int(v)
		params.Type = &t
	}

	result, err := zabbix.Call("usermacro.updateglobal", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update global macro: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":       "Global macro updated successfully",
		"globalmacroid": globalmacroid,
		"response":      response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
