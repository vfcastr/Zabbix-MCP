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

type GlobalMacroCreateParams struct {
	Macro       string `json:"macro"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Type        int    `json:"type,omitempty"`
}

func CreateGlobalMacro(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_global_macro",
			mcp.WithDescription("Create a new global macro in Zabbix."),
			mcp.WithString("macro", mcp.Description("Macro name (e.g., {$MYMACRO})"), mcp.Required()),
			mcp.WithString("value", mcp.Description("Macro value"), mcp.Required()),
			mcp.WithString("description", mcp.Description("Description of the macro")),
			mcp.WithNumber("type", mcp.Description("Macro type: 0=text (default), 1=secret, 2=vault secret")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return createGlobalMacroHandler(ctx, req, logger)
		},
	}
}

func createGlobalMacroHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	macro, _ := args["macro"].(string)
	value, _ := args["value"].(string)

	if macro == "" || value == "" {
		return mcp.NewToolResultError("macro and value are required"), nil
	}

	params := GlobalMacroCreateParams{
		Macro: macro,
		Value: value,
	}

	if v, ok := args["description"].(string); ok && v != "" {
		params.Description = v
	}
	if v, ok := args["type"].(float64); ok {
		params.Type = int(v)
	}

	result, err := zabbix.Call("usermacro.createglobal", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create global macro: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "Global macro created successfully",
		"macro":    macro,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
