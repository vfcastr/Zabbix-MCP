// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
)

var (
	activeClients sync.Map
)

const (
	ZabbixURL           = "ZABBIX_URL"
	ZabbixToken         = "ZABBIX_TOKEN"
	ZabbixUser          = "ZABBIX_USER"
	ZabbixPassword      = "ZABBIX_PASSWORD"
	ZabbixSkipTLSVerify = "ZABBIX_SKIP_VERIFY"
	ZabbixHeaderToken   = "X-Zabbix-Token"
	ZabbixHeaderURL     = "X-Zabbix-URL"
)

const DefaultZabbixURL = "http://127.0.0.1/api_jsonrpc.php"

// contextKey is a type alias to avoid lint warnings
type contextKey string

// ZabbixClient represents a client for the Zabbix API 7.0
type ZabbixClient struct {
	URL        string
	AuthToken  string
	HTTPClient *http.Client
	Logger     *log.Logger
}

// ZabbixRequest represents a JSON-RPC request to the Zabbix API
type ZabbixRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

// ZabbixResponse represents a JSON-RPC response from the Zabbix API
type ZabbixResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ZabbixError    `json:"error,omitempty"`
	ID      int             `json:"id"`
}

// ZabbixError represents an error from the Zabbix API
type ZabbixError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (e *ZabbixError) Error() string {
	return fmt.Sprintf("Zabbix API error %d: %s - %s", e.Code, e.Message, e.Data)
}

// getEnv retrieves the value of an environment variable or returns a fallback value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// NewZabbixClient creates a new Zabbix client for the given session
func NewZabbixClient(sessionId string, zabbixURL string, skipTLSVerify bool, authToken string, logger *log.Logger) (*ZabbixClient, error) {
	// Create HTTP client with optional TLS skip
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
	}
	httpClient := &http.Client{Transport: tr}

	client := &ZabbixClient{
		URL:        zabbixURL,
		AuthToken:  authToken,
		HTTPClient: httpClient,
		Logger:     logger,
	}

	// Store client for session
	activeClients.Store(sessionId, client)

	return client, nil
}

// GetZabbixClient retrieves the Zabbix client for the given session
func GetZabbixClient(sessionId string) *ZabbixClient {
	if value, ok := activeClients.Load(sessionId); ok {
		return value.(*ZabbixClient)
	}
	return nil
}

// DeleteZabbixClient removes the Zabbix client for the given session
func DeleteZabbixClient(sessionId string) {
	activeClients.Delete(sessionId)
}

// GetZabbixClientFromContext extracts Zabbix client from the MCP context
func GetZabbixClientFromContext(ctx context.Context, logger *log.Logger) (*ZabbixClient, error) {
	session := server.ClientSessionFromContext(ctx)
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}

	logger.WithField("session_id", session.SessionID()).Debug("Retrieving Zabbix client for session")

	// Try to get existing client
	client := GetZabbixClient(session.SessionID())
	if client != nil {
		return client, nil
	}

	logger.WithField("session_id", session.SessionID()).Warn("Zabbix client not found, creating a new one")

	return CreateZabbixClientForSession(ctx, session, logger)
}

// CreateZabbixClientForSession creates a new Zabbix client for a session
func CreateZabbixClientForSession(ctx context.Context, session server.ClientSession, logger *log.Logger) (*ZabbixClient, error) {
	// Get Zabbix URL from context or environment
	zabbixURL, ok := ctx.Value(contextKey(ZabbixURL)).(string)
	if !ok || zabbixURL == "" {
		zabbixURL = getEnv(ZabbixURL, DefaultZabbixURL)
	}

	// Get auth token from context or environment
	authToken, ok := ctx.Value(contextKey(ZabbixToken)).(string)
	if !ok || authToken == "" {
		authToken = getEnv(ZabbixToken, "")
		if authToken == "" {
			// Try user/password authentication (will need to call user.login)
			return nil, fmt.Errorf("zabbix token not provided for session")
		}
	}

	// Check for TLS skip verification
	skipTLSVerify := false
	if skipEnv := getEnv(ZabbixSkipTLSVerify, "false"); skipEnv == "true" || skipEnv == "1" {
		skipTLSVerify = true
	}

	newClient, err := NewZabbixClient(session.SessionID(), zabbixURL, skipTLSVerify, authToken, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Zabbix client: %v", err)
	}

	logger.WithFields(log.Fields{
		"session_id": session.SessionID(),
		"zabbix_url": zabbixURL,
	}).Info("Created Zabbix client for session")

	return newClient, nil
}

