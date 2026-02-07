// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package maintenance

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

func CreateMaintenance(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_maintenance",
			mcp.WithDescription("Create a maintenance period."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Maintenance name")),
			mcp.WithString("active_since", mcp.Required(), mcp.Description("Start time (Unix timestamp or RFC3339)")),
			mcp.WithString("active_till", mcp.Required(), mcp.Description("End time (Unix timestamp or RFC3339)")),
			mcp.WithString("hostids", mcp.Description("Comma-separated host IDs")),
			mcp.WithString("groupids", mcp.Description("Comma-separated group IDs")),
			mcp.WithNumber("period", mcp.Description("Duration in seconds (default: 3600)")),
			mcp.WithString("description", mcp.Description("Description")),
			mcp.WithNumber("maintenance_type", mcp.Description("Type: 0=with data, 1=without")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return createMaintenanceHandler(ctx, req, logger)
		},
	}
}

func createMaintenanceHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing arguments"), nil
	}

	name, _ := args["name"].(string)
	activeSinceStr, _ := args["active_since"].(string)
	activeTillStr, _ := args["active_till"].(string)
	if name == "" || activeSinceStr == "" || activeTillStr == "" {
		return mcp.NewToolResultError("name, active_since, and active_till are required"), nil
	}

	activeSince := parseTime(activeSinceStr)
	activeTill := parseTime(activeTillStr)
	if activeSince == 0 || activeTill == 0 {
		return mcp.NewToolResultError("Invalid time format"), nil
	}

	period := 3600
	if v, ok := args["period"].(float64); ok && v > 0 {
		period = int(v)
	}

	params := client.MaintenanceCreateParams{
		Name:        name,
		ActiveSince: activeSince,
		ActiveTill:  activeTill,
		Timeperiods: []client.MaintenanceTimeperiod{{TimeperiodType: 0, StartDate: activeSince, Period: period}},
	}

	if v, ok := args["hostids"].(string); ok && v != "" {
		params.HostIDs = splitAndTrim(v)
	}
	if v, ok := args["groupids"].(string); ok && v != "" {
		params.GroupIDs = splitAndTrim(v)
	}
	if v, ok := args["description"].(string); ok {
		params.Description = v
	}
	if v, ok := args["maintenance_type"].(float64); ok {
		params.MaintenanceType = int(v)
	}

	result, err := zabbix.Call("maintenance.create", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	var response struct {
		MaintenanceIDs []string `json:"maintenanceids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Maintenance created", "maintenanceids": response.MaintenanceIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}

func parseTime(s string) int64 {
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		return ts
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Unix()
	}
	return 0
}
