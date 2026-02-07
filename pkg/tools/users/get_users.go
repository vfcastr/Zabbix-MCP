// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package users

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
	"github.com/vfcastr/Zabbix-MCP/pkg/utils"
)

type User struct {
	UserID      string `json:"userid"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	URL         string `json:"url"`
	AutoLogin   string `json:"autologin"`
	AutoLogout  string `json:"autologout"`
	Lang        string `json:"lang"`
	Refresh     string `json:"refresh"`
	Theme       string `json:"theme"`
	RowsPerPage string `json:"rows_per_page"`
	Timezone    string `json:"timezone"`
	RoleID      string `json:"roleid"`
}

type UserGetParams struct {
	Output           interface{} `json:"output,omitempty"`
	UserIDs          []string    `json:"userids,omitempty"`
	UserGroupIDs     []string    `json:"usrgrpids,omitempty"`
	MediaTypeIDs     []string    `json:"mediatypeids,omitempty"`
	Search           interface{} `json:"search,omitempty"`
	SelectUserGroups interface{} `json:"selectUsrgrps,omitempty"`
	SelectMedias     interface{} `json:"selectMedias,omitempty"`
	SelectRole       interface{} `json:"selectRole,omitempty"`
	Limit            int         `json:"limit,omitempty"`
}

func GetUsers(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_users",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("Retrieve users from Zabbix."),
			mcp.WithString("userids", mcp.Description("Comma-separated list of user IDs to filter by")),
			mcp.WithString("usrgrpids", mcp.Description("Comma-separated list of user group IDs to filter by")),
			mcp.WithString("search", mcp.Description("Search users by username or name")),
			mcp.WithNumber("limit", mcp.Description("Max users to return (default: 100)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getUsersHandler(ctx, req, logger)
		},
	}
}

func getUsersHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	zabbix, err := client.GetZabbixClientFromContext(ctx, logger)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Zabbix client: %v", err)), nil
	}

	params := UserGetParams{
		Output:           "extend",
		SelectUserGroups: "extend",
		SelectRole:       "extend",
		Limit:            100,
	}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok && args != nil {
		if v, ok := args["userids"].(string); ok && v != "" {
			params.UserIDs = splitAndTrim(v)
		}
		if v, ok := args["usrgrpids"].(string); ok && v != "" {
			params.UserGroupIDs = splitAndTrim(v)
		}
		if v, ok := args["search"].(string); ok && v != "" {
			params.Search = map[string]string{"username": v, "name": v}
		}
		if v, ok := args["limit"].(float64); ok && v > 0 {
			params.Limit = int(v)
		}
	}

	result, err := zabbix.Call("user.get", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get users: %v", err)), nil
	}

	var users []User
	json.Unmarshal(result, &users)
	jsonData, _ := json.MarshalIndent(users, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}

func splitAndTrim(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
