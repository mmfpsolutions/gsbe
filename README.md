# GSBE - GoSlimBlockExplorer

A lightweight, self-hosted block explorer for SHA256d cryptocurrency nodes. Browse blocks, view transactions, and monitor mempool activity through a clean web interface — no database, no external dependencies, no CLI commands.

## Features

- Browse recent blocks with height, hash, difficulty, and transaction count
- View full block details with complete transaction lists
- Inspect transaction inputs, outputs, values, and addresses
- Coinbase text decoding (hex to ASCII) — see pool identifiers like "GoSlimStratum"
- Mempool statistics (transaction count, size, fees)
- Search by block height or block hash
- Multi-node support — configure and switch between multiple coin nodes
- SegWit-aware — displays witness commitments and witness data

## Requirements

- One or more SHA256d cryptocurrency nodes with REST API enabled (Bitcoin, DigiByte, Bitcoin Cash, etc.)
- Docker (recommended) or Go 1.24+
- No database required
- No `txindex=1` required
- Works with pruned nodes

## Quick Start

### Docker (Recommended)

```bash
docker run -d \
  --name gsbe \
  -p 3007:3007 \
  -v ./config:/app/config \
  -v ./logs:/app/logs \
  ghcr.io/mmfpsolutions/gsbe:latest
```

Open `http://localhost:3007` and add your node connections on the config page.

### Docker Compose

```yaml
services:
  gsbe:
    image: ghcr.io/mmfpsolutions/gsbe:latest
    container_name: gsbe
    ports:
      - "3007:3007"
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
    restart: unless-stopped
```

### Build from Source

```bash
git clone https://github.com/mmfpsolutions/gsbe.git
cd gsbe
npm install
npm run build:css
go build -o gsbe ./cmd/server
./gsbe
```

### Local Docker Build

```bash
./build-local.sh
docker run -d -p 3007:3007 -v ./config:/app/config -v ./logs:/app/logs gsbe:local
```

## Configuration

GSBE uses a JSON configuration file at `config/config.json`. A default config is created on first run.

```json
{
  "port": 3007,
  "title": "Block Explorer",
  "nodes": [
    {
      "id": "dgb_main",
      "name": "DigiByte Mainnet",
      "symbol": "DGB",
      "host": "192.168.1.100",
      "port": 14022,
      "network": "mainnet",
      "rest_enabled": true
    }
  ],
  "logging": {
    "level": "info",
    "logToFile": true,
    "logFilePath": "/app/logs/gsbe.log"
  }
}
```

### Node REST API

GSBE communicates with blockchain nodes via the REST API (same port as RPC, no authentication required). Ensure your node has REST enabled:

**Bitcoin/DigiByte/BCH** — add to your node config:
```
rest=1
```

## Supported Coins

Any SHA256d-based cryptocurrency with a Bitcoin-compatible REST API:

- Bitcoin (BTC)
- DigiByte (DGB)
- Bitcoin Cash (BCH)
- Bitcoin II (BC2)
- And other Bitcoin-derived coins

Works on mainnet, testnet, and regtest networks.

## Architecture

- **Backend for Frontend** — Go server proxies REST calls to configured nodes
- **No database** — all data fetched live from nodes
- **Single binary** — templates and static assets embedded via `go:embed`
- **No authentication** — open access, designed for local/private network use
- **No auto-polling** — manual refresh to avoid excessive node REST calls

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/status` | Version, uptime, memory |
| GET | `/api/v1/nodes` | List configured nodes |
| GET | `/api/v1/{node}/chain` | Blockchain info |
| GET | `/api/v1/{node}/blocks/recent?count=10` | Recent block headers |
| GET | `/api/v1/{node}/block/{hashOrHeight}` | Full block detail |
| GET | `/api/v1/{node}/tx/{txid}?blockhash={hash}` | Transaction detail |
| GET | `/api/v1/{node}/mempool` | Mempool statistics |
| GET | `/api/v1/{node}/search?q={query}` | Search by height or hash |

## Tech Stack

- Go 1.24+
- chi/v5 HTTP router
- Tailwind CSS
- Vanilla JavaScript
- Docker (Alpine-based, multi-arch: amd64/arm64)

## Limitations

- **No transaction search by txid** — without `txindex=1`, transactions can only be viewed from within their block. Navigate to a block first, then click a transaction.
- **Pruned node range** — pruned nodes can only serve blocks within their pruning window. Older blocks will show as unavailable.
- **No address balance lookups** — GSBE is a block browser, not a full address indexer.

## License

See [LICENSE](LICENSE) for details.

## Credits

Built by [MMFP Solutions](https://mmfpsolutions.io)
