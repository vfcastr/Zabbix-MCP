// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package lld

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

type LLDRuleCopyParams struct {
	DiscoveryIDs []string                 `json:"discoveryids"`
	HostIDs      []map[string]interface{} `json:"hostids,omitempty"`
}

func CopyLLDRule(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("copy_lld_rule",
			mcp.WithDescription("Copy low-level discovery rules to the specified hosts. This copies all item prototypes, trigger prototypes, graph prototypes, and host prototypes from the original discovery rules."),
			mcp.WithString("discoveryids", mcp.Description("Comma-separated list of LLD rule IDs to copy"), mcp.Required()),
			mcp.WithString("hostids", mcp.Description("Comma-separated list of destination host IDs to copy the LLD rules to"), mcp.Required()),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return copyLLDRuleHandler(ctx, req, logger)
		},
	}
}

func copyLLDRuleHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	discoveryidsStr, ok := args["discoveryids"].(string)
	if !ok || discoveryidsStr == "" {
		return mcp.NewToolResultError("discoveryids is required"), nil
	}

	hostidsStr, ok := args["hostids"].(string)
	if !ok || hostidsStr == "" {
		return mcp.NewToolResultError("hostids is required"), nil
	}

	var discoveryids []string
	for _, id := range strings.Split(discoveryidsStr, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			discoveryids = append(discoveryids, trimmed)
		}
	}

	// Build hostids as array of objects with hostid key (required by Zabbix API)
	var hostids []map[string]interface{}
	for _, id := range strings.Split(hostidsStr, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			hostids = append(hostids, map[string]interface{}{"hostid": trimmed})
		}
	}

	params := map[string]interface{}{
		"discoveryids": discoveryids,
		"hostids":      hostids,
	}

	result, err := zabbix.Call("discoveryrule.copy", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to copy LLD rules: %v", err)), nil
	}

	var response interface{}
	json.Unmarshal(result, &response)
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":      "LLD rules copied successfully",
		"discoveryids": discoveryids,
		"hostids":      hostidsStr,
		"response":     response,
	}, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}
