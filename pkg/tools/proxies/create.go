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

// CreateProxy creates a tool for creating a new Zabbix proxy
func CreateProxy(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("zabbix_create_proxy",
			mcp.WithDescription("Create a new proxy in Zabbix."),
			mcp.WithString("name",
				mcp.Description("Name of the proxy"),
				mcp.Required(),
			),
			mcp.WithNumber("operating_mode",
				mcp.Description("Proxy mode: 0=active (default), 1=passive"),
				mcp.Required(),
			),
			mcp.WithString("proxy_groupid",
				mcp.Description("ID of the proxy group to add the proxy to"),
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
				mcp.Description("PSK identity (required if tls_connect/accept=2)"),
			),
			mcp.WithString("tls_psk",
				mcp.Description("PSK value (required if tls_connect/accept=2)"),
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
			return createProxyHandler(ctx, req, logger)
		},
	}
}

func createProxyHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	logger.Debug("Handling zabbix_create_proxy request")

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

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}

	operatingMode, ok := args["operating_mode"].(float64)
	if !ok {
		return mcp.NewToolResultError("operating_mode is required"), nil
	}

	params := client.ProxyCreateParams{
		Name:          name,
		OperatingMode: int(operatingMode),
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
		params.TlsConnect = int(v)
	}
	if v, ok := args["tls_accept"].(float64); ok {
		params.TlsAccept = int(v)
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

	result, err := zabbix.Call("proxy.create", params)
	if err != nil {
		logger.WithError(err).Error("Failed to create proxy")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create proxy: %v", err)), nil
	}

	var response struct {
		ProxyIDs []string `json:"proxyids"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		logger.WithError(err).Error("Failed to parse create proxy response")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message":  "Proxy created successfully",
		"proxyids": response.ProxyIDs,
	}, "", "  ")

	logger.WithField("proxyids", response.ProxyIDs).Info("Successfully created proxy")
	return mcp.NewToolResultText(string(jsonData)), nil
}
