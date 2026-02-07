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

type GlobalMacroGetParams struct {
	Output         interface{} `json:"output,omitempty"`
	GlobalMacro    bool        `json:"globalmacro"`
	GlobalMacroIDs []string    `json:"globalmacroids,omitempty"`
	Search         interface{} `json:"search,omitempty"`
	SortField      []string    `json:"sortfield,omitempty"`
	Limit          int         `json:"limit,omitempty"`
}

func GetGlobalMacros(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_global_macros",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("Retrieve global macros from Zabbix."),
			mcp.WithString("globalmacroids", mcp.Description("Comma-separated list of global macro IDs to filter by")),
			mcp.WithString("search", mcp.Description("Search macros by name")),
			mcp.WithNumber("limit", mcp.Description("Max macros to return (default: 100)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getGlobalMacrosHandler(ctx, req, logger)
		},
	}
}

func getGlobalMacrosHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	params := GlobalMacroGetParams{
		Output:      "extend",
		GlobalMacro: true,
		Limit:       100,
	}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		if v, ok := args["globalmacroids"].(string); ok && v != "" {
			params.GlobalMacroIDs = splitGlobalMacroIDs(v)
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
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get global macros: %v", err)), nil
	}

	var macros []UserMacro
	json.Unmarshal(result, &macros)
	jsonData, _ := json.MarshalIndent(macros, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}

func splitGlobalMacroIDs(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
