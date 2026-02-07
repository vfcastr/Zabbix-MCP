// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
)

func UnlinkTemplate(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("unlink_template",
			mcp.WithDescription("Unlink templates from a host."),
			mcp.WithString("hostid", mcp.Required(), mcp.Description("Host ID")),
			mcp.WithString("templateids", mcp.Required(), mcp.Description("Comma-separated template IDs")),
			mcp.WithBoolean("clear", mcp.Description("If true, also delete items/triggers from unlinked templates")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return unlinkTemplateHandler(ctx, req, logger)
		},
	}
}

func unlinkTemplateHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing arguments"), nil
	}

	hostid, _ := args["hostid"].(string)
	templateidsStr, _ := args["templateids"].(string)
	if hostid == "" || templateidsStr == "" {
		return mcp.NewToolResultError("hostid and templateids required"), nil
	}

	var templates []map[string]string
	for _, tid := range splitAndTrim(templateidsStr) {
		templates = append(templates, map[string]string{"templateid": tid})
	}

	params := map[string]interface{}{"hostid": hostid}

	if clear, ok := args["clear"].(bool); ok && clear {
		params["templates_clear"] = templates
	} else {
		params["templates_clear"] = templates
	}

	result, err := zabbix.Call("host.update", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	var response struct {
		HostIDs []string `json:"hostids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Templates unlinked", "hostids": response.HostIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
