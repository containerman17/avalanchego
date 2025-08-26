# P2P Peer Version Discovery - Implementation Summary

## What This Tool Does

This tool connects directly to Avalanche network peers using the P2P gossip protocol to discover their avalanchego versions, specifically for subnet validators.

## How It Works

### 1. Direct P2P Connection (Port 9651)
- Creates a TCP connection to Avalanche nodes on port 9651 (P2P port)
- Generates Ed25519 TLS certificates for node identity
- Performs TLS 1.3 handshake with the peer

### 2. Avalanche Protocol Handshake
- Sends an Avalanche handshake message containing:
  - Network ID (1 for mainnet)
  - Our version information
  - **Tracked subnets** (including the target subnet ID)
  - Timestamp and signatures

### 3. Version Exchange
- Peer responds with their handshake containing their avalanchego version
- Format: `avalanche/major.minor.patch` (e.g., `avalanche/1.11.12`)

### 4. Peer Discovery
- Can request `GetPeerList` to discover more peers
- Filter peers by those tracking the target subnet
- Connect to each subnet validator to get their version

## Running the Tool

```bash
# Direct P2P version
go run ./cmd/peerVersionDiscovery/direct_p2p.go

# Simple P2P using test utilities
go run ./cmd/peerVersionDiscovery/simple_p2p.go

# Full P2P with subnet discovery
go run ./cmd/peerVersionDiscovery/p2p_discovery.go
```

## Example Connection Flow

```
1. TCP Connect to peer:9651
   ↓
2. TLS Handshake (Ed25519 cert)
   ↓
3. Send Avalanche Handshake (with subnet ID)
   ↓
4. Receive Peer Handshake (contains version)
   ↓
5. Request GetPeerList
   ↓
6. Connect to subnet-specific peers
   ↓
7. Collect all versions
```

## Key Implementation Details

- **Protocol**: Avalanche P2P gossip protocol
- **Port**: 9651 (not 9650 which is for RPC)
- **Security**: TLS 1.3 with Ed25519 certificates
- **Messages**: Protobuf-encoded
- **Subnet Filtering**: Include subnet ID in handshake `trackedSubnets` field

## Requirements

- Network access to Avalanche mainnet nodes
- Ability to connect to TCP port 9651
- Go 1.21+ with avalanchego dependencies

## Actual Code That Connects

The core connection logic:

```go
// 1. Generate node identity
tlsCert, _ := staking.NewTLSCert()

// 2. TCP connection
conn, _ := net.Dial("tcp", "node-ip:9651")

// 3. TLS upgrade
tlsConfig := peer.TLSConfig(*tlsCert, nil)
clientUpgrader := peer.NewTLSClientUpgrader(tlsConfig, metrics)
peerID, tlsConn, _, _ := clientUpgrader.Upgrade(conn)

// 4. Send handshake with subnet ID
handshake, _ := messageCreator.Handshake(
    networkID,
    timestamp,
    myIP,
    version,
    major, minor, patch,
    timestamp,
    signature,
    blsSig,
    []ids.ID{targetSubnetID}, // Key: specify subnet we're interested in
    supportedACPs,
    objectedACPs,
    knownPeers,
    salt,
    false,
)
tlsConn.Write(handshake.Bytes())

// 5. Read response containing peer's version
response := tlsConn.Read(buffer)
// Parse response to extract version
```

This creates a direct P2P connection to Avalanche validators and retrieves their versions without needing RPC access.