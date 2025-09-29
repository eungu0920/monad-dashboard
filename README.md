# Monad Dashboard

A comprehensive real-time monitoring dashboard for Monad blockchain nodes, inspired by Firedancer's architecture.

![Monad Dashboard](./screenshot.png)

## Features

- **Real-time Metrics**: Live monitoring of node performance, transaction throughput, and network status
- **Transaction Waterfall**: Visual representation of transaction flow through the Monad pipeline
- **Performance Analytics**: EVM execution metrics, parallel processing success rates, and gas usage
- **Network Monitoring**: Peer connections, block propagation, and consensus participation
- **Modern UI**: React-based responsive interface with real-time WebSocket updates
- **Single Binary**: Embedded frontend assets for easy deployment

## Architecture

```
┌─────────────────┐    WebSocket/HTTP   ┌──────────────────┐
│  Web Frontend   │ ◄─────────────────► │ Dashboard Server │
│  (React/TS)     │                     │  (Go/Gin)        │
└─────────────────┘                     └──────────────────┘
                                                │
                                                │ IPC/Unix Socket
                                                ▼
                                        ┌──────────────────┐
                                        │   Monad Node     │
                                        │  (BFT + Exec)    │
                                        └──────────────────┘
```

## Transaction Pipeline Visualization

The dashboard visualizes Monad's transaction processing pipeline:

```
RPC → Mempool → Signature Verify → Nonce Dedup → EVM Execution → BFT Consensus → State Persistence → Broadcast
```

Each stage shows:
- Throughput (transactions per second)
- Drop rates and failure reasons
- Parallel execution success rates
- Processing latencies

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- A running Monad BFT and Execution node

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd monad-dashboard

# Install dependencies
make install

# Build and run
make run
```

The dashboard will be available at `http://localhost:8080`

### Development Mode

For development with hot-reload:

```bash
# Terminal 1: Start frontend dev server
make dev

# Terminal 2: Start backend dev server
make backend-dev
```

- Frontend dev server: `http://localhost:5173`
- Backend dev server: `http://localhost:8080`

## Build Commands

```bash
make install    # Install all dependencies
make build      # Build complete application
make run        # Build and run application
make dev        # Start frontend development server
make clean      # Clean build artifacts
make help       # Show all available commands
```

## Configuration

Create a `config.toml` file to customize the dashboard:

```toml
[server]
bind_address = "127.0.0.1"
port = 8080

[monad]
bft_ipc_path = "/tmp/monad-bft.sock"
execution_rpc_url = "http://127.0.0.1:8545"
node_name = "monad-validator-01"

[metrics]
collection_interval = 1000  # milliseconds
history_retention = 3600   # seconds
```

## API Endpoints

### REST API
- `GET /api/v1/health` - Health check
- `GET /api/v1/metrics` - Current node metrics
- `GET /api/v1/waterfall` - Transaction pipeline data

### WebSocket
- `GET /ws` - Real-time metrics stream

## Metrics Overview

### Node Information
- Version, Chain ID, Node Name
- Uptime, Status
- Connection status

### Transaction Metrics
- **Throughput**: Real-time TPS
- **Parallel Execution**: Success rate of parallel EVM processing
- **Mempool**: Current transaction pool size
- **Gas Usage**: Average gas price and consumption

### Consensus Metrics
- **Block Height**: Current blockchain height
- **Block Time**: Average time between blocks
- **Validator Info**: Count and participation rates
- **Network**: Peer connections and latency

### Execution Metrics
- **EVM Performance**: Execution times and success rates
- **State Management**: State size and update frequency
- **Conflict Resolution**: Parallel execution conflict rates

## Development

### Frontend Architecture

- **React 18** with TypeScript
- **Vite** for fast development and building
- **Recharts** for performance charts
- **Custom WebSocket hooks** for real-time data
- **CSS custom properties** for theming

### Backend Architecture

- **Gin framework** for HTTP server
- **Gorilla WebSocket** for real-time communication
- **Embedded assets** using Go's embed package
- **Modular metrics collection** system

### Adding New Metrics

1. Update the `MonadMetrics` struct in `backend/metrics.go`
2. Add corresponding TypeScript types in `frontend/src/types/index.ts`
3. Update the UI components to display new metrics
4. Add data collection from Monad components

## Deployment

### Single Binary Deployment

```bash
# Build the complete application
make build

# Copy to target server
scp ./monad-dashboard user@server:/opt/monad/
ssh user@server chmod +x /opt/monad/monad-dashboard

# Run on server
./monad-dashboard
```

### Systemd Service

```ini
[Unit]
Description=Monad Dashboard
After=network.target

[Service]
Type=simple
User=monad
WorkingDirectory=/opt/monad
ExecStart=/opt/monad/monad-dashboard
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-metric`
3. Make changes and test locally: `make dev`
4. Build and test: `make build && make run`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by [Firedancer's GUI architecture](https://github.com/firedancer-io/firedancer)
- Built for the Monad blockchain ecosystem
- Uses modern web technologies for optimal performance