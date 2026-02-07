// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package proxies

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
	"github.com/vfcastr/Zabbix-MCP/pkg/utils"
)

// DeleteProxies creates a tool for deleting Zabbix proxies
func DeleteProxies(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_delete_proxies",
			mcp.WithDescription("Delete one or more proxies from the Zabbix server."),
			mcp.WithString("proxyids",
				mcp.Description("Comma-separated list of proxy IDs to delete"),
				mcp.Required(),
			),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return deleteProxiesHandler(ctx, req, logger)
		},
	}
}

func deleteProxiesHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	logger.Debug("Handling zabbix_delete_proxies request")

	// Get Zabbix client from context
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to get Zabbix client")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}

	proxyIDsStr, ok := args["proxyids"].(string)
	if !ok || proxyIDsStr == "" {
		return mcp.NewToolResultError("proxyids is required"), nil
	}

	proxyIDs := utils.SplitAndTrim(proxyIDsStr)

	result, err := zabbix.Call("proxy.delete", proxyIDs)
	if err != nil {
		logger.WithError(err).Error("Failed to delete proxies")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete proxies: %v", err)), nil
	}

	var response struct {
		ProxyIDs []string `json:"proxyids"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		logger.WithError(err).Error("Failed to parse delete proxies response")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "Proxies deleted successfully",
		"proxyids": response.ProxyIDs,
	}, "", "  ")

	logger.WithField("proxy_count", len(response.ProxyIDs)).Info("Successfully deleted proxies")
	return mcp.NewToolResultText(string(jsonData)), nil
}
