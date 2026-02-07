// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package macros

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

type UserMacro struct {
	HostMacroID   string `json:"hostmacroid,omitempty"`
	GlobalMacroID string `json:"globalmacroid,omitempty"`
	HostID        string `json:"hostid,omitempty"`
	Macro         string `json:"macro"`
	Value         string `json:"value"`
	Description   string `json:"description,omitempty"`
	Type          string `json:"type"`
}

type UserMacroGetParams struct {
	Output         interface{} `json:"output,omitempty"`
	GlobalMacro    bool        `json:"globalmacro,omitempty"`
	GlobalMacroIDs []string    `json:"globalmacroids,omitempty"`
	GroupIDs       []string    `json:"groupids,omitempty"`
	HostIDs        []string    `json:"hostids,omitempty"`
	HostMacroIDs   []string    `json:"hostmacroids,omitempty"`
	TemplateIDs    []string    `json:"templateids,omitempty"`
	Search         interface{} `json:"search,omitempty"`
	SelectHosts    interface{} `json:"selectHosts,omitempty"`
	SortField      []string    `json:"sortfield,omitempty"`
	Limit          int         `json:"limit,omitempty"`
}

func GetUserMacros(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_user_macros",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("Retrieve host-level user macros from Zabbix."),
			mcp.WithString("hostids", mcp.Description("Comma-separated list of host IDs to filter by")),
			mcp.WithString("hostmacroids", mcp.Description("Comma-separated list of host macro IDs to filter by")),
			mcp.WithString("groupids", mcp.Description("Comma-separated list of host group IDs to filter by")),
			mcp.WithString("templateids", mcp.Description("Comma-separated list of template IDs to filter by")),
			mcp.WithString("search", mcp.Description("Search macros by name")),
			mcp.WithNumber("limit", mcp.Description("Max macros to return (default: 100)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getUserMacrosHandler(ctx, req, logger)
		},
	}
}

func getUserMacrosHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	params := UserMacroGetParams{
		Output:      "extend",
		SelectHosts: "extend",
		Limit:       100,
	}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		if v, ok := args["hostids"].(string); ok && v != "" {
			params.HostIDs = splitAndTrim(v)
		}
		if v, ok := args["hostmacroids"].(string); ok && v != "" {
			params.HostMacroIDs = splitAndTrim(v)
		}
		if v, ok := args["groupids"].(string); ok && v != "" {
			params.GroupIDs = splitAndTrim(v)
		}
		if v, ok := args["templateids"].(string); ok && v != "" {
			params.TemplateIDs = splitAndTrim(v)
		}
		if v, ok := args["search"].(string); ok && v != "" {
			params.Search = map[string]string{"macro": v}
		}
		if v, ok := args["limit"].(float64); ok && v > 0 {
			params.Limit = int(v)
		}
	}

	result, err := zabbix.Call("usermacro.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get user macros: %v", err)), nil
	}

	var macros []UserMacro
	json.Unmarshal(result, &macros)
	jsonData, _ := json.MarshalIndent(macros, "", "  ")
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
