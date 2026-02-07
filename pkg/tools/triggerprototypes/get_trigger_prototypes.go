// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package triggerprototypes

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

type TriggerPrototype struct {
	TriggerID   string `json:"triggerid"`
	Description string `json:"description"`
	Expression  string `json:"expression"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	Comments    string `json:"comments"`
}

type TriggerPrototypeGetParams struct {
	Output              interface{} `json:"output,omitempty"`
	TriggerIDs          []string    `json:"triggerids,omitempty"`
	HostIDs             []string    `json:"hostids,omitempty"`
	DiscoveryIDs        []string    `json:"discoveryids,omitempty"`
	TemplateIDs         []string    `json:"templateids,omitempty"`
	Search              interface{} `json:"search,omitempty"`
	SelectDiscoveryRule interface{} `json:"selectDiscoveryRule,omitempty"`
	SelectFunctions     interface{} `json:"selectFunctions,omitempty"`
	SelectItems         interface{} `json:"selectItems,omitempty"`
	SelectTags          interface{} `json:"selectTags,omitempty"`
	MinSeverity         *int        `json:"min_severity,omitempty"`
	SortField           []string    `json:"sortfield,omitempty"`
	Limit               int         `json:"limit,omitempty"`
}

func GetTriggerPrototypes(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_trigger_prototypes",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("Retrieve trigger prototypes from Zabbix."),
			mcp.WithString("triggerids", mcp.Description("Comma-separated list of trigger prototype IDs to filter by")),
			mcp.WithString("hostids", mcp.Description("Comma-separated list of host IDs to filter by")),
			mcp.WithString("discoveryids", mcp.Description("Comma-separated list of LLD rule IDs to filter by")),
			mcp.WithString("templateids", mcp.Description("Comma-separated list of template IDs to filter by")),
			mcp.WithString("search", mcp.Description("Search trigger prototypes by description")),
			mcp.WithNumber("min_severity", mcp.Description("Minimum severity (0-5)")),
			mcp.WithNumber("limit", mcp.Description("Max trigger prototypes to return (default: 100)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getTriggerPrototypesHandler(ctx, req, logger)
		},
	}
}

func getTriggerPrototypesHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	params := TriggerPrototypeGetParams{
		Output:              "extend",
		SelectDiscoveryRule: "extend",
		SelectTags:          "extend",
		Limit:               100,
	}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		if v, ok := args["triggerids"].(string); ok && v != "" {
			params.TriggerIDs = splitAndTrim(v)
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
			params.Search = map[string]string{"description": v}
		}
		if v, ok := args["min_severity"].(float64); ok {
			sev := int(v)
			params.MinSeverity = &sev
		}
		if v, ok := args["limit"].(float64); ok && v > 0 {
			params.Limit = int(v)
		}
	}

	result, err := zabbix.Call("triggerprototype.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get trigger prototypes: %v", err)), nil
	}

	var triggers []TriggerPrototype
	json.Unmarshal(result, &triggers)
	jsonData, _ := json.MarshalIndent(triggers, "", "  ")
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
