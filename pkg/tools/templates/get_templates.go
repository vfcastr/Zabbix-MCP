// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package templates

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

type Template struct {
	TemplateID  string       `json:"templateid"`
	Host        string       `json:"host"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Tags        []client.Tag `json:"tags"`
}

func GetTemplates(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_templates",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("List templates from Zabbix."),
			mcp.WithString("templateids", mcp.Description("Comma-separated template IDs")),
			mcp.WithString("hostids", mcp.Description("Comma-separated host IDs")),
			mcp.WithString("search", mcp.Description("Search by name")),
			mcp.WithNumber("limit", mcp.Description("Max templates (default: 100)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getTemplatesHandler(ctx, req, logger)
		},
	}
}

func getTemplatesHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	params := client.TemplateGetParams{Output: "extend", SelectHosts: "extend", SelectTags: "extend", Limit: 100}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		if v, ok := args["templateids"].(string); ok && v != "" {
			params.TemplateIDs = splitAndTrim(v)
		}
		if v, ok := args["hostids"].(string); ok && v != "" {
			params.HostIDs = splitAndTrim(v)
		}
		if v, ok := args["search"].(string); ok && v != "" {
			params.Search = map[string]string{"name": v}
		}
		if v, ok := args["limit"].(float64); ok && v > 0 {
			params.Limit = int(v)
		}
	}

	result, err := zabbix.Call("template.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	var templates []Template
	json.Unmarshal(result, &templates)
	jsonData, _ := json.MarshalIndent(templates, "", "  ")
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
