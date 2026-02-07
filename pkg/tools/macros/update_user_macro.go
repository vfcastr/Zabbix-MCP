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

type UserMacroUpdateParams struct {
	HostMacroID string `json:"hostmacroid"`
	Macro       string `json:"macro,omitempty"`
	Value       string `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
	Type        *int   `json:"type,omitempty"`
}

func UpdateUserMacro(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("update_user_macro",
			mcp.WithDescription("Update an existing host-level user macro in Zabbix."),
			mcp.WithString("hostmacroid", mcp.Description("Host macro ID to update"), mcp.Required()),
			mcp.WithString("macro", mcp.Description("New macro name")),
			mcp.WithString("value", mcp.Description("New macro value")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithNumber("type", mcp.Description("Macro type: 0=text, 1=secret, 2=vault secret")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateUserMacroHandler(ctx, req, logger)
		},
	}
}

func updateUserMacroHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	hostmacroid, ok := args["hostmacroid"].(string)
	if !ok || hostmacroid == "" {
		return mcp.NewToolResultError("hostmacroid is required"), nil
	}

	params := UserMacroUpdateParams{HostMacroID: hostmacroid}

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

	result, err := zabbix.Call("usermacro.update", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update user macro: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":     "User macro updated successfully",
		"hostmacroid": hostmacroid,
		"response":    response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