// NewSessionHandler initializes a new Zabbix client for the session
func NewSessionHandler(ctx context.Context, session server.ClientSession, logger *log.Logger) {
	_, err := CreateZabbixClientForSession(ctx, session, logger)
	if err != nil {
		logger.WithError(err).Error("NewSessionHandler failed to create Zabbix client")
		return
	}
}

// EndSessionHandler cleans up the Zabbix client when the session ends
func EndSessionHandler(_ context.Context, session server.ClientSession, logger *log.Logger) {
	DeleteZabbixClient(session.SessionID())
	logger.WithField("session_id", session.SessionID()).Info("Cleaned up Zabbix client for session")
}

// Call makes a JSON-RPC call to the Zabbix API
func (c *ZabbixClient) Call(method string, params interface{}) (json.RawMessage, error) {
	request := ZabbixRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	c.Logger.WithFields(log.Fields{
		"method": method,
		"url":    c.URL,
	}).Debug("Making Zabbix API call")

	req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json-rpc")

	// Zabbix 7.0 uses Authorization header with Bearer token
	if c.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var zabbixResp ZabbixResponse
	if err := json.Unmarshal(body, &zabbixResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if zabbixResp.Error != nil {
		return nil, zabbixResp.Error
	}

	return zabbixResp.Result, nil
}

// Tag represents a Zabbix tag
type Tag struct {
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

// Host related API calls

// HostGetParams represents parameters for host.get API call
type HostGetParams struct {
	Output           interface{} `json:"output,omitempty"`
	HostIDs          []string    `json:"hostids,omitempty"`
	GroupIDs         []string    `json:"groupids,omitempty"`
	TemplateIDs      []string    `json:"templateids,omitempty"`
	Search           interface{} `json:"search,omitempty"`
	SearchByAny      bool        `json:"searchByAny,omitempty"`
	SelectGroups     interface{} `json:"selectHostGroups,omitempty"`
	SelectTemplates  interface{} `json:"selectParentTemplates,omitempty"`
	SelectInterfaces interface{} `json:"selectInterfaces,omitempty"`
	SelectTags       interface{} `json:"selectTags,omitempty"`
	SelectMacros     interface{} `json:"selectMacros,omitempty"`
	SelectInventory  interface{} `json:"selectInventory,omitempty"`
	Limit            int         `json:"limit,omitempty"`
}

// HostCreateParams represents parameters for host.create API call
type HostCreateParams struct {
	Host           string              `json:"host"`
	Name           string              `json:"name,omitempty"`
	Groups         []map[string]string `json:"groups"`
	Interfaces     []HostInterface     `json:"interfaces,omitempty"`
	Templates      []map[string]string `json:"templates,omitempty"`
	Description    string              `json:"description,omitempty"`
	Status         int                 `json:"status,omitempty"`
	Tags           []Tag               `json:"tags,omitempty"`
	Macros         []Macro             `json:"macros,omitempty"`
	InventoryMode  int                 `json:"inventory_mode,omitempty"` // -1=disabled, 0=manual, 1=automatic
	Inventory      map[string]string   `json:"inventory,omitempty"`
	TlsConnect     int                 `json:"tls_connect,omitempty"` // 1=No encryption, 2=PSK, 4=Certificate
	TlsAccept      int                 `json:"tls_accept,omitempty"`  // 1=No encryption, 2=PSK, 4=Certificate
	TlsPskIdentity string              `json:"tls_psk_identity,omitempty"`
	TlsPsk         string              `json:"tls_psk,omitempty"`
	TlsIssuer      string              `json:"tls_issuer,omitempty"`
	TlsSubject     string              `json:"tls_subject,omitempty"`
	MonitoredBy    int                 `json:"monitored_by,omitempty"`  // 0=Server, 1=Proxy, 2=Proxy group
	ProxyID        string              `json:"proxyid,omitempty"`       // For monitored_by=1
	ProxyGroupID   string              `json:"proxy_groupid,omitempty"` // For monitored_by=2
	IpmiAuthType   int                 `json:"ipmi_authtype,omitempty"`
	IpmiPrivilege  int                 `json:"ipmi_privilege,omitempty"`
	IpmiUsername   string              `json:"ipmi_username,omitempty"`
	IpmiPassword   string              `json:"ipmi_password,omitempty"`
}

// HostInterface represents a host interface for Zabbix
type HostInterface struct {
	InterfaceID string      `json:"interfaceid,omitempty"`
	Type        string      `json:"type"`  // 1=agent, 2=SNMP, 3=IPMI, 4=JMX
	Main        string      `json:"main"`  // 1=default, 0=not default
	UseIP       string      `json:"useip"` // 1=use IP, 0=use DNS
	IP          string      `json:"ip"`
	DNS         string      `json:"dns"`
	Port        string      `json:"port"`
	Details     interface{} `json:"details,omitempty"`
}

type HostInterfaceDetails struct {
	Version        int    `json:"version,omitempty"`        // 1=SNMPv1, 2=SNMPv2c, 3=SNMPv3
	Bulk           int    `json:"bulk,omitempty"`           // 0=don't use bulk, 1=use bulk
	Community      string `json:"community,omitempty"`      // SNMPv1/v2c community
	SecurityName   string `json:"securityname,omitempty"`   // SNMPv3 security name
	SecurityLevel  int    `json:"securitylevel,omitempty"`  // 0=noAuthNoPriv, 1=authNoPriv, 2=authPriv
	AuthPassphrase string `json:"authpassphrase,omitempty"` // SNMPv3 auth passphrase
	PrivPassphrase string `json:"privpassphrase,omitempty"` // SNMPv3 priv passphrase
	AuthProtocol   int    `json:"authprotocol,omitempty"`   // 0=MD5, 1=SHA1, 2=SHA224, 3=SHA256, 4=SHA384, 5=SHA512
	PrivProtocol   int    `json:"privprotocol,omitempty"`   // 0=DES, 1=AES128, 2=AES192, 3=AES256, 4=AES192C, 5=AES256C
	ContextName    string `json:"contextname,omitempty"`    // SNMPv3 context name
}

type Macro struct {
	Macro       string `json:"macro"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"` // 0=Text, 1=Secret, 2=Vault
}

// HostUpdateParams represents parameters for host.update API call
type HostUpdateParams struct {
	HostID         string            `json:"hostid"`
	Host           string            `json:"host,omitempty"`
	Name           string            `json:"name,omitempty"`
	Status         *int              `json:"status,omitempty"`
	Description    string            `json:"description,omitempty"`
	Tags           []Tag             `json:"tags,omitempty"`
	Macros         []Macro           `json:"macros,omitempty"`
	InventoryMode  *int              `json:"inventory_mode,omitempty"`
	Inventory      map[string]string `json:"inventory,omitempty"`
	TlsConnect     *int              `json:"tls_connect,omitempty"`
	TlsAccept      *int              `json:"tls_accept,omitempty"`
	TlsPskIdentity string            `json:"tls_psk_identity,omitempty"`
	TlsPsk         string            `json:"tls_psk,omitempty"`
	TlsIssuer      string            `json:"tls_issuer,omitempty"`
	TlsSubject     string            `json:"tls_subject,omitempty"`
	MonitoredBy    *int              `json:"monitored_by,omitempty"`
	ProxyID        string            `json:"proxyid,omitempty"`
	ProxyGroupID   string            `json:"proxy_groupid,omitempty"`
	IpmiAuthType   *int              `json:"ipmi_authtype,omitempty"`
	IpmiPrivilege  *int              `json:"ipmi_privilege,omitempty"`
	IpmiUsername   string            `json:"ipmi_username,omitempty"`
	IpmiPassword   string            `json:"ipmi_password,omitempty"`
}

// Item related API calls

// ItemGetParams represents parameters for item.get API call
type ItemGetParams struct {
	Output      interface{} `json:"output,omitempty"`
	ItemIDs     []string    `json:"itemids,omitempty"`
	HostIDs     []string    `json:"hostids,omitempty"`
	Search      interface{} `json:"search,omitempty"`
	SelectHosts interface{} `json:"selectHosts,omitempty"`
	SelectTags  interface{} `json:"selectTags,omitempty"`
	Limit       int         `json:"limit,omitempty"`
}

// ItemCreateParams represents parameters for item.create API call
type ItemCreateParams struct {
	Name        string `json:"name"`
	Key         string `json:"key_"`
	HostID      string `json:"hostid"`
	InterfaceID string `json:"interfaceid,omitempty"`
	Type        int    `json:"type"`       // 0=Zabbix agent, 2=Zabbix trapper, etc.
	ValueType   int    `json:"value_type"` // 0=numeric float, 1=character, 3=numeric unsigned, 4=text
	Delay       string `json:"delay,omitempty"`
	History     string `json:"history,omitempty"`
	Trends      string `json:"trends,omitempty"`
	Tags        []Tag  `json:"tags,omitempty"`
}

// ItemUpdateParams represents parameters for item.update API call
type ItemUpdateParams struct {
	ItemID string `json:"itemid"`
	Status *int   `json:"status,omitempty"` // 0=enabled, 1=disabled
	Delay  string `json:"delay,omitempty"`
	Name   string `json:"name,omitempty"`
	Tags   []Tag  `json:"tags,omitempty"`
}

// Trigger related API calls

// TriggerGetParams represents parameters for trigger.get API call
type TriggerGetParams struct {
	Output        interface{} `json:"output,omitempty"`
	TriggerIDs    []string    `json:"triggerids,omitempty"`
	HostIDs       []string    `json:"hostids,omitempty"`
	MinSeverity   int         `json:"min_severity,omitempty"`
	SelectHosts   interface{} `json:"selectHosts,omitempty"`
	SelectTags    interface{} `json:"selectTags,omitempty"`
	ExpandComment bool        `json:"expandComment,omitempty"`
	Limit         int         `json:"limit,omitempty"`
}

// TriggerCreateParams represents parameters for trigger.create API call
type TriggerCreateParams struct {
	Description string `json:"description"`
	Expression  string `json:"expression"`
	Priority    int    `json:"priority,omitempty"` // 0=not classified, 1=info, 2=warning, 3=average, 4=high, 5=disaster
	Status      int    `json:"status,omitempty"`   // 0=enabled, 1=disabled
	Comments    string `json:"comments,omitempty"`
	Tags        []Tag  `json:"tags,omitempty"`
}

// TriggerUpdateParams represents parameters for trigger.update API call
type TriggerUpdateParams struct {
	TriggerID   string `json:"triggerid"`
	Description string `json:"description,omitempty"`
	Priority    *int   `json:"priority,omitempty"`
	Status      *int   `json:"status,omitempty"`
	Tags        []Tag  `json:"tags,omitempty"`
}

// Template related API calls

// TemplateGetParams represents parameters for template.get API call
type TemplateGetParams struct {
	Output      interface{} `json:"output,omitempty"`
	TemplateIDs []string    `json:"templateids,omitempty"`
	HostIDs     []string    `json:"hostids,omitempty"`
	Search      interface{} `json:"search,omitempty"`
	SelectHosts interface{} `json:"selectHosts,omitempty"`
	SelectTags  interface{} `json:"selectTags,omitempty"`
	Limit       int         `json:"limit,omitempty"`
}

// TemplateCreateParams represents parameters for template.create API call
type TemplateCreateParams struct {
	Host        string              `json:"host"`
	Groups      []map[string]string `json:"groups"`
	Name        string              `json:"name,omitempty"`
	Description string              `json:"description,omitempty"`
	Tags        []Tag               `json:"tags,omitempty"`
}

// TemplateUpdateParams represents parameters for template.update API call
type TemplateUpdateParams struct {
	TemplateID  string `json:"templateid"`
	Host        string `json:"host,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Tags        []Tag  `json:"tags,omitempty"`
}

// Maintenance related API calls

// MaintenanceGetParams represents parameters for maintenance.get API call
type MaintenanceGetParams struct {
	Output            interface{} `json:"output,omitempty"`
	MaintenanceIDs    []string    `json:"maintenanceids,omitempty"`
	HostIDs           []string    `json:"hostids,omitempty"`
	GroupIDs          []string    `json:"groupids,omitempty"`
	SelectHosts       interface{} `json:"selectHosts,omitempty"`
	SelectGroups      interface{} `json:"selectHostGroups,omitempty"`
	SelectTimeperiods interface{} `json:"selectTimeperiods,omitempty"`
	Limit             int         `json:"limit,omitempty"`
}

// MaintenanceCreateParams represents parameters for maintenance.create API call
type MaintenanceCreateParams struct {
	Name            string                  `json:"name"`
	ActiveSince     int64                   `json:"active_since"`
	ActiveTill      int64                   `json:"active_till"`
	HostIDs         []string                `json:"hostids,omitempty"`
	GroupIDs        []string                `json:"groupids,omitempty"`
	Timeperiods     []MaintenanceTimeperiod `json:"timeperiods"`
	Description     string                  `json:"description,omitempty"`
	MaintenanceType int                     `json:"maintenance_type,omitempty"` // 0=with data collection, 1=without
}

// MaintenanceTimeperiod represents a time period for maintenance
type MaintenanceTimeperiod struct {
	TimeperiodType int   `json:"timeperiod_type"` // 0=one time, 2=daily, 3=weekly, 4=monthly
	StartDate      int64 `json:"start_date,omitempty"`
	Period         int   `json:"period"` // duration in seconds
}

// MaintenanceUpdateParams represents parameters for maintenance.update API call
type MaintenanceUpdateParams struct {
	MaintenanceID string `json:"maintenanceid"`
	Name          string `json:"name,omitempty"`
	ActiveTill    int64  `json:"active_till,omitempty"`
	Description   string `json:"description,omitempty"`
}

// Host Group related API calls

// HostGroupGetParams represents parameters for hostgroup.get API call
type HostGroupGetParams struct {
	Output      interface{} `json:"output,omitempty"`
	GroupIDs    []string    `json:"groupids,omitempty"`
	HostIDs     []string    `json:"hostids,omitempty"`
	Search      interface{} `json:"search,omitempty"`
	SelectHosts interface{} `json:"selectHosts,omitempty"`
	Limit       int         `json:"limit,omitempty"`
}

// HostGroupCreateParams represents parameters for hostgroup.create API call
type HostGroupCreateParams struct {
	Name string `json:"name"`
}

// HostGroupUpdateParams represents parameters for hostgroup.update API call
type HostGroupUpdateParams struct {
	GroupID string `json:"groupid"`
	Name    string `json:"name"`
}

// Template Group related API calls
// Note: Zabbix 7.0 uses templategroup.* API

// TemplateGroupGetParams represents parameters for templategroup.get API call
type TemplateGroupGetParams struct {
	Output          interface{} `json:"output,omitempty"`
	GroupIDs        []string    `json:"groupids,omitempty"`
	TemplateIDs     []string    `json:"templateids,omitempty"`
	Search          interface{} `json:"search,omitempty"`
	SelectTemplates interface{} `json:"selectTemplates,omitempty"`
	Limit           int         `json:"limit,omitempty"`
}

// TemplateGroupCreateParams represents parameters for templategroup.create API call
type TemplateGroupCreateParams struct {
	Name string `json:"name"`
}

// TemplateGroupUpdateParams represents parameters for templategroup.update API call
type TemplateGroupUpdateParams struct {
	GroupID string `json:"groupid"`
	Name    string `json:"name"`
}

// Proxy related API calls

// Proxy represents a Zabbix proxy
type Proxy struct {
	ProxyID              string `json:"proxyid"`
	Name                 string `json:"name"`
	ProxyGroupID         string `json:"proxy_groupid"`
	LocalAddress         string `json:"local_address"`
	LocalPort            string `json:"local_port"`
	OperatingMode        string `json:"operating_mode"` // 0=active, 1=passive
	Description          string `json:"description"`
	AllowedAddress       string `json:"allowed_addresses"`
	Address              string `json:"address"`
	Port                 string `json:"port"`
	TlsConnect           string `json:"tls_connect"`
	TlsAccept            string `json:"tls_accept"`
	TlsIssuer            string `json:"tls_issuer"`
	TlsSubject           string `json:"tls_subject"`
	TlsPskIdentity       string `json:"tls_psk_identity"`
	TlsPsk               string `json:"tls_psk"`
	CustomTimeouts       string `json:"custom_timeouts"`
	TimeoutZabbixAgent   string `json:"timeout_zabbix_agent"`
	TimeoutSimpleCheck   string `json:"timeout_simple_check"`
	TimeoutSnmpAgent     string `json:"timeout_snmp_agent"`
	TimeoutExternalCheck string `json:"timeout_external_check"`
	TimeoutDbMonitor     string `json:"timeout_db_monitor"`
	TimeoutHttpAgent     string `json:"timeout_http_agent"`
	TimeoutSshAgent      string `json:"timeout_ssh_agent"`
	TimeoutTelnetAgent   string `json:"timeout_telnet_agent"`
	TimeoutScript        string `json:"timeout_script"`
	LastAccess           string `json:"last_access"`
	Version              string `json:"version"`
	Compatibility        string `json:"compatibility"`
	State                string `json:"state"`
}

// ProxyGetParams represents parameters for proxy.get API call
type ProxyGetParams struct {
	Output        interface{} `json:"output,omitempty"`
	ProxyIDs      []string    `json:"proxyids,omitempty"`
	ProxyGroupIDs []string    `json:"proxy_groupids,omitempty"`
	Search        interface{} `json:"search,omitempty"`
	SelectHosts   interface{} `json:"selectHosts,omitempty"` // Returns associated hosts
	Limit         int         `json:"limit,omitempty"`
}

// ProxyCreateParams represents parameters for proxy.create API call
type ProxyCreateParams struct {
	Name                 string              `json:"name"`
	OperatingMode        int                 `json:"operating_mode"` // 0=active, 1=passive
	ProxyGroupID         string              `json:"proxy_groupid,omitempty"`
	LocalAddress         string              `json:"local_address,omitempty"`
	LocalPort            string              `json:"local_port,omitempty"`
	Description          string              `json:"description,omitempty"`
	AllowedAddress       string              `json:"allowed_addresses,omitempty"` // For active proxies
	Address              string              `json:"address,omitempty"`           // For passive proxies
	Port                 string              `json:"port,omitempty"`              // For passive proxies
	TlsConnect           int                 `json:"tls_connect,omitempty"`       // 1=No encryption, 2=PSK, 4=Certificate
	TlsAccept            int                 `json:"tls_accept,omitempty"`        // 1=No encryption, 2=PSK, 4=Certificate
	TlsPskIdentity       string              `json:"tls_psk_identity,omitempty"`
	TlsPsk               string              `json:"tls_psk,omitempty"`
	TlsIssuer            string              `json:"tls_issuer,omitempty"`
	TlsSubject           string              `json:"tls_subject,omitempty"`
	CustomTimeouts       int                 `json:"custom_timeouts,omitempty"` // 0=disabled, 1=enabled
	TimeoutZabbixAgent   string              `json:"timeout_zabbix_agent,omitempty"`
	TimeoutSimpleCheck   string              `json:"timeout_simple_check,omitempty"`
	TimeoutSnmpAgent     string              `json:"timeout_snmp_agent,omitempty"`
	TimeoutExternalCheck string              `json:"timeout_external_check,omitempty"`
	TimeoutDbMonitor     string              `json:"timeout_db_monitor,omitempty"`
	TimeoutHttpAgent     string              `json:"timeout_http_agent,omitempty"`
	TimeoutSshAgent      string              `json:"timeout_ssh_agent,omitempty"`
	TimeoutTelnetAgent   string              `json:"timeout_telnet_agent,omitempty"`
	TimeoutScript        string              `json:"timeout_script,omitempty"`
	Hosts                []map[string]string `json:"hosts,omitempty"` // [{"hostid": "123"}]
}

// ProxyUpdateParams represents parameters for proxy.update API call
type ProxyUpdateParams struct {
	ProxyID              string              `json:"proxyid"`
	Name                 string              `json:"name,omitempty"`
	OperatingMode        *int                `json:"operating_mode,omitempty"`
	ProxyGroupID         string              `json:"proxy_groupid,omitempty"`
	LocalAddress         string              `json:"local_address,omitempty"`
	LocalPort            string              `json:"local_port,omitempty"`
	Description          string              `json:"description,omitempty"`
	AllowedAddress       string              `json:"allowed_addresses,omitempty"`
	Address              string              `json:"address,omitempty"`
	Port                 string              `json:"port,omitempty"`
	TlsConnect           *int                `json:"tls_connect,omitempty"`
	TlsAccept            *int                `json:"tls_accept,omitempty"`
	TlsPskIdentity       string              `json:"tls_psk_identity,omitempty"`
	TlsPsk               string              `json:"tls_psk,omitempty"`
	TlsIssuer            string              `json:"tls_issuer,omitempty"`
	TlsSubject           string              `json:"tls_subject,omitempty"`
	CustomTimeouts       *int                `json:"custom_timeouts,omitempty"`
	TimeoutZabbixAgent   string              `json:"timeout_zabbix_agent,omitempty"`
	TimeoutSimpleCheck   string              `json:"timeout_simple_check,omitempty"`
	TimeoutSnmpAgent     string              `json:"timeout_snmp_agent,omitempty"`
	TimeoutExternalCheck string              `json:"timeout_external_check,omitempty"`
	TimeoutDbMonitor     string              `json:"timeout_db_monitor,omitempty"`
	TimeoutHttpAgent     string              `json:"timeout_http_agent,omitempty"`
	TimeoutSshAgent      string              `json:"timeout_ssh_agent,omitempty"`
	TimeoutTelnetAgent   string              `json:"timeout_telnet_agent,omitempty"`
	TimeoutScript        string              `json:"timeout_script,omitempty"`
	Hosts                []map[string]string `json:"hosts,omitempty"`
}

// Proxy Group related API calls

// ProxyGroup represents a Zabbix proxy group
type ProxyGroup struct {
	ProxyGroupID  string  `json:"proxy_groupid"`
	Name          string  `json:"name"`
	FailoverDelay string  `json:"failover_delay"`
	MinOnline     string  `json:"min_online"`
	Description   string  `json:"description"`
	State         string  `json:"state"` // 0=unknown, 1=online, 2=degrading, 3=recovering, 4=offline
	Proxies       []Proxy `json:"proxies,omitempty"`
}

// ProxyGroupGetParams represents parameters for proxygroup.get API call
type ProxyGroupGetParams struct {
	Output        interface{} `json:"output,omitempty"`
	ProxyGroupIDs []string    `json:"proxy_groupids,omitempty"`
	Search        interface{} `json:"search,omitempty"`
	SelectProxies interface{} `json:"selectProxies,omitempty"`
	Limit         int         `json:"limit,omitempty"`
}

// ProxyGroupCreateParams represents parameters for proxygroup.create API call
type ProxyGroupCreateParams struct {
	Name          string `json:"name"`
	FailoverDelay string `json:"failover_delay,omitempty"`
	MinOnline     string `json:"min_online,omitempty"`
	Description   string `json:"description,omitempty"`
}

// ProxyGroupUpdateParams represents parameters for proxygroup.update API call
type ProxyGroupUpdateParams struct {
	ProxyGroupID  string `json:"proxy_groupid"`
	Name          string `json:"name,omitempty"`
	FailoverDelay string `json:"failover_delay,omitempty"`
	MinOnline     string `json:"min_online,omitempty"`
	Description   string `json:"description,omitempty"`
}
