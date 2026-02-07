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

type ItemPrototypeCreateParams struct {
	RuleID      string `json:"ruleid"`
	HostID      string `json:"hostid"`
	Name        string `json:"name"`
	Key         string `json:"key_"`
	Type        int    `json:"type"`
	ValueType   int    `json:"value_type"`
	Delay       string `json:"delay,omitempty"`
	Units       string `json:"units,omitempty"`
	Description string `json:"description,omitempty"`
}

func CreateItemPrototype(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_item_prototype",
			mcp.WithDescription("Create a new item prototype in Zabbix."),
			mcp.WithString("ruleid", mcp.Description("ID of the LLD rule this prototype belongs to"), mcp.Required()),
			mcp.WithString("hostid", mcp.Description("Host ID to create the item prototype for"), mcp.Required()),
			mcp.WithString("name", mcp.Description("Name of the item prototype"), mcp.Required()),
			mcp.WithString("key_", mcp.Description("Item key for the prototype"), mcp.Required()),
			mcp.WithNumber("type", mcp.Description("Item type: 0=Zabbix agent, 2=Zabbix trapper, 3=Simple check, 5=Internal, 7=Zabbix agent (active), etc."), mcp.Required()),
			mcp.WithNumber("value_type", mcp.Description("Value type: 0=float, 1=char, 2=log, 3=unsigned, 4=text"), mcp.Required()),
			mcp.WithString("delay", mcp.Description("Update interval")),
			mcp.WithString("units", mcp.Description("Value units")),
			mcp.WithString("description", mcp.Description("Description")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return createItemPrototypeHandler(ctx, req, logger)
		},
	}
}

func createItemPrototypeHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	ruleid, _ := args["ruleid"].(string)
	hostid, _ := args["hostid"].(string)
	name, _ := args["name"].(string)
	key, _ := args["key_"].(string)
	itemType, _ := args["type"].(float64)
	valueType, _ := args["value_type"].(float64)

	if ruleid == "" || hostid == "" || name == "" || key == "" {
		return mcp.NewToolResultError("ruleid, hostid, name, and key_ are required"), nil
	}

	params := ItemPrototypeCreateParams{
		RuleID:    ruleid,
		HostID:    hostid,
		Name:      name,
		Key:       key,
		Type:      int(itemType),
		ValueType: int(valueType),
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

	result, err := zabbix.Call("itemprototype.create", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create item prototype: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "Item prototype created successfully",
		"name":     name,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
