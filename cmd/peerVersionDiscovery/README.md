# Peer Version Discovery Tool

A simple proof-of-concept tool to discover and analyze Avalanche subnet validator versions.

## Overview

This tool connects to an Avalanche node via RPC and discovers all peers that are tracking a specific subnet. It then displays their avalanchego versions and provides a summary of version distribution.

## Features

- Connects to local or remote Avalanche node via RPC API
- Discovers all peers tracking a specific subnet
- Displays validator node IDs, IPs, and versions
- Shows version distribution and identifies outdated nodes
- Works with mainnet

## Files

- `main.go` - Basic version that shows connected peers tracking the subnet
- `advanced_main.go` - Advanced version that queries P-Chain for all validators and shows which ones we're connected to

## Usage

### Prerequisites

You need to have an Avalanche node running with the info API enabled. By default, the tool connects to `127.0.0.1:9651`.

### Running the basic tool

From the avalanchego repository root:

```bash
go run ./cmd/peerVersionDiscovery/main.go
```

### Running the advanced tool

```bash
go run ./cmd/peerVersionDiscovery/advanced_main.go
```

Or with a custom node URL:

```bash
AVALANCHE_NODE_URL="http://your-node:9650" go run ./cmd/peerVersionDiscovery/advanced_main.go
```

### Example Output (Basic Version)

```
Avalanche Subnet Peer Version Discovery
======================================
Target Subnet: 23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib

Connecting to node at: http://127.0.0.1:9651
Connected to node: NodeID-xxx
Node version: avalanche/1.11.x

Fetching peer information...
Total peers: 250

Subnet 23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib Validators:
================================================================================
Node ID                                      IP                   Version
--------------------------------------------------------------------------------
NodeID-xxx                                   1.2.3.4:9651         avalanche/1.11.0
...

Version distribution:
  avalanche/1.11.0: 5 nodes
  avalanche/1.11.1: 10 nodes

WARNING: 5 nodes (33%) are running versions different from avalanche/1.11.1
```

### Example Output (Advanced Version)

```
Advanced Subnet Validator Discovery (with P-Chain)
=================================================
Target Subnet: 23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib

Querying P-Chain for subnet validators...

Found 20 validators for subnet 23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib

Validator Details:
====================================================================================================
Node ID                                      Status               IP                   Version
----------------------------------------------------------------------------------------------------
NodeID-xxx                                   Connected            1.2.3.4:9651         avalanche/1.11.0
NodeID-yyy                                   Not Connected        Unknown              Unknown
...

Total subnet validators: 20
Connected to: 12 (60%)
Not connected to: 8

Version distribution (connected validators only):
  avalanche/1.11.0: 5 nodes
  avalanche/1.11.1: 7 nodes
```

## Limitations

- This is a proof of concept - very simple and dirty implementation
- Only shows peers that your node is connected to (not all subnet validators)
- Requires access to a node with info API enabled
- Does not query validators directly, only through connected peers
- The advanced version shows all validators but can only get version info for connected ones

## Future Improvements

To get ALL subnet validators (not just connected peers):
1. Query the P-Chain to get the full validator set for the subnet ✓ (done in advanced version)
2. Implement direct P2P connections to all validators
3. Use the Avalanche network protocol to handshake and get version info
4. Add support for multiple subnets
5. Export data in structured format (JSON, CSV)
6. Aggregate data from multiple nodes to get more complete coverage