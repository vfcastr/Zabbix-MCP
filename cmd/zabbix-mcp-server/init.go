// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// initLogger initializes and returns a configured logger
func initLogger(logFile string) (*log.Logger, error) {
	logger := log.New()

	// Set log format
	logger.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// Set log level from environment or default to Info
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}

	// Configure output
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
		logger.SetOutput(file)
	} else {
		// For stdio mode, log to stderr to avoid corrupting MCP protocol
		logger.SetOutput(os.Stderr)
	}

	return logger, nil
}

// addCommonFlags adds flags common to multiple commands
func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().String("transport-host", DefaultBindAddress, "Host to bind to")
	cmd.Flags().String("transport-port", DefaultBindPort, "Port to bind to")
	cmd.Flags().String("endpoint", DefaultEndPointPath, "Endpoint path for MCP")
}
