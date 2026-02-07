// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vfcastr/Zabbix-MCP/pkg/client"
	"github.com/vfcastr/Zabbix-MCP/pkg/tools"
	"github.com/vfcastr/Zabbix-MCP/version"
)

const (
	DefaultBindAddress  = "127.0.0.1"
	DefaultBindPort     = "8080"
	DefaultEndPointPath = "/mcp"
)

func runHTTPServer(logger *log.Logger, host string, port string, endpointPath string) error {
	mcpServer := NewServer(version.Version, logger)
	tools.InitTools(mcpServer, logger)

	return httpServerInit(context.Background(), mcpServer, logger, host, port, endpointPath)
}

func httpServerInit(ctx context.Context, mcpServer *server.MCPServer, logger *log.Logger, host string, port string, endpointPath string) error {
	addr := fmt.Sprintf("%s:%s", host, port)
	logger.WithFields(log.Fields{
		"address":  addr,
		"endpoint": endpointPath,
	}).Info("Starting HTTP server")

	// Create HTTP server with streaming support
	httpServer := server.NewStreamableHTTPServer(mcpServer)

	// Create mux and add routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", client.HealthHandler(logger))

	// MCP endpoint
	mux.Handle(endpointPath, client.BuildMiddlewareStack(httpServer, logger))

	// Create the HTTP server
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Channel to receive server errors
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		logger.WithField("address", addr).Info("HTTP server listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		logger.WithField("signal", sig).Info("Received shutdown signal")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}

		logger.Info("Server shutdown complete")
	}

	return nil
}

func runStdioServer(logger *log.Logger) error {
	mcpServer := NewServer(version.Version, logger)
	tools.InitTools(mcpServer, logger)

	logger.Info("Starting Zabbix MCP server in stdio mode")
	return server.ServeStdio(mcpServer)
}

// NewServer creates a new MCP server instance
func NewServer(ver string, logger *log.Logger, opts ...server.ServerOption) *server.MCPServer {
	defaultOpts := []server.ServerOption{
		server.WithToolCapabilities(true),
	}

	allOpts := append(defaultOpts, opts...)

	mcpServer := server.NewMCPServer(
		"zabbix-mcp-server",
		ver,
		allOpts...,
	)

	logger.WithFields(log.Fields{
		"version": ver,
	}).Debug("Created new Zabbix MCP server")

	return mcpServer
}

func main() {
	rootCmd := &cobra.Command{
		Use:     "zabbix-mcp-server",
		Short:   "Zabbix MCP Server",
		Long:    `A Zabbix MCP server that provides tools for managing Zabbix hosts, items, triggers, templates, and maintenance.`,
		Version: version.Version,
	}

	stdioCmd := &cobra.Command{
		Use:   "stdio",
		Short: "Start stdio server (default)",
		Long:  `Start a server that communicates using stdio transport. This is the default mode for MCP clients like Claude for Desktop.`,
		Run: func(cmd *cobra.Command, _ []string) {
			logFile, err := rootCmd.PersistentFlags().GetString("log-file")
			if err != nil {
				stdlog.Fatal("Failed to get log file:", err)
			}
			logger, err := initLogger(logFile)
			if err != nil {
				stdlog.Fatal("Failed to initialize logger:", err)
			}
			if err := runStdioServer(logger); err != nil {
				logger.WithError(err).Fatal("Failed to run stdio server")
			}
		},
	}

	streamableHTTPCmd := &cobra.Command{
		Use:   "streamable-http",
		Short: "Start StreamableHTTP server",
		Long: `Start a server that communicates using the StreamableHTTP transport.
This mode allows clients to interact with the Zabbix MCP server over HTTP.
You can specify the host, port, and endpoint path to customize where the server listens.`,
		Run: func(cmd *cobra.Command, _ []string) {
			logFile, err := rootCmd.PersistentFlags().GetString("log-file")
			if err != nil {
				stdlog.Fatal("Failed to get log file:", err)
			}
			logger, err := initLogger(logFile)
			if err != nil {
				stdlog.Fatal("Failed to initialize logger:", err)
			}
			host, _ := cmd.Flags().GetString("transport-host")
			port, _ := cmd.Flags().GetString("transport-port")
			endpointPath := getEndpointPath(cmd)

			if err := runHTTPServer(logger, host, port, endpointPath); err != nil {
				logger.WithError(err).Fatal("Failed to run HTTP server")
			}
		},
	}

	// Legacy http command (deprecated)
	httpCmd := &cobra.Command{
		Use:        "http",
		Short:      "Start HTTP server (deprecated)",
		Long:       `This command is deprecated. Please use 'streamable-http' instead.`,
		Deprecated: "Use 'streamable-http' instead",
		Run: func(cmd *cobra.Command, args []string) {
			streamableHTTPCmd.Run(cmd, args)
		},
	}

	// Set default Run for rootCmd
	rootCmd.Run = func(cmd *cobra.Command, _ []string) {
		logFile, _ := cmd.PersistentFlags().GetString("log-file")
		logger, err := initLogger(logFile)
		if err != nil {
			stdlog.Fatal("Failed to initialize logger:", err)
		}

		// Check for HTTP mode from environment
		if shouldUseHTTPMode() {
			host := getHTTPHost()
			port := getHTTPPort()
			endpointPath := getEndpointPath(cmd)
			if err := runHTTPServer(logger, host, port, endpointPath); err != nil {
				logger.WithError(err).Fatal("Failed to run HTTP server")
			}
			return
		}

		// Default to stdio mode
		if err := runStdioServer(logger); err != nil {
			logger.WithError(err).Fatal("Failed to run stdio server")
		}
	}

	// Add persistent flags
	rootCmd.PersistentFlags().String("log-file", "", "Log file path (defaults to stderr)")

	// Add commands
	addCommonFlags(stdioCmd)
	addCommonFlags(streamableHTTPCmd)
	addCommonFlags(httpCmd)

	rootCmd.AddCommand(stdioCmd)
	rootCmd.AddCommand(streamableHTTPCmd)
	rootCmd.AddCommand(httpCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// shouldUseHTTPMode checks if environment variables indicate HTTP mode
func shouldUseHTTPMode() bool {
	mode := os.Getenv("TRANSPORT_MODE")
	return strings.ToLower(mode) == "http" || strings.ToLower(mode) == "streamable-http"
}

// getHTTPPort returns the port from environment variables or default
func getHTTPPort() string {
	if port := os.Getenv("TRANSPORT_PORT"); port != "" {
		return port
	}
	return DefaultBindPort
}

// getHTTPHost returns the host from environment variables or default
func getHTTPHost() string {
	if host := os.Getenv("TRANSPORT_HOST"); host != "" {
		return host
	}
	return DefaultBindAddress
}

// getEndpointPath returns the endpoint path from flag, environment, or default
func getEndpointPath(cmd *cobra.Command) string {
	if cmd != nil {
		if endpoint, _ := cmd.Flags().GetString("endpoint"); endpoint != "" {
			return path.Clean("/" + endpoint)
		}
	}
	if endpoint := os.Getenv("MCP_ENDPOINT"); endpoint != "" {
		return path.Clean("/" + endpoint)
	}
	return DefaultEndPointPath
}
