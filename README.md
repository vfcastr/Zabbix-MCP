<p align="center">
  <img src="https://img.shields.io/badge/Zabbix-7.0+-red?style=for-the-badge&logo=zabbix" alt="Zabbix 7.0+">
  <img src="https://img.shields.io/badge/Go-1.24+-blue?style=for-the-badge&logo=go" alt="Go 1.24+">
  <img src="https://img.shields.io/badge/MCP-Server-green?style=for-the-badge" alt="MCP Server">
  <img src="https://img.shields.io/badge/License-MPL--2.0-yellow?style=for-the-badge" alt="License MPL-2.0">
</p>

<h1 align="center">
  <img width="48" height="48" src="https://img.icons8.com/fluency/48/model-context-protocol.png" alt="model-context-protocol"/>
  Zabbix MCP Server
</h1>

<p align="center">
  <strong>Model Context Protocol (MCP) server implementation for Zabbix 7.0 LTS</strong>
</p>

<p align="center">
  <a href="#-features">Features</a> •
  <a href="#-prerequisites">Prerequisites</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-configuration">Configuration</a> •
  <a href="#-tools">Tools</a>
</p>

---

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io/introduction) server implementation for **Zabbix 7.0 LTS**. Provides comprehensive integration with Zabbix for managing hosts, items, triggers, templates, maintenance, problems, events, user macros, LLD rules, proxies, user management, and more.

> **🐳 Recommended**: We recommend running via Docker as it is the most tested deployment method.

> **Security Note:** The MCP server is intended for local use only. If using HTTP transport, configure `MCP_ALLOWED_ORIGINS` to restrict access to trusted origins.

## 🚀 Features

- **79 MCP Tools** covering the full Zabbix API
- Host, Item, Trigger, and Template Management
- Host Groups, Template Groups, and Proxy Groups
- Problem and Event Management with acknowledgement
- User, User Group, and User Role Management
- User Macro Management (Host and Global)
- Low-Level Discovery (LLD) Rules and Prototypes
- Proxy Management
- Audit Log Access
- Built-in Zabbix API Documentation Search
- Stdio and HTTP transports
- Session-based Zabbix client with structured logging

## 📋 Prerequisites

- Docker (recommended)
- Go 1.24+ (if building from source)
- Zabbix 7.0 LTS server
- Valid Zabbix API Token

## ⚡ Quick Start

### Build the Docker Image
```bash
docker build -t zabbix-mcp:latest .
```

### Run in HTTP Mode
```bash
docker run -p 8080:8080 --rm \
    -e ZABBIX_URL='http://your-zabbix-server/api_jsonrpc.php' \
    -e ZABBIX_TOKEN='your-api-token' \
    -e TRANSPORT_MODE='http' \
    zabbix-mcp:latest
```

### Run in Stdio Mode
```bash
docker run -i --rm \
    -e ZABBIX_URL='http://your-zabbix-server/api_jsonrpc.php' \
    -e ZABBIX_TOKEN='your-api-token' \
    zabbix-mcp:latest
```

## ⚙️ Configuration

### MCP Client Configuration

Configuration examples are provided in the repository for popular MCP clients:

#### Claude Code (`claude.json`)
```json
{
    "mcpServers": {
        "Zabbix": {
            "type": "http",
            "url": "YOUR-ZABBIX-MCP-SERVER-URL/mcp",
            "headers": {
                "X-Zabbix-URL": "YOUR-ZABBIX-URL",
                "X-Zabbix-Token": "YOUR-ZABBIX-TOKEN"
            }
        }
    }
}
```

#### GitHub Copilot (`github_copilot.json`)
```json
{
    "servers": {
        "Zabbix MCP Server": {
            "type": "http",
            "url": "YOUR-ZABBIX-MCP-SERVER-URL/mcp",
            "headers": {
                "X-Zabbix-URL": "YOUR-ZABBIX-URL/api_jsonrpc.php",
                "X-Zabbix-Token": "YOUR-ZABBIX-TOKEN"
            }
        }
    }
}
```

