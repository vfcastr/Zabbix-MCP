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

// CreateTemplate creates a tool for creating a new template in Zabbix
func CreateTemplate(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_template",
			mcp.WithDescription("Create a new template in Zabbix."),
			mcp.WithString("host", mcp.Required(), mcp.Description("Technical name of the template")),
			mcp.WithString("groupids", mcp.Required(), mcp.Description("Comma-separated list of template group IDs")),
			mcp.WithString("name", mcp.Description("Visible name (defaults to technical name)")),
			mcp.WithString("description", mcp.Description("Description")),
			mcp.WithString("tags", mcp.Description("Tags in format key:value,key2:value2")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return createTemplateHandler(ctx, req, logger)
		},
	}
}

func createTemplateHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing arguments"), nil
	}

	host, _ := args["host"].(string)
	groupidsStr, _ := args["groupids"].(string)
	if host == "" || groupidsStr == "" {
		return mcp.NewToolResultError("host and groupids are required"), nil
	}

	var groups []map[string]string
	for _, gid := range strings.Split(groupidsStr, ",") {
		gid = strings.TrimSpace(gid)
		if gid != "" {
			groups = append(groups, map[string]string{"groupid": gid})
		}
	}

	params := client.TemplateCreateParams{
		Host:   host,
		Groups: groups,
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

	result, err := zabbix.Call("template.create", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create template: %v", err)), nil
	}

	var response struct {
		TemplateIDs []string `json:"templateids"`
	}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{"message": "Template created", "templateids": response.TemplateIDs}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
