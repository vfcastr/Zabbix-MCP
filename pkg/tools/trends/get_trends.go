// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package trends

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

type Trend struct {
	ItemID   string `json:"itemid"`
	Clock    string `json:"clock"`
	Num      string `json:"num"`
	ValueMin string `json:"value_min"`
	ValueAvg string `json:"value_avg"`
	ValueMax string `json:"value_max"`
}

type TrendGetParams struct {
	Output   interface{} `json:"output,omitempty"`
	ItemIDs  []string    `json:"itemids,omitempty"`
	TimeFrom int64       `json:"time_from,omitempty"`
	TimeTill int64       `json:"time_till,omitempty"`
	Limit    int         `json:"limit,omitempty"`
}

func GetTrends(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_trends",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("Retrieve trend values calculated by Zabbix server for presentation or further processing. Trends are hourly aggregated data (min, avg, max)."),
			mcp.WithString("itemids", mcp.Description("Comma-separated list of item IDs to get trends for"), mcp.Required()),
			mcp.WithNumber("time_from", mcp.Description("Return only trends after this Unix timestamp")),
			mcp.WithNumber("time_till", mcp.Description("Return only trends before this Unix timestamp")),
			mcp.WithNumber("limit", mcp.Description("Max trend records to return (default: 100)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getTrendsHandler(ctx, req, logger)
		},
	}
}

func getTrendsHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	itemidsStr, ok := args["itemids"].(string)
	if !ok || itemidsStr == "" {
		return mcp.NewToolResultError("itemids is required"), nil
	}

	params := TrendGetParams{
		Output:  "extend",
		ItemIDs: splitAndTrim(itemidsStr),
		Limit:   100,
	}

	if v, ok := args["time_from"].(float64); ok && v > 0 {
		params.TimeFrom = int64(v)
	}
	if v, ok := args["time_till"].(float64); ok && v > 0 {
		params.TimeTill = int64(v)
	}
	if v, ok := args["limit"].(float64); ok && v > 0 {
		params.Limit = int(v)
	}

	result, err := zabbix.Call("trend.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get trends: %v", err)), nil
	}

	var trends []Trend
	json.Unmarshal(result, &trends)
	jsonData, _ := json.MarshalIndent(trends, "", "  ")
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
