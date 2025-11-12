# pvec - Proxmox VE Terminal Client

A terminal-based user interface (TUI) for managing Proxmox Virtual Environment VMs and containers. Built with Go and designed for efficient keyboard-driven workflows.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.23-blue.svg)
![Coverage](https://img.shields.io/badge/coverage-85.5%25-brightgreen.svg)

## Features

- 📊 **Real-time Monitoring**: Scrolling table view of all VMs and containers with color-coded status
- ⚡ **Quick Actions**: Start, shutdown, reboot, and stop VMs/CTs with function keys
- ⌨️ **Keyboard Shortcuts**: Fully keyboard-driven interface (Function Keys + letter shortcuts)
- 🔄 **Auto-refresh**: Configurable automatic refresh of VM/CT status
- 🎨 **Color-coded Display**: Visual indicators for running (green), stopped (red), and paused (yellow) states
- 🔒 **Secure**: Token-based authentication with optional TLS verification
- 📝 **Configuration**: JSON-based configuration file with CLI override
- 🐳 **Docker Support**: Ready-to-use Docker image

## Quick Start

### Installation

#### From Source

```bash
# Clone the repository
git clone https://github.com/tsupplis/pvec.git
cd pvec

# Build and install
make install

# Or just build
make build
```

#### Using Docker

```bash
# Build Docker image
make docker-build

# Run with configuration
make docker-run
```

### Configuration

Create a configuration file at `~/.pvecrc` (or specify with `-c` flag):

```json
{
  "api_url": "https://your-proxmox-server:8006",
  "token_id": "your-user@pam!your-token-name",
  "token_secret": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "refresh_interval": "5s",
  "skip_tls_verify": true
}
```

#### Configuration Options

- **api_url**: Your Proxmox VE server URL (include port, typically 8006)
- **token_id**: API token ID in format `user@realm!token-name`
- **token_secret**: API token secret (UUID format)
- **refresh_interval**: How often to refresh the VM list (e.g., "5s", "10s", "1m")
- **skip_tls_verify**: Set to `true` to skip TLS certificate verification (useful for self-signed certs)

#### Creating a Proxmox API Token

1. Log into Proxmox VE web interface
2. Navigate to Datacenter → Permissions → API Tokens
3. Click "Add" to create a new token
4. Choose user (e.g., `root@pam`) and token name
5. Uncheck "Privilege Separation" if you want full access
6. Copy the Token Secret (you'll only see it once!)

### Usage

```bash
# Run with default config (~/.pvecrc)
pvec

# Run with custom config file
pvec -c /path/to/config.json
pvec --config /path/to/config.json
```

## Keyboard Shortcuts

### Function Keys

- **F1** / **h**: Show help dialog
- **F2** / **c**: Edit configuration
- **F3** / **s**: Start selected VM/CT
- **F4** / **d**: Shutdown selected VM/CT (graceful)
- **F5** / **r**: Reboot selected VM/CT
- **F6** / **t**: Stop selected VM/CT (force)
- **F10** / **q**: Quit application

### Navigation

- **↑/↓**: Navigate through VM/CT list
- **PgUp/PgDn**: Scroll page up/down
- **Home/End**: Jump to first/last item

## Display

The main list shows the following information for each VM/CT:

| Column | Description |
|--------|-------------|
| VMID | Unique identifier for the VM or container |
| Name | VM/CT name |
| Type | `VM` (QEMU) or `CT` (LXC container) |
| Status | `running`, `stopped`, or `paused` (color-coded) |
| Node | Proxmox node hosting the VM/CT |
| CPU | CPU usage percentage (color warning at 80%+) |
| Memory | Memory usage / Total memory (color warning at 80%+) |
| Uptime | Time since last boot (days, hours, minutes) |

### Color Coding

- 🟢 **Green**: Running VMs/CTs
- 🔴 **Red**: Stopped VMs/CTs
- 🟡 **Yellow**: Paused VMs/CTs
- 🟠 **Orange**: High resource usage (CPU/Memory ≥80%)



## Troubleshooting

### TLS Certificate Errors

If you see TLS certificate errors:

1. Set `skip_tls_verify: true` in your config (recommended for self-signed certs)
2. Or add your Proxmox CA certificate to system trust store

### Connection Refused

- Verify Proxmox server is running and accessible
- Check firewall rules (port 8006)
- Ensure API token has correct permissions

### No VMs/CTs Displayed

- Verify token has read permissions on datacenter/cluster
- Check token is not restricted to specific resources
- Try with "root@pam" user token without privilege separation

### Permission Errors

API token needs at least these privileges:
- `VM.Audit` - View VMs
- `VM.PowerMgmt` - Start/stop VMs
- `Sys.Audit` - View cluster status

## Contributing

Contributions are welcome! Please see [docs/dev.md](docs/dev.md) for development guidelines.

## Documentation

- [Development Guide](docs/dev.md) - Architecture, building, testing, and contributing
- [Code Analysis Report](docs/code_analysis.md) - Code quality metrics
- [Color System](docs/colors.md) - UI color theme documentation

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## Acknowledgments

- Built with [tview](https://github.com/rivo/tview) - Terminal UI library
- Uses [tcell](https://github.com/gdamore/tcell) - Terminal handling
- Configuration via [viper](https://github.com/spf13/viper)
- Testing with [testify](https://github.com/stretchr/testify)
