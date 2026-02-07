// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package tools

import (
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/alerts"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/auditlog"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/docs"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/events"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/hostgroups"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/hosts"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/itemprototypes"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/items"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/lld"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/macros"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/maintenance"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/problems"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/proxies"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/proxygroups"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/templategroups"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/templates"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/trends"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/triggerprototypes"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/triggers"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/usergroups"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/userroles"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools/users"
)

// InitTools registers all Zabbix MCP tools with the server
func InitTools(mcpServer *server.MCPServer, logger *log.Logger) {

	// Tools for Host management
	getHostsTool := hosts.GetHosts(logger)
	mcpServer.AddTool(getHostsTool.Tool, getHostsTool.Handler)

	createHostTool := hosts.CreateHost(logger)
	mcpServer.AddTool(createHostTool.Tool, createHostTool.Handler)

	updateHostTool := hosts.UpdateHost(logger)
	mcpServer.AddTool(updateHostTool.Tool, updateHostTool.Handler)

	deleteHostTool := hosts.DeleteHost(logger)
	mcpServer.AddTool(deleteHostTool.Tool, deleteHostTool.Handler)

	// Tools for Item management
	getItemsTool := items.GetItems(logger)
	mcpServer.AddTool(getItemsTool.Tool, getItemsTool.Handler)

	createItemTool := items.CreateItem(logger)
	mcpServer.AddTool(createItemTool.Tool, createItemTool.Handler)

	updateItemTool := items.UpdateItem(logger)
	mcpServer.AddTool(updateItemTool.Tool, updateItemTool.Handler)

	deleteItemTool := items.DeleteItem(logger)
	mcpServer.AddTool(deleteItemTool.Tool, deleteItemTool.Handler)

	getHistoryTool := items.GetHistory(logger)
	mcpServer.AddTool(getHistoryTool.Tool, getHistoryTool.Handler)

	// Tools for Trigger management
	getTriggersTool := triggers.GetTriggers(logger)
	mcpServer.AddTool(getTriggersTool.Tool, getTriggersTool.Handler)

	createTriggerTool := triggers.CreateTrigger(logger)
	mcpServer.AddTool(createTriggerTool.Tool, createTriggerTool.Handler)

	updateTriggerTool := triggers.UpdateTrigger(logger)
	mcpServer.AddTool(updateTriggerTool.Tool, updateTriggerTool.Handler)

	deleteTriggerTool := triggers.DeleteTrigger(logger)
	mcpServer.AddTool(deleteTriggerTool.Tool, deleteTriggerTool.Handler)

	// Tools for Template management
	getTemplatesTool := templates.GetTemplates(logger)
	mcpServer.AddTool(getTemplatesTool.Tool, getTemplatesTool.Handler)

	linkTemplateTool := templates.LinkTemplate(logger)
	mcpServer.AddTool(linkTemplateTool.Tool, linkTemplateTool.Handler)

	unlinkTemplateTool := templates.UnlinkTemplate(logger)
	mcpServer.AddTool(unlinkTemplateTool.Tool, unlinkTemplateTool.Handler)

	createTemplateTool := templates.CreateTemplate(logger)
	mcpServer.AddTool(createTemplateTool.Tool, createTemplateTool.Handler)

	updateTemplateTool := templates.UpdateTemplate(logger)
	mcpServer.AddTool(updateTemplateTool.Tool, updateTemplateTool.Handler)

	deleteTemplateTool := templates.DeleteTemplate(logger)
	mcpServer.AddTool(deleteTemplateTool.Tool, deleteTemplateTool.Handler)

	// Tools for Maintenance management
	getMaintenanceTool := maintenance.GetMaintenance(logger)
	mcpServer.AddTool(getMaintenanceTool.Tool, getMaintenanceTool.Handler)

	createMaintenanceTool := maintenance.CreateMaintenance(logger)
	mcpServer.AddTool(createMaintenanceTool.Tool, createMaintenanceTool.Handler)

	updateMaintenanceTool := maintenance.UpdateMaintenance(logger)
	mcpServer.AddTool(updateMaintenanceTool.Tool, updateMaintenanceTool.Handler)

	deleteMaintenanceTool := maintenance.DeleteMaintenance(logger)
	mcpServer.AddTool(deleteMaintenanceTool.Tool, deleteMaintenanceTool.Handler)

	// Tools for Host Group management
	getHostGroupsTool := hostgroups.GetHostGroups(logger)
	mcpServer.AddTool(getHostGroupsTool.Tool, getHostGroupsTool.Handler)

	createHostGroupTool := hostgroups.CreateHostGroup(logger)
	mcpServer.AddTool(createHostGroupTool.Tool, createHostGroupTool.Handler)

	updateHostGroupTool := hostgroups.UpdateHostGroup(logger)
	mcpServer.AddTool(updateHostGroupTool.Tool, updateHostGroupTool.Handler)

	deleteHostGroupTool := hostgroups.DeleteHostGroup(logger)
	mcpServer.AddTool(deleteHostGroupTool.Tool, deleteHostGroupTool.Handler)

	// Tools for Template Group management
	getTemplateGroupsTool := templategroups.GetTemplateGroups(logger)
	mcpServer.AddTool(getTemplateGroupsTool.Tool, getTemplateGroupsTool.Handler)

	createTemplateGroupTool := templategroups.CreateTemplateGroup(logger)
	mcpServer.AddTool(createTemplateGroupTool.Tool, createTemplateGroupTool.Handler)

	updateTemplateGroupTool := templategroups.UpdateTemplateGroup(logger)
	mcpServer.AddTool(updateTemplateGroupTool.Tool, updateTemplateGroupTool.Handler)

	deleteTemplateGroupTool := templategroups.DeleteTemplateGroup(logger)
	mcpServer.AddTool(deleteTemplateGroupTool.Tool, deleteTemplateGroupTool.Handler)

	// Tools for Proxy management
	getProxiesTool := proxies.GetProxies(logger)
	mcpServer.AddTool(getProxiesTool.Tool, getProxiesTool.Handler)

	createProxyTool := proxies.CreateProxy(logger)
	mcpServer.AddTool(createProxyTool.Tool, createProxyTool.Handler)

	updateProxyTool := proxies.UpdateProxy(logger)
	mcpServer.AddTool(updateProxyTool.Tool, updateProxyTool.Handler)

	deleteProxyTool := proxies.DeleteProxies(logger)
	mcpServer.AddTool(deleteProxyTool.Tool, deleteProxyTool.Handler)

	// Tools for Proxy Group management
	getProxyGroupsTool := proxygroups.GetProxyGroups(logger)
	mcpServer.AddTool(getProxyGroupsTool.Tool, getProxyGroupsTool.Handler)

	createProxyGroupTool := proxygroups.CreateProxyGroup(logger)
	mcpServer.AddTool(createProxyGroupTool.Tool, createProxyGroupTool.Handler)

	updateProxyGroupTool := proxygroups.UpdateProxyGroup(logger)
	mcpServer.AddTool(updateProxyGroupTool.Tool, updateProxyGroupTool.Handler)

	deleteProxyGroupTool := proxygroups.DeleteProxyGroups(logger)
	mcpServer.AddTool(deleteProxyGroupTool.Tool, deleteProxyGroupTool.Handler)

	// Tools for Problem management
	getProblemsTool := problems.GetProblems(logger)
	mcpServer.AddTool(getProblemsTool.Tool, getProblemsTool.Handler)

	// Tools for Event management
	getEventsTool := events.GetEvents(logger)
	mcpServer.AddTool(getEventsTool.Tool, getEventsTool.Handler)

	acknowledgeEventTool := events.AcknowledgeEvent(logger)
	mcpServer.AddTool(acknowledgeEventTool.Tool, acknowledgeEventTool.Handler)

	// Tools for Trend management
	getTrendsTool := trends.GetTrends(logger)
	mcpServer.AddTool(getTrendsTool.Tool, getTrendsTool.Handler)

	// Tools for Alert management
	getAlertsTool := alerts.GetAlerts(logger)
	mcpServer.AddTool(getAlertsTool.Tool, getAlertsTool.Handler)

	// Tools for User management
	getUsersTool := users.GetUsers(logger)
	mcpServer.AddTool(getUsersTool.Tool, getUsersTool.Handler)

	createUserTool := users.CreateUser(logger)
	mcpServer.AddTool(createUserTool.Tool, createUserTool.Handler)

	updateUserTool := users.UpdateUser(logger)
	mcpServer.AddTool(updateUserTool.Tool, updateUserTool.Handler)

	deleteUserTool := users.DeleteUser(logger)
	mcpServer.AddTool(deleteUserTool.Tool, deleteUserTool.Handler)

	// Tools for User Group management
	getUserGroupsTool := usergroups.GetUserGroups(logger)
	mcpServer.AddTool(getUserGroupsTool.Tool, getUserGroupsTool.Handler)

	createUserGroupTool := usergroups.CreateUserGroup(logger)
	mcpServer.AddTool(createUserGroupTool.Tool, createUserGroupTool.Handler)

	updateUserGroupTool := usergroups.UpdateUserGroup(logger)
	mcpServer.AddTool(updateUserGroupTool.Tool, updateUserGroupTool.Handler)

	deleteUserGroupTool := usergroups.DeleteUserGroup(logger)
	mcpServer.AddTool(deleteUserGroupTool.Tool, deleteUserGroupTool.Handler)

	// Tools for User Role management
	getUserRolesTool := userroles.GetUserRoles(logger)
	mcpServer.AddTool(getUserRolesTool.Tool, getUserRolesTool.Handler)

	createUserRoleTool := userroles.CreateUserRole(logger)
	mcpServer.AddTool(createUserRoleTool.Tool, createUserRoleTool.Handler)

	updateUserRoleTool := userroles.UpdateUserRole(logger)
	mcpServer.AddTool(updateUserRoleTool.Tool, updateUserRoleTool.Handler)

	deleteUserRoleTool := userroles.DeleteUserRole(logger)
	mcpServer.AddTool(deleteUserRoleTool.Tool, deleteUserRoleTool.Handler)

	// Tools for User Macro management (host-level)
	getUserMacrosTool := macros.GetUserMacros(logger)
	mcpServer.AddTool(getUserMacrosTool.Tool, getUserMacrosTool.Handler)

	createUserMacroTool := macros.CreateUserMacro(logger)
	mcpServer.AddTool(createUserMacroTool.Tool, createUserMacroTool.Handler)

	updateUserMacroTool := macros.UpdateUserMacro(logger)
	mcpServer.AddTool(updateUserMacroTool.Tool, updateUserMacroTool.Handler)

	deleteUserMacroTool := macros.DeleteUserMacro(logger)
	mcpServer.AddTool(deleteUserMacroTool.Tool, deleteUserMacroTool.Handler)

	// Tools for Global Macro management
	getGlobalMacrosTool := macros.GetGlobalMacros(logger)
	mcpServer.AddTool(getGlobalMacrosTool.Tool, getGlobalMacrosTool.Handler)

	createGlobalMacroTool := macros.CreateGlobalMacro(logger)
	mcpServer.AddTool(createGlobalMacroTool.Tool, createGlobalMacroTool.Handler)

	updateGlobalMacroTool := macros.UpdateGlobalMacro(logger)
	mcpServer.AddTool(updateGlobalMacroTool.Tool, updateGlobalMacroTool.Handler)

	deleteGlobalMacroTool := macros.DeleteGlobalMacro(logger)
	mcpServer.AddTool(deleteGlobalMacroTool.Tool, deleteGlobalMacroTool.Handler)

	// Tools for LLD Rule management
	getLLDRulesTool := lld.GetLLDRules(logger)
	mcpServer.AddTool(getLLDRulesTool.Tool, getLLDRulesTool.Handler)

	createLLDRuleTool := lld.CreateLLDRule(logger)
	mcpServer.AddTool(createLLDRuleTool.Tool, createLLDRuleTool.Handler)

	updateLLDRuleTool := lld.UpdateLLDRule(logger)
	mcpServer.AddTool(updateLLDRuleTool.Tool, updateLLDRuleTool.Handler)

	deleteLLDRuleTool := lld.DeleteLLDRule(logger)
	mcpServer.AddTool(deleteLLDRuleTool.Tool, deleteLLDRuleTool.Handler)

	copyLLDRuleTool := lld.CopyLLDRule(logger)
	mcpServer.AddTool(copyLLDRuleTool.Tool, copyLLDRuleTool.Handler)

	// Tools for Item Prototype management
	getItemPrototypesTool := itemprototypes.GetItemPrototypes(logger)
	mcpServer.AddTool(getItemPrototypesTool.Tool, getItemPrototypesTool.Handler)

	createItemPrototypeTool := itemprototypes.CreateItemPrototype(logger)
	mcpServer.AddTool(createItemPrototypeTool.Tool, createItemPrototypeTool.Handler)

	updateItemPrototypeTool := itemprototypes.UpdateItemPrototype(logger)
	mcpServer.AddTool(updateItemPrototypeTool.Tool, updateItemPrototypeTool.Handler)

	deleteItemPrototypeTool := itemprototypes.DeleteItemPrototype(logger)
	mcpServer.AddTool(deleteItemPrototypeTool.Tool, deleteItemPrototypeTool.Handler)

	// Tools for Trigger Prototype management
	getTriggerPrototypesTool := triggerprototypes.GetTriggerPrototypes(logger)
	mcpServer.AddTool(getTriggerPrototypesTool.Tool, getTriggerPrototypesTool.Handler)

	createTriggerPrototypeTool := triggerprototypes.CreateTriggerPrototype(logger)
	mcpServer.AddTool(createTriggerPrototypeTool.Tool, createTriggerPrototypeTool.Handler)

	updateTriggerPrototypeTool := triggerprototypes.UpdateTriggerPrototype(logger)
	mcpServer.AddTool(updateTriggerPrototypeTool.Tool, updateTriggerPrototypeTool.Handler)

	deleteTriggerPrototypeTool := triggerprototypes.DeleteTriggerPrototype(logger)
	mcpServer.AddTool(deleteTriggerPrototypeTool.Tool, deleteTriggerPrototypeTool.Handler)

	// Tools for Audit Log
	getAuditLogTool := auditlog.GetAuditLog(logger)
	mcpServer.AddTool(getAuditLogTool.Tool, getAuditLogTool.Handler)

	// Tools for Documentation
	getZabbixDocsTool := docs.GetZabbixDocs(logger)
	mcpServer.AddTool(getZabbixDocsTool.Tool, getZabbixDocsTool.Handler)
}
