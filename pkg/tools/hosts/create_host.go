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

// CreateHost creates a tool for creating a new host in Zabbix
func CreateHost(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_host",
			mcp.WithDescription("Create a new host in the Zabbix server."),
			mcp.WithString("host",
				mcp.Required(),
				mcp.Description("Technical name of the host"),
			),
			mcp.WithString("groupids",
				mcp.Required(),
				mcp.Description("Comma-separated list of host group IDs to add the host to"),
			),
			mcp.WithString("name",
				mcp.Description("Visible name of the host (defaults to technical name)"),
			),
			mcp.WithString("ip",
				mcp.Description("IP address for the default agent interface"),
			),
			mcp.WithString("dns",
				mcp.Description("DNS name for the default interface"),
			),
			mcp.WithString("port",
				mcp.Description("Port for the default interface (default: 10050)"),
			),
			mcp.WithString("templateids",
				mcp.Description("Comma-separated list of template IDs to link"),
			),
			mcp.WithString("description",
				mcp.Description("Description of the host"),
			),
			mcp.WithString("tags",
				mcp.Description("Tags in format key:value,key2:value2"),
			),
			mcp.WithString("interfaces",
				mcp.Description("JSON string of interfaces (overrides ip/dns/port arguments). Allows advanced config like SNMP details."),
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
			return createHostHandler(ctx, req, logger)
		},
	}
}

func createHostHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	logger.Debug("Handling create_host request")

	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to get Zabbix client")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || args == nil {
		return mcp.NewToolResultError("Missing or invalid arguments"), nil
	}

	host, ok := args["host"].(string)
	if !ok || host == "" {
		return mcp.NewToolResultError("host parameter is required"), nil
	}

	groupidsStr, ok := args["groupids"].(string)
	if !ok || groupidsStr == "" {
		return mcp.NewToolResultError("groupids parameter is required"), nil
	}

	var groups []map[string]string
	for _, gid := range splitAndTrim(groupidsStr) {
		groups = append(groups, map[string]string{"groupid": gid})
	}

	params := client.HostCreateParams{
		Host:   host,
		Groups: groups,
	}

	if name, ok := args["name"].(string); ok && name != "" {
		params.Name = name
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
		params.InventoryMode = int(mode)
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

	// Proxy
	if v, ok := args["monitored_by"].(float64); ok {
		params.MonitoredBy = int(v)
	}
	if v, ok := args["proxyid"].(string); ok {
		params.ProxyID = v
	}
	if v, ok := args["proxy_groupid"].(string); ok {
		params.ProxyGroupID = v
	}

	// IPMI
	if v, ok := args["ipmi_authtype"].(float64); ok {
		params.IpmiAuthType = int(v)
	}
	if v, ok := args["ipmi_privilege"].(float64); ok {
		params.IpmiPrivilege = int(v)
	}
	if v, ok := args["ipmi_username"].(string); ok {
		params.IpmiUsername = v
	}
	if v, ok := args["ipmi_password"].(string); ok {
		params.IpmiPassword = v
	}

	// Interfaces - Check for advanced JSON input first
	if interfacesStr, ok := args["interfaces"].(string); ok && interfacesStr != "" {
		var interfaces []client.HostInterface
		if err := json.Unmarshal([]byte(interfacesStr), &interfaces); err != nil {
			logger.WithError(err).Warn("Failed to parse interfaces JSON")
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse interfaces JSON: %v", err)), nil
		}
		params.Interfaces = interfaces
	} else {
		// Fallback to simple IP/DNS/Port
		ip, hasIP := args["ip"].(string)
		dns, hasDNS := args["dns"].(string)
		if hasIP || hasDNS {
			iface := client.HostInterface{
				Type: "1",
				Main: "1",
				Port: "10050",
			}
			if port, ok := args["port"].(string); ok && port != "" {
				iface.Port = port
			}
			if hasIP && ip != "" {
				iface.UseIP = "1"
				iface.IP = ip
			} else {
				iface.UseIP = "0"
				iface.DNS = dns
			}
			params.Interfaces = []client.HostInterface{iface}
		}
	}

	if templateidsStr, ok := args["templateids"].(string); ok && templateidsStr != "" {
		var templates []map[string]string
		for _, tid := range splitAndTrim(templateidsStr) {
			templates = append(templates, map[string]string{"templateid": tid})
		}
		params.Templates = templates
	}

	result, err := zabbix.Call("host.create", params)
	if err != nil {
		logger.WithError(err).Error("Failed to create host")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create host: %v", err)), nil
	}

	var response struct {
		HostIDs []string `json:"hostids"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		logger.WithError(err).Error("Failed to parse create response")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"message": "Host created successfully",
		"hostids": response.HostIDs,
	}, "", "  ")

	logger.WithField("hostids", response.HostIDs).Info("Successfully created host")
	return mcp.NewToolResultText(string(jsonData)), nil
}
