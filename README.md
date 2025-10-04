# Zcrypt

A cryptographic log chain system with tamper-proof verification using Ed25519 signatures and SHA-256 hash chains.

## Overview

Zcrypt is a distributed logging system that maintains immutable audit trails through cryptographic signatures and blockchain-inspired hash linking. Each log entry is signed with Ed25519 and linked to the previous entry via SHA-256 hashes, making tampering mathematically detectable.

### Key Features

- **Cryptographic Signing**: Ed25519 digital signatures for each log entry
- **Tamper-Proof Chain**: SHA-256 hash linking ensures integrity
- **Local & Server Modes**: Maintain logs locally or sync to central server
- **Verification Tools**: Built-in chain integrity verification
- **Agent Management**: Register and track multiple logging agents
- **REST API**: Full HTTP API for server integration

## Architecture

```
┌─────────────┐         ┌──────────────┐
│   Agent     │         │    Server    │
│  (CLI Tool) │────────▶│  (REST API)  │
└─────────────┘         └──────────────┘
      │                        │
      ▼                        ▼
┌─────────────┐         ┌──────────────┐
│ Local Chain │         │ Server Chain │
│  ~/.zcrypt/ │         │ Centralized  │
└─────────────┘         └──────────────┘
```

## Installation

### Prerequisites

- Go 1.25 or higher

### Build from Source

```bash
git clone https://github.com/amshithnair/zcrypt.git
cd zcrypt

# Build the CLI agent
go build -o zcrypt ./agent

# Build the server
go build -o zcrypt-server ./server
```

### Install

```bash
# Install CLI globally
go install ./agent

# Or move to PATH
sudo mv zcrypt /usr/local/bin/
```

## Quick Start

### 1. Generate Keypair

```bash
zcrypt genkey
```

This creates:
- `zcrypt_private.key` - Your Ed25519 private key (keep secure!)
- `zcrypt_public.key` - Your Ed25519 public key

### 2. Create Local Log Entry

```bash
zcrypt log "System started successfully"
```

### 3. Verify Chain Integrity

```bash
zcrypt chain-verify
```

### 4. Start Server (Optional)

```bash
# In a separate terminal
./zcrypt-server
```

Server runs on `http://localhost:8080` by default.

### 5. Send Logs to Server

```bash
# Set server URL (optional, defaults to localhost:8080)
export ZCRYPT_SERVER="http://your-server:8080"

# Register agent
zcrypt register-agent my-agent "My Agent Name"

# Send log to server
zcrypt send-to-server "Deployed version 1.2.3"
```

## CLI Commands

### Local Operations

| Command | Description |
|---------|-------------|
| `zcrypt genkey` | Generate Ed25519 keypair |
| `zcrypt log "message"` | Sign and store log entry locally |
| `zcrypt verify "message" <signature>` | Verify a log signature |
| `zcrypt chain-verify` | Verify entire local chain integrity |
| `zcrypt chain-stats` | Display local chain statistics |
| `zcrypt chain-export` | Export local chain as JSON |

### Server Operations

| Command | Description |
|---------|-------------|
| `zcrypt send-to-server "message"` | Send log to central server |
| `zcrypt server-stats` | Get server statistics |
| `zcrypt server-verify` | Verify server chain integrity |
| `zcrypt register-agent <id> <name>` | Register agent with server |

## API Reference

### Endpoints

#### Health Check
```http
GET /api/v1/health
```

#### Submit Log
```http
POST /api/v1/logs
Content-Type: application/json

{
  "message": "Log message",
  "signature": "hex_encoded_signature",
  "pubkey": "hex_encoded_public_key",
  "agent_id": "agent_identifier",
  "metadata": {
    "user": "username",
    "hostname": "server01"
  }
}
```

#### Get Logs
```http
GET /api/v1/logs?limit=100&offset=0
```

#### Get Log by Index
```http
GET /api/v1/logs/:id
```

#### Get Logs by Time Range
```http
GET /api/v1/logs/range?start=2024-01-01T00:00:00Z&end=2024-12-31T23:59:59Z
```

#### Verify Chain
```http
POST /api/v1/verify/chain
```

#### Register Agent
```http
POST /api/v1/agents/register
Content-Type: application/json

{
  "agent_id": "my-agent",
  "pubkey": "hex_encoded_public_key",
  "name": "My Agent"
}
```

#### Get Statistics
```http
GET /api/v1/stats
```

## Configuration

### Environment Variables

- `ZCRYPT_SERVER` - Server URL (default: `http://localhost:8080`)
- `HOME` - User home directory for storing keys and chain data

### File Locations

- Keys: `./zcrypt_private.key`, `./zcrypt_public.key`
- Local chain: `~/.zcrypt/logs.chain`
- Server chain: `./server_logs.chain` (when running server)
- Exports: `./zcrypt_chain_export.json`

## How It Works

### Log Chain Structure

Each log entry contains:

```json
{
  "timestamp": "2024-10-04T12:00:00Z",
  "message": "Log message",
  "signature": "ed25519_signature_hex",
  "pubkey": "public_key_hex",
  "prev_hash": "sha256_of_previous_entry",
  "current_hash": "sha256_of_this_entry",
  "metadata": {}
}
```

### Hash Calculation

```
hash = SHA256(timestamp|message|signature|pubkey|prev_hash)
```

### Chain Verification

1. Verify each entry's signature using Ed25519
2. Recalculate each entry's hash
3. Verify hash links between entries
4. Check genesis entry has `prev_hash = "0"`

Any tampering breaks the chain and is immediately detected.

## Use Cases

- **Audit Logging**: Tamper-proof audit trails for compliance
- **System Monitoring**: Cryptographically signed system events
- **Security Events**: Immutable security incident logs
- **CI/CD Pipelines**: Verifiable deployment histories
- **IoT Device Logs**: Secure logging from distributed devices
- **Blockchain Applications**: Off-chain verifiable event logs

## Security Considerations

- **Private Key Protection**: Store `zcrypt_private.key` securely (0600 permissions)
- **Key Distribution**: Share public keys through secure channels
- **Server Security**: Use HTTPS in production environments
- **Backup Strategy**: Regularly backup chain files
- **Signature Verification**: Always verify signatures before trusting log entries

## Development

### Run Tests

```bash
go test ./crypto -v
```

### Project Structure

```
zcrypt/
├── agent/          # CLI client
│   └── main.go
├── server/         # REST API server
│   └── main.go
├── crypto/         # Core cryptography and chain logic
│   ├── chain.go
│   ├── chain_test.go
│   └── keys.go
├── utils/          # HTTP client utilities
│   └── client.go
├── go.mod
└── README.md
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Author

**Amshith Nair**
- GitHub: [@amshithnair](https://github.com/amshithnair)

## Acknowledgments

- Ed25519 implementation from Go's `crypto/ed25519`
- Web framework: [Fiber](https://github.com/gofiber/fiber)
- Inspired by blockchain and cryptographic audit logging systems
