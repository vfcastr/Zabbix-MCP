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

func CreateItem(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_item",
			mcp.WithDescription("Create a new monitoring item in Zabbix."),
			mcp.WithString("hostid", mcp.Required(), mcp.Description("Host ID")),
			mcp.WithString("interfaceid", mcp.Description("Interface ID (required for Zabbix agent items)")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Item name")),
			mcp.WithString("key_", mcp.Required(), mcp.Description("Item key")),
			mcp.WithNumber("type", mcp.Description("Item type (default: 0=Zabbix agent)")),
			mcp.WithNumber("value_type", mcp.Description("Value type (default: 3=numeric unsigned)")),
			mcp.WithString("delay", mcp.Description("Update interval (default: 1m)")),
			mcp.WithString("tags", mcp.Description("Tags in format key:value,key2:value2")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return createItemHandler(ctx, req, logger)
		},
	}
}

func createItemHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing or invalid arguments"), nil
	}

	hostid, _ := args["hostid"].(string)
	name, _ := args["name"].(string)
	key, _ := args["key_"].(string)
	if hostid == "" || name == "" || key == "" {
		return mcp.NewToolResultError("hostid, name, and key_ are required"), nil
	}

	params := client.ItemCreateParams{HostID: hostid, Name: name, Key: key, Type: 0, ValueType: 3, Delay: "1m"}

	if v, ok := args["interfaceid"].(string); ok && v != "" {
		params.InterfaceID = v
	}

	if v, ok := args["type"].(float64); ok {
		params.Type = int(v)
	}
	if v, ok := args["value_type"].(float64); ok {
		params.ValueType = int(v)
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

	result, err := zabbix.Call("item.create", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create item: %v", err)), nil
	}

	var response struct {
		ItemIDs []string `json:"itemids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Item created", "itemids": response.ItemIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
