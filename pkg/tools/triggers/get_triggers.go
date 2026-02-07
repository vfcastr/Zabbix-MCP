// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package triggers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
	"github.com/vfcastr/Zabbix-MCP/pkg/utils"
)

type Trigger struct {
	TriggerID   string       `json:"triggerid"`
	Description string       `json:"description"`
	Expression  string       `json:"expression"`
	Priority    string       `json:"priority"`
	Status      string       `json:"status"`
	Tags        []client.Tag `json:"tags"`
}

func GetTriggers(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_triggers",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("List triggers from Zabbix."),
			mcp.WithString("triggerids", mcp.Description("Comma-separated trigger IDs")),
			mcp.WithString("hostids", mcp.Description("Comma-separated host IDs")),
			mcp.WithNumber("min_severity", mcp.Description("Minimum severity (0-5)")),
			mcp.WithNumber("limit", mcp.Description("Max triggers (default: 100)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getTriggersHandler(ctx, req, logger)
		},
	}
}

func getTriggersHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	params := client.TriggerGetParams{Output: "extend", SelectHosts: "extend", SelectTags: "extend", Limit: 100}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		if v, ok := args["triggerids"].(string); ok && v != "" {
			params.TriggerIDs = splitAndTrim(v)
		}
		if v, ok := args["hostids"].(string); ok && v != "" {
			params.HostIDs = splitAndTrim(v)
		}
		if v, ok := args["min_severity"].(float64); ok {
			params.MinSeverity = int(v)
		}
		if v, ok := args["limit"].(float64); ok && v > 0 {
			params.Limit = int(v)
		}
	}

	result, err := zabbix.Call("trigger.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	var triggers []Trigger
	json.Unmarshal(result, &triggers)
	jsonData, _ := json.MarshalIndent(triggers, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}

func splitAndTrim(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			if t := trim(s[start:i]); t != "" {
				result = append(result, t)
			}
			start = i + 1
		}
	}
	if t := trim(s[start:]); t != "" {
		result = append(result, t)
	}
	return result
}

func trim(s string) string {
	start, end := 0, len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}
