# Copyright vfcastr 2025
# SPDX-License-Identifier: MPL-2.0

# This Dockerfile contains multiple targets.
# Use 'docker build --target=<name> .' to build one.

# ===================================
#
#   Non-release images.
#
# ===================================

# devbuild compiles the binary
# -----------------------------------
FROM golang:1.24-alpine AS devbuild
ARG VERSION="1.0.0"

# Set the working directory
WORKDIR /build

# Install git for version info
RUN apk add --no-cache git ca-certificates

RUN go env -w GOMODCACHE=/root/.cache/go-build

# Copy go.mod first (go.sum may not exist yet)
COPY go.mod ./

# Download dependencies and generate go.sum
RUN --mount=type=cache,target=/root/.cache/go-build go mod download || true

COPY . ./

# Ensure dependencies are downloaded and tidy
RUN --mount=type=cache,target=/root/.cache/go-build go mod tidy

# Build the server
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build \
    -ldflags="-s -w -X github.com/vfcastr/Zabbix-MCP/version.Version=${VERSION} -X github.com/vfcastr/Zabbix-MCP/version.GitCommit=$(git rev-parse HEAD 2>/dev/null || echo 'unknown') -X github.com/vfcastr/Zabbix-MCP/version.BuildDate=$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
    -o zabbix-mcp-server ./cmd/zabbix-mcp-server

# dev runs the binary from devbuild
# -----------------------------------
FROM scratch AS dev
ARG VERSION="1.0.0"

# Set the working directory
WORKDIR /server

# Copy the binary from the build stage
COPY --from=devbuild /build/zabbix-mcp-server .
COPY --from=devbuild /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=devbuild /build/zabbix_documentation.md .

# Command to run the server (stdio mode is default)
CMD ["./zabbix-mcp-server", "stdio"]

# ===================================
#
#   Release images
#
# ===================================

# release image for production
# -----------------------------------
FROM scratch AS release
ARG VERSION

ENV VERSION=$VERSION

WORKDIR /server

COPY --from=devbuild /build/zabbix-mcp-server /bin/zabbix-mcp-server
COPY --from=devbuild /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

LABEL version=$VERSION
LABEL description="Zabbix MCP Server for Zabbix 7.0 LTS"
LABEL maintainer="vfcastr"

CMD ["/bin/zabbix-mcp-server", "stdio"]

# ===================================
#
#   Set default target to 'dev'.
#
# ===================================
FROM dev
