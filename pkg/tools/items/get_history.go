// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package items

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
	"github.com/vfcastr/Zabbix-MCP/pkg/utils"
)

// HistoryEntry represents a single history record from Zabbix
type HistoryEntry struct {
	ItemID string `json:"itemid"`
	Clock  string `json:"clock"`
	Value  string `json:"value"`
	NS     string `json:"ns,omitempty"`
}

// HistoryGetParams represents parameters for history.get API call
type HistoryGetParams struct {
	Output    interface{} `json:"output"`
	ItemIDs   []string    `json:"itemids,omitempty"`
	HostIDs   []string    `json:"hostids,omitempty"`
	History   int         `json:"history"` // 0=float, 1=char, 2=log, 3=unsigned, 4=text
	TimeFrom  int64       `json:"time_from,omitempty"`
	TimeTill  int64       `json:"time_till,omitempty"`
	SortField []string    `json:"sortfield,omitempty"`
	SortOrder []string    `json:"sortorder,omitempty"`
	Limit     int         `json:"limit,omitempty"`
}

// GetHistory creates a tool to retrieve historical values from Zabbix
func GetHistory(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_history",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("Get historical values for monitoring items. Returns the most recent values for CPU, memory, or any monitored metric."),
			mcp.WithString("itemids", mcp.Required(), mcp.Description("Comma-separated list of item IDs to get history for")),
			mcp.WithNumber("history_type", mcp.Description("Value type: 0=float (default), 1=char, 2=log, 3=unsigned int, 4=text")),
			mcp.WithNumber("time_from", mcp.Description("Unix timestamp - start of the time range (default: 1 hour ago)")),
			mcp.WithNumber("time_till", mcp.Description("Unix timestamp - end of the time range (default: now)")),
			mcp.WithNumber("limit", mcp.Description("Max records to return (default: 10, max: 1000)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getHistoryHandler(ctx, req, logger)
		},
	}
}

func getHistoryHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	// Default parameters
	now := time.Now().Unix()
	params := HistoryGetParams{
		Output:    "extend",
		History:   0,          // float by default (for CPU %, memory, etc.)
		TimeFrom:  now - 3600, // Last hour
		TimeTill:  now,        // Until now
		SortField: []string{"clock"},
		SortOrder: []string{"DESC"},
		Limit:     10,
	}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		// Required: itemids
		if v, ok := args["itemids"].(string); ok && v != "" {
			params.ItemIDs = splitAndTrim(v)
		} else {
			return mcp.NewToolResultError("itemids is required"), nil
		}

		// Optional: history_type
		if v, ok := args["history_type"].(float64); ok {
			params.History = int(v)
		}

		// Optional: time_from
		if v, ok := args["time_from"].(float64); ok && v > 0 {
			params.TimeFrom = int64(v)
		}

		// Optional: time_till
		if v, ok := args["time_till"].(float64); ok && v > 0 {
			params.TimeTill = int64(v)
		}

		// Optional: limit
		if v, ok := args["limit"].(float64); ok && v > 0 {
			if v > 1000 {
				v = 1000
			}
			params.Limit = int(v)
		}
	}

	logger.WithFields(log.Fields{
		"itemids":   params.ItemIDs,
		"time_from": params.TimeFrom,
		"time_till": params.TimeTill,
		"limit":     params.Limit,
	}).Debug("Getting history")

	result, err := zabbix.Call("history.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get history: %v", err)), nil
	}

	var history []HistoryEntry
	json.Unmarshal(result, &history)

	// Format output with human-readable timestamps
	type FormattedEntry struct {
		ItemID    string `json:"itemid"`
		Timestamp string `json:"timestamp"`
		Value     string `json:"value"`
	}

	formatted := make([]FormattedEntry, len(history))
	for i, h := range history {
		var ts int64
		fmt.Sscanf(h.Clock, "%d", &ts)
		formatted[i] = FormattedEntry{
			ItemID:    h.ItemID,
			Timestamp: time.Unix(ts, 0).Format("2006-01-02 15:04:05"),
			Value:     h.Value,
		}
	}

	jsonData, _ := json.MarshalIndent(formatted, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
