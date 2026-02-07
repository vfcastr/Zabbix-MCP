// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

type EventAcknowledgeParams struct {
	EventIDs      []string `json:"eventids"`
	Action        int      `json:"action,omitempty"`
	Message       string   `json:"message,omitempty"`
	Severity      *int     `json:"severity,omitempty"`
	SuppressUntil int64    `json:"suppress_until,omitempty"`
	CauseEventID  string   `json:"cause_eventid,omitempty"`
}

func AcknowledgeEvent(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("acknowledge_event",
			mcp.WithDescription("Acknowledge events or update them (add message, change severity, close, suppress, etc.)."),
			mcp.WithString("eventids", mcp.Description("Comma-separated list of event IDs to acknowledge"), mcp.Required()),
			mcp.WithNumber("action", mcp.Description("Action bitmask: 1=close, 2=acknowledge, 4=add message, 8=change severity, 16=unacknowledge, 32=suppress, 64=unsuppress, 128=change rank, 256=change symptoms to cause")),
			mcp.WithString("message", mcp.Description("Message to add to the event")),
			mcp.WithNumber("severity", mcp.Description("New severity (0-5) when action includes change severity (8)")),
			mcp.WithNumber("suppress_until", mcp.Description("Unix timestamp until which to suppress the event")),
			mcp.WithString("cause_eventid", mcp.Description("Cause event ID when changing symptom to cause")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return acknowledgeEventHandler(ctx, req, logger)
		},
	}
}

func acknowledgeEventHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	eventidsStr, ok := args["eventids"].(string)
	if !ok || eventidsStr == "" {
		return mcp.NewToolResultError("eventids is required"), nil
	}

	params := EventAcknowledgeParams{
		EventIDs: splitEventIDs(eventidsStr),
		Action:   2, // Default: acknowledge
	}

	if v, ok := args["action"].(float64); ok {
		params.Action = int(v)
	}
	if v, ok := args["message"].(string); ok && v != "" {
		params.Message = v
	}
	if v, ok := args["severity"].(float64); ok {
		sev := int(v)
		params.Severity = &sev
	}
	if v, ok := args["suppress_until"].(float64); ok && v > 0 {
		params.SuppressUntil = int64(v)
	}
	if v, ok := args["cause_eventid"].(string); ok && v != "" {
		params.CauseEventID = v
	}

	result, err := zabbix.Call("event.acknowledge", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to acknowledge events: %v", err)), nil
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "Events acknowledged successfully",
		"eventids": params.EventIDs,
		"response": response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}

func splitEventIDs(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