#### Google Gemini (`gemini.json`)
```json
{
  "name": "zabbix-mcp-server",
  "version": "1.0.0",
  "description": "Zabbix MCP Server for managing Zabbix 7.0 LTS",
  "publisher": "vfcastr",
  "mcp": {
    "command": "./zabbix-mcp-server",
    "args": ["stdio"],
    "env": {
      "ZABBIX_URL": "${ZABBIX_URL}",
      "ZABBIX_TOKEN": "${ZABBIX_TOKEN}"
    }
  }
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ZABBIX_URL` | Zabbix API URL | `http://127.0.0.1/api_jsonrpc.php` |
| `ZABBIX_TOKEN` | Zabbix API Token | (required) |
| `ZABBIX_SKIP_VERIFY` | Skip TLS verification | `false` |
| `TRANSPORT_MODE` | Transport mode (`http` or `stdio`) | `stdio` |
| `TRANSPORT_PORT` | HTTP port | `8080` |
| `LOG_LEVEL` | Log level | `info` |

## 🛠️ Tools

**Total: 79 Tools Included**

### 🖥️ Host Management
| Tool | Description |
|------|-------------|
| `get_hosts` | List hosts with filters |
| `create_host` | Create a new host |
| `update_host` | Update host properties |
| `delete_host` | Delete hosts |

### 📂 Host Group Management
| Tool | Description |
|------|-------------|
| `zabbix_get_host_groups` | List host groups |
| `zabbix_create_host_group` | Create host group |
| `zabbix_update_host_group` | Update host group |
| `zabbix_delete_host_group` | Delete host groups |

### 📊 Item Management
| Tool | Description |
|------|-------------|
| `get_items` | List monitoring items |
| `create_item` | Create item |
| `update_item` | Update item |
| `delete_item` | Delete items |
| `get_history` | Get historical item values |

### ⚡ Trigger Management
| Tool | Description |
|------|-------------|
| `get_triggers` | List triggers |
| `create_trigger` | Create trigger |
| `update_trigger` | Update trigger |
| `delete_trigger` | Delete triggers |

### 📋 Template Management
| Tool | Description |
|------|-------------|
| `get_templates` | List templates |
| `create_template` | Create template |
| `update_template` | Update template |
| `delete_template` | Delete templates |
| `link_template` | Link template to host |
| `unlink_template` | Unlink template from host |

### 📁 Template Group Management
| Tool | Description |
|------|-------------|
| `zabbix_get_template_groups` | List template groups |
| `zabbix_create_template_group` | Create template group |
| `zabbix_update_template_group` | Update template group |
| `zabbix_delete_template_group` | Delete template groups |

### 🛠️ Maintenance Management
| Tool | Description |
|------|-------------|
| `get_maintenance` | List maintenance windows |
| `create_maintenance` | Create maintenance window |
| `update_maintenance` | Update maintenance window |
| `delete_maintenance` | Delete maintenance windows |

### 🚨 Problem & Event Management
| Tool | Description |
|------|-------------|
| `get_problems` | Get current problems |
| `get_events` | Get events |
| `acknowledge_event` | Acknowledge/update events |
| `get_trends` | Get trend data |
| `get_alerts` | Get alerts |

### 👥 User Management
| Tool | Description |
|------|-------------|
| `get_users` | List users |
| `create_user` | Create user |
| `update_user` | Update user |
| `delete_user` | Delete users |

### 👥 User Group Management
| Tool | Description |
|------|-------------|
| `get_user_groups` | List user groups |
| `create_user_group` | Create user group |
| `update_user_group` | Update user group |
| `delete_user_group` | Delete user groups |

### 🎭 User Role Management
| Tool | Description |
|------|-------------|
| `get_user_roles` | List user roles |
| `create_user_role` | Create user role |
| `update_user_role` | Update user role |
| `delete_user_role` | Delete user roles |

### 🔢 Macro Management
#### Global Macros
| Tool | Description |
|------|-------------|
| `get_global_macros` | Get global macros |
| `create_global_macro` | Create global macro |
| `update_global_macro` | Update global macro |
| `delete_global_macro` | Delete global macro |

