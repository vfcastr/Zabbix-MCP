// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package problems

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

type Problem struct {
	EventID      string       `json:"eventid"`
	Source       string       `json:"source"`
	Object       string       `json:"object"`
	ObjectID     string       `json:"objectid"`
	Clock        string       `json:"clock"`
	Name         string       `json:"name"`
	Acknowledged string       `json:"acknowledged"`
	Severity     string       `json:"severity"`
	REventID     string       `json:"r_eventid"`
	RClock       string       `json:"r_clock"`
	Suppressed   string       `json:"suppressed"`
	OpData       string       `json:"opdata"`
	Tags         []client.Tag `json:"tags,omitempty"`
}

type ProblemGetParams struct {
	Output             interface{} `json:"output,omitempty"`
	EventIDs           []string    `json:"eventids,omitempty"`
	GroupIDs           []string    `json:"groupids,omitempty"`
	HostIDs            []string    `json:"hostids,omitempty"`
	ObjectIDs          []string    `json:"objectids,omitempty"`
	Acknowledged       *bool       `json:"acknowledged,omitempty"`
	Suppressed         *bool       `json:"suppressed,omitempty"`
	Severities         []int       `json:"severities,omitempty"`
	Recent             *bool       `json:"recent,omitempty"`
	TimeFrom           int64       `json:"time_from,omitempty"`
	TimeTill           int64       `json:"time_till,omitempty"`
	SelectAcknowledges interface{} `json:"selectAcknowledges,omitempty"`
	SelectTags         interface{} `json:"selectTags,omitempty"`
	SortField          []string    `json:"sortfield,omitempty"`
	SortOrder          []string    `json:"sortorder,omitempty"`
	Limit              int         `json:"limit,omitempty"`
}

func GetProblems(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_problems",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("Retrieve problems according to the given parameters. Problems are sorted by severity and time in descending order by default."),
			mcp.WithString("eventids", mcp.Description("Comma-separated list of event IDs to filter by")),
			mcp.WithString("groupids", mcp.Description("Comma-separated list of host group IDs to filter by")),
			mcp.WithString("hostids", mcp.Description("Comma-separated list of host IDs to filter by")),
			mcp.WithString("objectids", mcp.Description("Comma-separated list of trigger IDs to filter by")),
			mcp.WithBoolean("acknowledged", mcp.Description("Filter by acknowledged status: true=only acknowledged, false=only unacknowledged")),
			mcp.WithBoolean("suppressed", mcp.Description("Filter by suppressed status: true=only suppressed, false=only unsuppressed")),
			mcp.WithString("severities", mcp.Description("Comma-separated list of severities to filter by (0-5: not classified, info, warning, average, high, disaster)")),
			mcp.WithBoolean("recent", mcp.Description("Return only recently created problems (default: true)")),
			mcp.WithNumber("time_from", mcp.Description("Return only problems that occurred after this Unix timestamp")),
			mcp.WithNumber("time_till", mcp.Description("Return only problems that occurred before this Unix timestamp")),
			mcp.WithNumber("limit", mcp.Description("Max problems to return (default: 100)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getProblemsHandler(ctx, req, logger)
		},
	}
}

func getProblemsHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	params := ProblemGetParams{
		Output:     "extend",
		SelectTags: "extend",
		SortField:  []string{"eventid"},
		SortOrder:  []string{"DESC"},
		Limit:      100,
	}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		if v, ok := args["eventids"].(string); ok && v != "" {
			params.EventIDs = splitAndTrim(v)
		}
		if v, ok := args["groupids"].(string); ok && v != "" {
			params.GroupIDs = splitAndTrim(v)
		}
		if v, ok := args["hostids"].(string); ok && v != "" {
			params.HostIDs = splitAndTrim(v)
		}
		if v, ok := args["objectids"].(string); ok && v != "" {
			params.ObjectIDs = splitAndTrim(v)
		}
		if v, ok := args["acknowledged"].(bool); ok {
			params.Acknowledged = &v
		}
		if v, ok := args["suppressed"].(bool); ok {
			params.Suppressed = &v
		}
		if v, ok := args["severities"].(string); ok && v != "" {
			params.Severities = parseSeverities(v)
		}
		if v, ok := args["recent"].(bool); ok {
			params.Recent = &v
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
	}

	result, err := zabbix.Call("problem.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get problems: %v", err)), nil
	}

	var problems []Problem
	json.Unmarshal(result, &problems)
	jsonData, _ := json.MarshalIndent(problems, "", "  ")
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

func parseSeverities(s string) []int {
	var result []int
	for _, part := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			var severity int
			fmt.Sscanf(trimmed, "%d", &severity)
			if severity >= 0 && severity <= 5 {
				result = append(result, severity)
			}
		}
	}
	return result
}
