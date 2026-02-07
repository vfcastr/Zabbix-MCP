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
	"github.com/vfcastr/Zabbix-MCP/pkg/utils"
)

type ItemPrototype struct {
	ItemID      string `json:"itemid"`
	HostID      string `json:"hostid"`
	RuleID      string `json:"ruleid"`
	Name        string `json:"name"`
	Key         string `json:"key_"`
	Type        string `json:"type"`
	ValueType   string `json:"value_type"`
	Delay       string `json:"delay"`
	Status      string `json:"status"`
	Units       string `json:"units"`
	Description string `json:"description"`
}

type ItemPrototypeGetParams struct {
	Output              interface{} `json:"output,omitempty"`
	ItemIDs             []string    `json:"itemids,omitempty"`
	HostIDs             []string    `json:"hostids,omitempty"`
	DiscoveryIDs        []string    `json:"discoveryids,omitempty"`
	TemplateIDs         []string    `json:"templateids,omitempty"`
	Search              interface{} `json:"search,omitempty"`
	SelectDiscoveryRule interface{} `json:"selectDiscoveryRule,omitempty"`
	SelectPreprocessing interface{} `json:"selectPreprocessing,omitempty"`
	SortField           []string    `json:"sortfield,omitempty"`
	Limit               int         `json:"limit,omitempty"`
}

func GetItemPrototypes(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_item_prototypes",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("Retrieve item prototypes from Zabbix."),
			mcp.WithString("itemids", mcp.Description("Comma-separated list of item prototype IDs to filter by")),
			mcp.WithString("hostids", mcp.Description("Comma-separated list of host IDs to filter by")),
			mcp.WithString("discoveryids", mcp.Description("Comma-separated list of LLD rule IDs to filter by")),
			mcp.WithString("templateids", mcp.Description("Comma-separated list of template IDs to filter by")),
			mcp.WithString("search", mcp.Description("Search item prototypes by name or key")),
			mcp.WithNumber("limit", mcp.Description("Max item prototypes to return (default: 100)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getItemPrototypesHandler(ctx, req, logger)
		},
	}
}

func getItemPrototypesHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	params := ItemPrototypeGetParams{
		Output:              "extend",
		SelectDiscoveryRule: "extend",
		Limit:               100,
	}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		if v, ok := args["itemids"].(string); ok && v != "" {
			params.ItemIDs = splitAndTrim(v)
		}
		if v, ok := args["hostids"].(string); ok && v != "" {
			params.HostIDs = splitAndTrim(v)
		}
		if v, ok := args["discoveryids"].(string); ok && v != "" {
			params.DiscoveryIDs = splitAndTrim(v)
		}
		if v, ok := args["templateids"].(string); ok && v != "" {
			params.TemplateIDs = splitAndTrim(v)
		}
		if v, ok := args["search"].(string); ok && v != "" {
			params.Search = map[string]string{"name": v, "key_": v}
		}
		if v, ok := args["limit"].(float64); ok && v > 0 {
			params.Limit = int(v)
		}
	}

	result, err := zabbix.Call("itemprototype.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get item prototypes: %v", err)), nil
	}

	var items []ItemPrototype
	json.Unmarshal(result, &items)
	jsonData, _ := json.MarshalIndent(items, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}

func splitAndTrim(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
