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
)

func UpdateMaintenance(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("update_maintenance",
			mcp.WithDescription("Update a maintenance period."),
			mcp.WithString("maintenanceid", mcp.Required(), mcp.Description("Maintenance ID")),
			mcp.WithString("name", mcp.Description("New name")),
			mcp.WithString("active_till", mcp.Description("New end time")),
			mcp.WithString("description", mcp.Description("New description")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateMaintenanceHandler(ctx, req, logger)
		},
	}
}

func updateMaintenanceHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing arguments"), nil
	}

	maintenanceid, _ := args["maintenanceid"].(string)
	if maintenanceid == "" {
		return mcp.NewToolResultError("maintenanceid required"), nil
	}

	params := client.MaintenanceUpdateParams{MaintenanceID: maintenanceid}
	if v, ok := args["name"].(string); ok && v != "" {
		params.Name = v
	}
	if v, ok := args["active_till"].(string); ok && v != "" {
		if ts := parseTime(v); ts > 0 {
			params.ActiveTill = ts
		}
	}
	if v, ok := args["description"].(string); ok {
		params.Description = v
	}

	result, err := zabbix.Call("maintenance.update", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	var response struct {
		MaintenanceIDs []string `json:"maintenanceids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Maintenance updated", "maintenanceids": response.MaintenanceIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
