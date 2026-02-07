// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package hosts

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

// UpdateHost creates a tool for updating an existing host in Zabbix
func UpdateHost(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("update_host",
			mcp.WithDescription("Update an existing host in the Zabbix server."),
			mcp.WithString("hostid",
				mcp.Required(),
				mcp.Description("ID of the host to update"),
			),
			mcp.WithString("host",
				mcp.Description("New technical name of the host"),
			),
			mcp.WithString("name",
				mcp.Description("New visible name of the host"),
			),
			mcp.WithNumber("status",
				mcp.Description("Host status: 0=monitored, 1=unmonitored"),
			),
			mcp.WithString("description",
				mcp.Description("New description of the host"),
			),
			mcp.WithString("tags",
				mcp.Description("Tags in format key:value,key2:value2"),
			),
			mcp.WithString("macros",
				mcp.Description("JSON string of host macros, e.g. [{\"macro\":\"{$MYVAR}\",\"value\":\"123\"}]"),
			),
			mcp.WithNumber("inventory_mode",
				mcp.Description("Inventory mode: -1=disabled, 0=manual, 1=automatic"),
			),
			mcp.WithString("inventory",
				mcp.Description("JSON string of inventory fields"),
			),
			mcp.WithNumber("tls_connect",
				mcp.Description("Connections to host: 1=No encryption, 2=PSK, 4=Certificate"),
			),
			mcp.WithNumber("tls_accept",
				mcp.Description("Connections from host: 1=No encryption, 2=PSK, 4=Certificate"),
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
			mcp.WithNumber("monitored_by",
				mcp.Description("Monitored by: 0=Server (default), 1=Proxy, 2=Proxy group"),
			),
			mcp.WithString("proxyid",
				mcp.Description("Proxy ID (if monitored_by=1)"),
			),
			mcp.WithString("proxy_groupid",
				mcp.Description("Proxy Group ID (if monitored_by=2)"),
			),
			mcp.WithNumber("ipmi_authtype",
				mcp.Description("IPMI authentication type"),
			),
			mcp.WithNumber("ipmi_privilege",
				mcp.Description("IPMI privilege level"),
			),
			mcp.WithString("ipmi_username",
				mcp.Description("IPMI username"),
			),
			mcp.WithString("ipmi_password",
				mcp.Description("IPMI password"),
			),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return updateHostHandler(ctx, req, logger)
		},
	}
}

func updateHostHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	logger.Debug("Handling update_host request")

	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to get Zabbix client")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing or invalid arguments"), nil
	}

	hostid, ok := args["hostid"].(string)
	if !ok || hostid == "" {
		return mcp.NewToolResultError("hostid parameter is required"), nil
	}

	params := client.HostUpdateParams{
		HostID: hostid,
	}

	if host, ok := args["host"].(string); ok && host != "" {
		params.Host = host
	}
	if name, ok := args["name"].(string); ok && name != "" {
		params.Name = name
	}
	if status, ok := args["status"].(float64); ok {
		s := int(status)
		params.Status = &s
	}
	if desc, ok := args["description"].(string); ok {
		params.Description = desc
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

	// Advanced options
	if macrosStr, ok := args["macros"].(string); ok && macrosStr != "" {
		var macros []client.Macro
		if err := json.Unmarshal([]byte(macrosStr), &macros); err != nil {
			logger.WithError(err).Warn("Failed to parse macros JSON")
		} else {
			params.Macros = macros
		}
	}

	if mode, ok := args["inventory_mode"].(float64); ok {
		m := int(mode)
		params.InventoryMode = &m
	}

	if invStr, ok := args["inventory"].(string); ok && invStr != "" {
		var inventory map[string]string
		if err := json.Unmarshal([]byte(invStr), &inventory); err != nil {
			logger.WithError(err).Warn("Failed to parse inventory JSON")
		} else {
			params.Inventory = inventory
		}
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

	// Proxy
	if v, ok := args["monitored_by"].(float64); ok {
		val := int(v)
		params.MonitoredBy = &val
	}
	if v, ok := args["proxyid"].(string); ok {
		params.ProxyID = v
	}
	if v, ok := args["proxy_groupid"].(string); ok {
		params.ProxyGroupID = v
	}

	// IPMI
	if v, ok := args["ipmi_authtype"].(float64); ok {
		val := int(v)
		params.IpmiAuthType = &val
	}
	if v, ok := args["ipmi_privilege"].(float64); ok {
		val := int(v)
		params.IpmiPrivilege = &val
	}
	if v, ok := args["ipmi_username"].(string); ok {
		params.IpmiUsername = v
	}
	if v, ok := args["ipmi_password"].(string); ok {
		params.IpmiPassword = v
	}

	result, err := zabbix.Call("host.update", params)
	if err != nil {
		logger.WithError(err).Error("Failed to update host")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update host: %v", err)), nil
	}

	var response struct {
		HostIDs []string `json:"hostids"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		logger.WithError(err).Error("Failed to parse update response")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message": "Host updated successfully",
		"hostids": response.HostIDs,
	}, "", "  ")

	logger.WithField("hostid", hostid).Info("Successfully updated host")
	return mcp.NewToolResultText(string(jsonData)), nil
}
