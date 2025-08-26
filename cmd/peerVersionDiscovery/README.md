# Peer Version Discovery Tool

A proof-of-concept tool to discover and analyze Avalanche subnet validator versions using P2P connections.

## Overview

This repository contains implementations for discovering Avalanche subnet validator versions:

1. **RPC-based discovery** - Uses the info API to query connected peers
2. **P2P-based discovery** - Connects directly to peers using the Avalanche gossip protocol
3. **Hybrid approach** - Combines P-Chain queries with P2P connections

## Files

- `main.go` - Basic RPC version that shows connected peers tracking the subnet
- `advanced_main.go` - Advanced version that queries P-Chain for all validators
- `p2p_discovery.go` - Full P2P implementation that connects via gossip protocol
- `simple_p2p.go` - Simplified P2P version using test peer utilities

## P2P Discovery Approach

The P2P approach works by:

1. **Creating TLS certificates** - Generate a node identity for handshaking
2. **Connecting to bootstrap nodes** - Start with known mainnet nodes
3. **Performing handshakes** - Exchange version information via the Avalanche protocol
4. **Requesting peer lists** - Get more peers tracking the target subnet
5. **Connecting to subnet validators** - Connect specifically to subnet validators

### Running P2P Discovery

```bash
# Simple P2P version (connects to bootstrap nodes)
go run ./cmd/peerVersionDiscovery/simple_p2p.go

# Full P2P version (attempts to discover subnet validators)
go run ./cmd/peerVersionDiscovery/p2p_discovery.go
```

### P2P Protocol Details

The Avalanche P2P protocol uses:
- **TLS 1.3** for secure connections
- **Ed25519** certificates for node identity
- **Protobuf** messages for communication
- **Port 9651** for P2P connections (9650 for RPC)

The handshake message includes:
- Network ID (mainnet = 1)  
- Node version (major.minor.patch)
- Tracked subnets
- Supported/objected ACPs

### Example P2P Code Structure

```go
// 1. Create TLS certificate
tlsCert, _ := staking.NewTLSCert()

// 2. Connect to peer
conn, _ := net.Dial("tcp", "peer-ip:9651")

// 3. Upgrade to TLS
tlsConn := tls.Client(conn, tlsConfig)

// 4. Send handshake
handshake := createHandshakeMessage(subnetID)
tlsConn.Write(handshake)

// 5. Parse response for version
response := readResponse(tlsConn)
version := parseVersion(response)
```

## Complete Solution

For a production implementation to get ALL subnet validator versions:

1. **Query P-Chain** for the complete validator set
2. **Use DHT/Kademlia** to discover peer IPs
3. **Connect via P2P** to each validator
4. **Handle reconnections** and retries
5. **Aggregate data** from multiple vantage points

## Running the Tools

### Prerequisites

- Go 1.21+
- Network access to Avalanche mainnet
- For RPC: Running node with info API enabled
- For P2P: Direct TCP access to port 9651

### Basic RPC Version

```bash
go run ./cmd/peerVersionDiscovery/main.go
```

### P2P Versions

```bash
# Simple P2P (test utility based)
go run ./cmd/peerVersionDiscovery/simple_p2p.go

# Full P2P implementation
go run ./cmd/peerVersionDiscovery/p2p_discovery.go
```

## Limitations

- P2P discovery requires direct network access (may be blocked by firewalls)
- Bootstrap nodes may not track all subnets
- Need to implement peer discovery to find subnet-specific validators
- Simplified message parsing (production would need full protobuf parsing)

## Future Improvements

1. Implement Kademlia DHT for peer discovery
2. Add GetPeerList message handling
3. Parse full protobuf responses properly
4. Cache discovered peers
5. Add retry logic for failed connections
6. Support multiple subnets in one run