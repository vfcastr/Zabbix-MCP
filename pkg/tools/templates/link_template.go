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

func LinkTemplate(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("link_template",
			mcp.WithDescription("Link templates to a host."),
			mcp.WithString("hostid", mcp.Required(), mcp.Description("Host ID")),
			mcp.WithString("templateids", mcp.Required(), mcp.Description("Comma-separated template IDs")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return linkTemplateHandler(ctx, req, logger)
		},
	}
}

func linkTemplateHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
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

	params := map[string]interface{}{
		"hostid":    hostid,
		"templates": templates,
	}

	result, err := zabbix.Call("host.update", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	var response struct {
		HostIDs []string `json:"hostids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Templates linked", "hostids": response.HostIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
