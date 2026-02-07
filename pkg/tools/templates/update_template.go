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

// UpdateTemplate creates a tool for updating an existing template in Zabbix
func UpdateTemplate(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("update_template",
			mcp.WithDescription("Update an existing template in Zabbix."),
			mcp.WithString("templateid", mcp.Required(), mcp.Description("Template ID")),
			mcp.WithString("host", mcp.Description("New technical name")),
			mcp.WithString("name", mcp.Description("New visible name")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithString("tags", mcp.Description("Tags in format key:value,key2:value2")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateTemplateHandler(ctx, req, logger)
		},
	}
}

func updateTemplateHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing arguments"), nil
	}

	templateid, _ := args["templateid"].(string)
	if templateid == "" {
		return mcp.NewToolResultError("templateid is required"), nil
	}

	params := client.TemplateUpdateParams{
		TemplateID: templateid,
	}

	if v, ok := args["host"].(string); ok && v != "" {
		params.Host = v
	}
	if v, ok := args["name"].(string); ok && v != "" {
		params.Name = v
	}
	if v, ok := args["description"].(string); ok {
		params.Description = v
	}

	if tagsStr, ok := args["tags"].(string); ok && tagsStr != "" {
		var tags []client.Tag
		for _, tagPart := range strings.Split(tagsStr, ",") {
			parts := strings.SplitN(strings.TrimSpace(tagPart), ":", 2)
			if len(parts) == 2 {
				tags = append(tags, client.Tag{Tag: strings.TrimSpace(parts[0]), Value: strings.TrimSpace(parts[1])})
			} else if len(parts) == 1 {
				tags = append(tags, client.Tag{Tag: strings.TrimSpace(parts[0]), Value: ""})
			}
		}
		params.Tags = tags
	}

	result, err := zabbix.Call("template.update", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update template: %v", err)), nil
	}

	var response struct {
		TemplateIDs []string `json:"templateids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Template updated", "templateids": response.TemplateIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
