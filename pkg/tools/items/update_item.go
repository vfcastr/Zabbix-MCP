// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package items

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

func UpdateItem(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("update_item",
			mcp.WithDescription("Update an existing item in Zabbix."),
			mcp.WithString("itemid", mcp.Required(), mcp.Description("Item ID")),
			mcp.WithString("name", mcp.Description("New item name")),
			mcp.WithNumber("status", mcp.Description("Status: 0=enabled, 1=disabled")),
			mcp.WithString("delay", mcp.Description("New update interval")),
			mcp.WithString("tags", mcp.Description("Tags in format key:value,key2:value2")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateItemHandler(ctx, req, logger)
		},
	}
}

func updateItemHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing or invalid arguments"), nil
	}

	itemid, _ := args["itemid"].(string)
	if itemid == "" {
		return mcp.NewToolResultError("itemid is required"), nil
	}

	params := client.ItemUpdateParams{ItemID: itemid}
	if v, ok := args["name"].(string); ok && v != "" {
		params.Name = v
	}
	if v, ok := args["status"].(float64); ok {
		s := int(v)
		params.Status = &s
	}
	if v, ok := args["delay"].(string); ok && v != "" {
		params.Delay = v
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

	result, err := zabbix.Call("item.update", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update item: %v", err)), nil
	}

	var response struct {
		ItemIDs []string `json:"itemids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Item updated", "itemids": response.ItemIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
