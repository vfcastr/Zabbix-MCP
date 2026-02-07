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

func DeleteMaintenance(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_maintenance",
			mcp.WithDescription("Delete maintenance periods."),
			mcp.WithString("maintenanceids", mcp.Required(), mcp.Description("Comma-separated maintenance IDs")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteMaintenanceHandler(ctx, req, logger)
		},
	}
}

func deleteMaintenanceHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing arguments"), nil
	}

	maintenanceidsStr, _ := args["maintenanceids"].(string)
	if maintenanceidsStr == "" {
		return mcp.NewToolResultError("maintenanceids required"), nil
	}

	result, err := zabbix.Call("maintenance.delete", splitAndTrim(maintenanceidsStr))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	var response struct {
		MaintenanceIDs []string `json:"maintenanceids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Maintenance deleted", "maintenanceids": response.MaintenanceIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