#### Host User Macros
| Tool | Description |
|------|-------------|
| `get_user_macros` | Get host macros |
| `create_user_macro` | Create host macro |
| `update_user_macro` | Update host macro |
| `delete_user_macro` | Delete host macro |

### 🌐 Proxy Management
| Tool | Description |
|------|-------------|
| `zabbix_get_proxies` | List proxies |
| `zabbix_create_proxy` | Create proxy |
| `zabbix_update_proxy` | Update proxy |
| `zabbix_delete_proxies` | Delete proxies |

### 🌐 Proxy Group Management
| Tool | Description |
|------|-------------|
| `zabbix_get_proxy_groups` | List proxy groups |
| `zabbix_create_proxy_group` | Create proxy group |
| `zabbix_update_proxy_group` | Update proxy group |
| `zabbix_delete_proxy_groups` | Delete proxy groups |

### 🔍 LLD (Low-Level Discovery)
| Tool | Description |
|------|-------------|
| `get_lld_rules` | Get discovery rules |
| `create_lld_rule` | Create LLD rule |
| `update_lld_rule` | Update LLD rule |
| `delete_lld_rule` | Delete LLD rules |
| `copy_lld_rule` | Copy LLD rule to hosts |

### 🧩 Prototypes
#### Item Prototypes
| Tool | Description |
|------|-------------|
| `get_item_prototypes` | Get item prototypes |
| `create_item_prototype` | Create item prototype |
| `update_item_prototype` | Update item prototype |
| `delete_item_prototype` | Delete item prototypes |

#### Trigger Prototypes
| Tool | Description |
|------|-------------|
| `get_trigger_prototypes` | Get trigger prototypes |
| `create_trigger_prototype` | Create trigger prototype |
| `update_trigger_prototype` | Update trigger prototype |
| `delete_trigger_prototype` | Delete trigger prototypes |

### 📜 Audit & Documentation
| Tool | Description |
|------|-------------|
| `get_audit_log` | Get audit log entries |
| `get_zabbix_docs` | Search Zabbix API documentation |

## 🏗️ Building from Source

```bash
git clone https://github.com/vfcastr/Zabbix-MCP.git
cd Zabbix-MCP
go build -ldflags="-X github.com/vfcastr/Zabbix-MCP/version.Version=1.0.0" -o zabbix-mcp-server ./cmd/zabbix-mcp-server
```

## 📂 Project Structure

```
Zabbix-MCP/
├── cmd/zabbix-mcp-server/     # Entry point
├── pkg/
│   ├── client/                # Zabbix API client
│   └── tools/                 # MCP tools (79 tools)
│       ├── hosts/             # Host management
│       ├── hostgroups/        # Host group management
│       ├── items/             # Item management
│       ├── triggers/          # Trigger management
│       ├── templates/         # Template management
│       ├── templategroups/    # Template group management
│       ├── maintenance/       # Maintenance windows
│       ├── problems/          # Problem management
│       ├── events/            # Event management
│       ├── trends/            # Trend data
│       ├── alerts/            # Alert management
│       ├── users/             # User management
│       ├── usergroups/        # User group management
│       ├── userroles/         # User role management
│       ├── macros/            # User macros (host & global)
│       ├── proxies/           # Proxy management
│       ├── proxygroups/       # Proxy group management
│       ├── lld/               # LLD rules
│       ├── itemprototypes/    # Item prototypes
│       ├── triggerprototypes/ # Trigger prototypes
│       ├── auditlog/          # Audit log
│       └── docs/              # Documentation tool
├── version/                   # Version info
├── claude.json                # Claude Code config example
├── github_copilot.json        # GitHub Copilot config example
├── gemini.json                # Google Gemini config example
├── zabbix_documentation.md    # Zabbix API docs
├── Dockerfile
├── go.mod
└── README.md
```

## 📝 License

MPL-2.0

## 👤 Author

Created by [vfcastr](https://github.com/vfcastr)
