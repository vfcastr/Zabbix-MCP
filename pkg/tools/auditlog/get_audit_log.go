// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package auditlog

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

type AuditLogEntry struct {
	AuditID      string `json:"auditid"`
	UserID       string `json:"userid"`
	Username     string `json:"username"`
	Clock        string `json:"clock"`
	IP           string `json:"ip"`
	Action       string `json:"action"`
	ResourceType string `json:"resourcetype"`
	ResourceID   string `json:"resourceid"`
	ResourceName string `json:"resourcename"`
	RecordSetID  string `json:"recordsetid"`
	Details      string `json:"details"`
}

type AuditLogGetParams struct {
	Output        interface{} `json:"output,omitempty"`
	AuditIDs      []string    `json:"auditids,omitempty"`
	UserIDs       []string    `json:"userids,omitempty"`
	Actions       []int       `json:"filter,omitempty"`
	TimeFrom      int64       `json:"time_from,omitempty"`
	TimeTill      int64       `json:"time_till,omitempty"`
	ResourceTypes []int       `json:"resourcetypes,omitempty"`
	SortField     []string    `json:"sortfield,omitempty"`
	SortOrder     string      `json:"sortorder,omitempty"`
	Limit         int         `json:"limit,omitempty"`
}

func GetAuditLog(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_audit_log",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("Retrieve audit log entries from Zabbix. Useful for tracking user actions, configuration changes, and system events."),
			mcp.WithString("auditids", mcp.Description("Comma-separated list of audit log entry IDs")),
			mcp.WithString("userids", mcp.Description("Comma-separated list of user IDs to filter by")),
			mcp.WithNumber("time_from", mcp.Description("Unix timestamp - return only entries after this time")),
			mcp.WithNumber("time_till", mcp.Description("Unix timestamp - return only entries before this time")),
			mcp.WithString("actions", mcp.Description("Comma-separated action IDs: 0=add, 1=update, 2=delete, 4=login, 5=failed_login, 6=history_clear, 7=logout, 8=execute, 9=config_refresh")),
			mcp.WithString("resourcetypes", mcp.Description("Comma-separated resource type IDs to filter by (e.g., 0=user, 2=host, 3=item, 4=trigger, 15=template)")),
			mcp.WithNumber("limit", mcp.Description("Max entries to return (default: 100, max: 1000)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getAuditLogHandler(ctx, req, logger)
		},
	}
}

func getAuditLogHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	params := map[string]interface{}{
		"output":    "extend",
		"sortfield": "clock",
		"sortorder": "DESC",
		"limit":     100,
	}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		if v, ok := args["auditids"].(string); ok && v != "" {
			params["auditids"] = splitAndTrim(v)
		}
		if v, ok := args["userids"].(string); ok && v != "" {
			params["userids"] = splitAndTrim(v)
		}
		if v, ok := args["time_from"].(float64); ok && v > 0 {
			params["time_from"] = int64(v)
		}
		if v, ok := args["time_till"].(float64); ok && v > 0 {
			params["time_till"] = int64(v)
		}
		if v, ok := args["actions"].(string); ok && v != "" {
			actions := []int{}
			for _, a := range splitAndTrim(v) {
				var action int
				fmt.Sscanf(a, "%d", &action)
				actions = append(actions, action)
			}
			if len(actions) > 0 {
				params["filter"] = map[string]interface{}{"action": actions}
			}
		}
		if v, ok := args["resourcetypes"].(string); ok && v != "" {
			resourceTypes := []int{}
			for _, r := range splitAndTrim(v) {
				var rt int
				fmt.Sscanf(r, "%d", &rt)
				resourceTypes = append(resourceTypes, rt)
			}
			if len(resourceTypes) > 0 {
				if filter, ok := params["filter"].(map[string]interface{}); ok {
					filter["resourcetype"] = resourceTypes
				} else {
					params["filter"] = map[string]interface{}{"resourcetype": resourceTypes}
				}
			}
		}
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit := int(v)
			if limit > 1000 {
				limit = 1000
			}
			params["limit"] = limit
		}
	}

	result, err := zabbix.Call("auditlog.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get audit log: %v", err)), nil
	}

	var entries []AuditLogEntry
	json.Unmarshal(result, &entries)
	jsonData, _ := json.MarshalIndent(entries, "", "  ")
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
