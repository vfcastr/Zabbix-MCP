// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version (added at compile time)
	Version = "1.0.0"

	// GitCommit is the git commit SHA (added at compile time)
	GitCommit = "unknown"

	// BuildDate is the build date (added at compile time)
	BuildDate = "unknown"
)

// GetVersion returns a formatted version string
func GetVersion() string {
	return fmt.Sprintf("zabbix-mcp-server %s (commit: %s, built: %s, %s/%s)",
		Version,
		GitCommit,
		BuildDate,
		runtime.GOOS,
		runtime.GOARCH,
	)
}

// GetVersionInfo returns version information as a map
func GetVersionInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"git_commit": GitCommit,
		"build_date": BuildDate,
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}
}
