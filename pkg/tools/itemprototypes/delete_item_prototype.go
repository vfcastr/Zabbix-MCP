// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package itemprototypes

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

func DeleteItemPrototype(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_item_prototype",
			mcp.WithDescription("Delete item prototypes from Zabbix."),
			mcp.WithString("itemids", mcp.Description("Comma-separated list of item prototype IDs to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteItemPrototypeHandler(ctx, req, logger)
		},
	}
}

func deleteItemPrototypeHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	itemidsStr, ok := args["itemids"].(string)
	if !ok || itemidsStr == "" {
		return mcp.NewToolResultError("itemids is required"), nil
	}

	var itemids []string
	for _, id := range strings.Split(itemidsStr, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			itemids = append(itemids, trimmed)
		}
	}

	result, err := zabbix.Call("itemprototype.delete", itemids)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete item prototypes: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "Item prototypes deleted successfully",
		"itemids":  itemids,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
