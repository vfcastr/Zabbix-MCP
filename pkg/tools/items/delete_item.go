// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package items

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

func DeleteItem(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_item",
			mcp.WithDescription("Delete items from Zabbix."),
			mcp.WithString("itemids", mcp.Required(), mcp.Description("Comma-separated item IDs")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteItemHandler(ctx, req, logger)
		},
	}
}

func deleteItemHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing or invalid arguments"), nil
	}

	itemidsStr, _ := args["itemids"].(string)
	if itemidsStr == "" {
		return mcp.NewToolResultError("itemids is required"), nil
	}

	result, err := zabbix.Call("item.delete", splitAndTrim(itemidsStr))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete items: %v", err)), nil
	}

	var response struct {
		ItemIDs []string `json:"itemids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Items deleted", "itemids": response.ItemIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
