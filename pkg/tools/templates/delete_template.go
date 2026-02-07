// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package templates

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

// DeleteTemplate creates a tool for deleting templates from Zabbix
func DeleteTemplate(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_template",
			mcp.WithDescription("Delete templates from Zabbix."),
			mcp.WithString("templateids", mcp.Required(), mcp.Description("Comma-separated list of template IDs to delete")),
			mcp.WithBoolean("clear", mcp.Description("If true, also delete items/triggers from unlinked templates (default: false)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteTemplateHandler(ctx, req, logger)
		},
	}
}

func deleteTemplateHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing arguments"), nil
	}

	templateidsStr, _ := args["templateids"].(string)
	if templateidsStr == "" {
		return mcp.NewToolResultError("templateids required"), nil
	}

	params := utilSplitAndTrim(templateidsStr)

	result, err := zabbix.Call("template.delete", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete templates: %v", err)), nil
	}

	var response struct {
		TemplateIDs []string `json:"templateids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Templates deleted", "templateids": response.TemplateIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}

func utilSplitAndTrim(s string) []string {
	var result []string
	for _, v := range strings.Split(s, ",") {
		if t := strings.TrimSpace(v); t != "" {
			result = append(result, t)
		}
	}
	return result
}
