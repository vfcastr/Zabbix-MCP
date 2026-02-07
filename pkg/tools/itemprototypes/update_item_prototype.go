// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package itemprototypes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

type ItemPrototypeUpdateParams struct {
	ItemID      string `json:"itemid"`
	Name        string `json:"name,omitempty"`
	Key         string `json:"key_,omitempty"`
	Delay       string `json:"delay,omitempty"`
	Units       string `json:"units,omitempty"`
	Description string `json:"description,omitempty"`
	Status      *int   `json:"status,omitempty"`
}

func UpdateItemPrototype(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("update_item_prototype",
			mcp.WithDescription("Update an existing item prototype in Zabbix."),
			mcp.WithString("itemid", mcp.Description("Item prototype ID to update"), mcp.Required()),
			mcp.WithString("name", mcp.Description("New name")),
			mcp.WithString("key_", mcp.Description("New item key")),
			mcp.WithString("delay", mcp.Description("New update interval")),
			mcp.WithString("units", mcp.Description("New value units")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithNumber("status", mcp.Description("Status: 0=enabled, 1=disabled")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateItemPrototypeHandler(ctx, req, logger)
		},
	}
}

func updateItemPrototypeHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	itemid, ok := args["itemid"].(string)
	if !ok || itemid == "" {
		return mcp.NewToolResultError("itemid is required"), nil
	}

	params := ItemPrototypeUpdateParams{ItemID: itemid}

	if v, ok := args["name"].(string); ok && v != "" {
		params.Name = v
	}
	if v, ok := args["key_"].(string); ok && v != "" {
		params.Key = v
	}
	if v, ok := args["delay"].(string); ok && v != "" {
		params.Delay = v
	}
	if v, ok := args["units"].(string); ok && v != "" {
		params.Units = v
	}
	if v, ok := args["description"].(string); ok && v != "" {
		params.Description = v
	}
	if v, ok := args["status"].(float64); ok {
		s := int(v)
		params.Status = &s
	}

	result, err := zabbix.Call("itemprototype.update", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update item prototype: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "Item prototype updated successfully",
		"itemid":   itemid,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
