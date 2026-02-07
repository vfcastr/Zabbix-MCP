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

// UpdateProxy creates a tool for updating a Zabbix proxy
func UpdateProxy(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_update_proxy",
			mcp.WithDescription("Update an existing proxy in Zabbix."),
			mcp.WithString("proxyid",
				mcp.Description("ID of the proxy to update"),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Name of the proxy"),
			),
			mcp.WithNumber("operating_mode",
				mcp.Description("Proxy mode: 0=active, 1=passive"),
			),
			mcp.WithString("proxy_groupid",
				mcp.Description("ID of the proxy group"),
			),
			mcp.WithString("local_address",
				mcp.Description("Local address (for active proxy)"),
			),
			mcp.WithString("local_port",
				mcp.Description("Local port (for active proxy)"),
			),
			mcp.WithString("address",
				mcp.Description("Address (for passive proxy)"),
			),
			mcp.WithString("port",
				mcp.Description("Port (for passive proxy)"),
			),
			mcp.WithString("allowed_addresses",
				mcp.Description("Allowed addresses (comma-separated, for active proxy)"),
			),
			mcp.WithString("description",
				mcp.Description("Description of the proxy"),
			),
			mcp.WithNumber("tls_connect",
				mcp.Description("Connections to proxy: 1=No encryption, 2=PSK, 4=Certificate"),
			),
			mcp.WithNumber("tls_accept",
				mcp.Description("Connections from proxy: 1=No encryption, 2=PSK, 4=Certificate"),
			),
			mcp.WithString("tls_psk_identity",
				mcp.Description("PSK identity"),
			),
			mcp.WithString("tls_psk",
				mcp.Description("PSK value"),
			),
			mcp.WithString("tls_issuer",
				mcp.Description("Certificate issuer"),
			),
			mcp.WithString("tls_subject",
				mcp.Description("Certificate subject"),
			),
			mcp.WithString("hostids",
				mcp.Description("Comma-separated list of host IDs to be monitored by this proxy"),
			),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateProxyHandler(ctx, req, logger)
		},
	}
}

func updateProxyHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	logger.Debug("Handling zabbix_update_proxy request")

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

	proxyID, ok := args["proxyid"].(string)
	if !ok || proxyID == "" {
		return mcp.NewToolResultError("proxyid is required"), nil
	}

	params := client.ProxyUpdateParams{
		ProxyID: proxyID,
	}

	if v, ok := args["name"].(string); ok {
		params.Name = v
	}
	if v, ok := args["operating_mode"].(float64); ok {
		val := int(v)
		params.OperatingMode = &val
	}
	if v, ok := args["proxy_groupid"].(string); ok {
		params.ProxyGroupID = v
	}
	if v, ok := args["local_address"].(string); ok {
		params.LocalAddress = v
	}
	if v, ok := args["local_port"].(string); ok {
		params.LocalPort = v
	}
	if v, ok := args["address"].(string); ok {
		params.Address = v
	}
	if v, ok := args["port"].(string); ok {
		params.Port = v
	}
	if v, ok := args["allowed_addresses"].(string); ok {
		params.AllowedAddress = v
	}
	if v, ok := args["description"].(string); ok {
		params.Description = v
	}

	// Encryption
	if v, ok := args["tls_connect"].(float64); ok {
		val := int(v)
		params.TlsConnect = &val
	}
	if v, ok := args["tls_accept"].(float64); ok {
		val := int(v)
		params.TlsAccept = &val
	}
	if v, ok := args["tls_psk_identity"].(string); ok {
		params.TlsPskIdentity = v
	}
	if v, ok := args["tls_psk"].(string); ok {
		params.TlsPsk = v
	}
	if v, ok := args["tls_issuer"].(string); ok {
		params.TlsIssuer = v
	}
	if v, ok := args["tls_subject"].(string); ok {
		params.TlsSubject = v
	}

	// Hosts
	if hostids, ok := args["hostids"].(string); ok && hostids != "" {
		for _, id := range utils.SplitAndTrim(hostids) {
			params.Hosts = append(params.Hosts, map[string]string{"hostid": id})
		}
	}

	result, err := zabbix.Call("proxy.update", params)
	if err != nil {
		logger.WithError(err).Error("Failed to update proxy")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update proxy: %v", err)), nil
	}

	var response struct {
		ProxyIDs []string `json:"proxyids"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		logger.WithError(err).Error("Failed to parse update proxy response")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "Proxy updated successfully",
		"proxyids": response.ProxyIDs,
	}, "", "  ")

	logger.WithField("proxyid", proxyID).Info("Successfully updated proxy")
	return mcp.NewToolResultText(string(jsonData)), nil
}
