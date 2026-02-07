// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package lld

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

type LLDRuleCreateParams struct {
	HostID      string `json:"hostid"`
	Name        string `json:"name"`
	Key         string `json:"key_"`
	Type        int    `json:"type"`
	Delay       string `json:"delay,omitempty"`
	Lifetime    string `json:"lifetime,omitempty"`
	Description string `json:"description,omitempty"`
}

func CreateLLDRule(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_lld_rule",
			mcp.WithDescription("Create a new low-level discovery rule in Zabbix."),
			mcp.WithString("hostid", mcp.Description("Host ID to create the LLD rule for"), mcp.Required()),
			mcp.WithString("name", mcp.Description("Name of the LLD rule"), mcp.Required()),
			mcp.WithString("key_", mcp.Description("Item key for the LLD rule"), mcp.Required()),
			mcp.WithNumber("type", mcp.Description("LLD rule type: 0=Zabbix agent, 2=Zabbix trapper, 3=Simple check, 5=Internal, 7=Zabbix agent (active), etc."), mcp.Required()),
			mcp.WithString("delay", mcp.Description("Update interval (e.g., '1h', '30m')")),
			mcp.WithString("lifetime", mcp.Description("Time period after which items discovered will be deleted (e.g., '30d')")),
			mcp.WithString("description", mcp.Description("Description of the LLD rule")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return createLLDRuleHandler(ctx, req, logger)
		},
	}
}

func createLLDRuleHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	hostid, _ := args["hostid"].(string)
	name, _ := args["name"].(string)
	key, _ := args["key_"].(string)
	itemType, _ := args["type"].(float64)

	if hostid == "" || name == "" || key == "" {
		return mcp.NewToolResultError("hostid, name, and key_ are required"), nil
	}

	params := LLDRuleCreateParams{
		HostID: hostid,
		Name:   name,
		Key:    key,
		Type:   int(itemType),
	}

	if v, ok := args["delay"].(string); ok && v != "" {
		params.Delay = v
	}
	if v, ok := args["lifetime"].(string); ok && v != "" {
		params.Lifetime = v
	}
	if v, ok := args["description"].(string); ok && v != "" {
		params.Description = v
	}

	result, err := zabbix.Call("discoveryrule.create", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create LLD rule: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "LLD rule created successfully",
		"name":     name,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
