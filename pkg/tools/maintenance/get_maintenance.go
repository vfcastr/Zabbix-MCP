// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package maintenance

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

type Maintenance struct {
	MaintenanceID   string `json:"maintenanceid"`
	Name            string `json:"name"`
	ActiveSince     string `json:"active_since"`
	ActiveTill      string `json:"active_till"`
	Description     string `json:"description,omitempty"`
	MaintenanceType string `json:"maintenance_type"`
}

func GetMaintenance(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_maintenance",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("List maintenance periods from Zabbix."),
			mcp.WithString("maintenanceids", mcp.Description("Comma-separated maintenance IDs")),
			mcp.WithString("hostids", mcp.Description("Comma-separated host IDs")),
			mcp.WithNumber("limit", mcp.Description("Max records (default: 100)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getMaintenanceHandler(ctx, req, logger)
		},
	}
}

func getMaintenanceHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	params := client.MaintenanceGetParams{Output: "extend", SelectHosts: "extend", SelectTimeperiods: "extend", Limit: 100}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		if v, ok := args["maintenanceids"].(string); ok && v != "" {
			params.MaintenanceIDs = splitAndTrim(v)
		}
		if v, ok := args["hostids"].(string); ok && v != "" {
			params.HostIDs = splitAndTrim(v)
		}
		if v, ok := args["limit"].(float64); ok && v > 0 {
			params.Limit = int(v)
		}
	}

	result, err := zabbix.Call("maintenance.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	var items []Maintenance
	json.Unmarshal(result, &items)
	jsonData, _ := json.MarshalIndent(items, "", "  ")
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
