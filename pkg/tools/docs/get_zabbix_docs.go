// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package docs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/vfcastr/Zabbix-MCP/pkg/utils"
)

const docFilePath = "/server/zabbix_documentation.md"
const maxResultLength = 50000

func GetZabbixDocs(logger *log.Logger) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_zabbix_docs",
			mcp.WithToolAnnotation(mcp.ToolAnnotation{IdempotentHint: utils.ToBoolPtr(true)}),
			mcp.WithDescription("Search and retrieve Zabbix API documentation. Use this tool to look up API methods, parameters, and examples from the official Zabbix documentation."),
			mcp.WithString("search", mcp.Description("Search term to find in the documentation (e.g., 'host.create', 'trigger', 'template'). If not provided, returns the table of contents.")),
			mcp.WithString("section", mcp.Description("Specific section to retrieve (e.g., 'Host', 'Template', 'Trigger', 'Item'). Use this to get the full documentation for a specific API.")),
			mcp.WithNumber("context_lines", mcp.Description("Number of lines before and after each match to include (default: 20)")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return getZabbixDocsHandler(ctx, req, logger)
		},
	}
}

func getZabbixDocsHandler(ctx context.Context, req mcp.CallToolRequest, logger *log.Logger) (*mcp.CallToolResult, error) {
	content, err := os.ReadFile(docFilePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read documentation file: %v", err)), nil
	}

	docContent := string(content)
	lines := strings.Split(docContent, "\n")

	args, _ := req.Params.Arguments.(map[string]interface{})
	searchTerm, _ := args["search"].(string)
	section, _ := args["section"].(string)
	contextLines := 20
	if cl, ok := args["context_lines"].(float64); ok && cl > 0 {
		contextLines = int(cl)
	}

	// If section is provided, return that section
	if section != "" {
		sectionContent := extractSection(lines, section)
		if sectionContent != "" {
			if len(sectionContent) > maxResultLength {
				sectionContent = sectionContent[:maxResultLength] + "\n\n... [truncated - use 'search' parameter for specific content]"
			}
			return mcp.NewToolResultText(sectionContent), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Section '%s' not found. Try searching for it instead.", section)), nil
	}

	// If search term is provided, search for it
	if searchTerm != "" {
		results := searchInDoc(lines, searchTerm, contextLines)
		if len(results) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No results found for '%s'. Try a different search term.", searchTerm)), nil
		}

		output := fmt.Sprintf("## Search Results for '%s'\n\nFound %d matches:\n\n", searchTerm, len(results))
		for _, result := range results {
			output += result + "\n---\n\n"
		}

		if len(output) > maxResultLength {
			output = output[:maxResultLength] + "\n\n... [truncated - narrow your search term]"
		}

		return mcp.NewToolResultText(output), nil
	}

	// Default: return table of contents (headers only)
	toc := extractTableOfContents(lines)
	result, _ := json.MarshalIndent(map[string]interface{}{
		"message":            "Zabbix API Documentation - Table of Contents",
		"usage":              "Use 'search' parameter to find specific topics or 'section' to retrieve a full section",
		"available_sections": toc,
	}, "", "  ")
	return mcp.NewToolResultText(string(result)), nil
}

func extractSection(lines []string, sectionName string) string {
	var result strings.Builder
	inSection := false
	sectionLevel := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			level := countHashes(line)
			headerText := strings.TrimSpace(strings.TrimLeft(line, "#"))

			if strings.Contains(strings.ToLower(headerText), strings.ToLower(sectionName)) && !inSection {
				inSection = true
				sectionLevel = level
				result.WriteString(line + "\n")
				continue
			}

			if inSection && level <= sectionLevel {
				break
			}
		}

		if inSection {
			result.WriteString(line + "\n")
		}
	}

	return result.String()
}

func searchInDoc(lines []string, searchTerm string, contextLines int) []string {
	var results []string
	searchLower := strings.ToLower(searchTerm)

	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), searchLower) {
			start := max(0, i-contextLines)
			end := min(len(lines), i+contextLines+1)

			var snippet strings.Builder
			snippet.WriteString(fmt.Sprintf("**Match at line %d:**\n```\n", i+1))
			for j := start; j < end; j++ {
				if j == i {
					snippet.WriteString(">>> " + lines[j] + " <<<\n")
				} else {
					snippet.WriteString(lines[j] + "\n")
				}
			}
			snippet.WriteString("```\n")
			results = append(results, snippet.String())

			if len(results) >= 10 {
				break
			}
		}
	}

	return results
}

func extractTableOfContents(lines []string) []string {
	var headers []string
	for _, line := range lines {
		if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "####") {
			headers = append(headers, strings.TrimSpace(line))
		}
	}
	return headers
}

func countHashes(line string) int {
	count := 0
	for _, c := range line {
		if c == '#' {
			count++
		} else {
			break
		}
	}
	return count
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
