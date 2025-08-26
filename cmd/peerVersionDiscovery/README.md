# Peer Version Discovery Tool

A command-line tool for discovering and analyzing AvalancheGo versions across subnet validators.

## Overview

This tool connects to the Avalanche mainnet and discovers peers that are validating a specific subnet, then analyzes their AvalancheGo versions to help identify validators that may need software updates.

## Target Subnet

Currently configured to analyze subnet: `23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib`

## Features

- 🔍 **Peer Discovery**: Finds all peers connected to the Avalanche mainnet
- 🎯 **Subnet Filtering**: Identifies peers tracking the specified subnet
- 📊 **Version Analysis**: Analyzes version distribution across subnet validators
- 💡 **Recommendations**: Provides update recommendations for outdated validators
- 🌐 **Mainnet Focus**: Excludes primary networks (P/C/X) to focus on custom subnets

## Building

```bash
cd /path/to/avalanchego
go build -o cmd/peerVersionDiscovery/peerVersionDiscovery ./cmd/peerVersionDiscovery
```

## Usage

```bash
./cmd/peerVersionDiscovery/peerVersionDiscovery
```

## Output Example

```
🔍 Avalanche Peer Version Discovery Tool
📡 Target Subnet: 23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib
🌐 Network: Mainnet
⏰ Started at: 09:32:35 UTC

🔄 Connecting to Avalanche mainnet...
⏳ Discovering peers (this would normally take 30-60 seconds)...
✅ Total connected peers: 8
🎯 Peers tracking target subnet: 4

📊 Subnet Peer Version Report:
══════════════════════════════════════════════════════════════════════════════════════════════════════════════
Node ID                                       Version              Last Received        Subnets Tracked     
──────────────────────────────────────────────────────────────────────────────────────────────────────────────
NodeID-2C9sUz32xYBKxVU3fbcia1Eck5GhyoHru      avalanche/1.11.11    09:32:38             3                   
NodeID-4aAi4qLS31wQLcfFQnuqkytUdxcDQtfH6      avalanche/1.11.11    09:26:38             2                   
NodeID-6xBYegdq7VhUijrT9zCxwxYLXqweyFCCc      avalanche/1.11.11    09:20:38             2                   
NodeID-AXCoXTavjDLbHveFmn9ejRX8Nfwku3veZ      avalanche/1.10.19    09:11:38             2                   
══════════════════════════════════════════════════════════════════════════════════════════════════════════════

📈 Analysis for Subnet 23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib:

🏷️  Version Distribution:
   avalanche/1.11.11   : 3 peers (75.0%)
   avalanche/1.10.19   : 1 peers (25.0%)

🔝 Latest version: avalanche/1.11.11 (3 peers)
⚠️  Oldest version: avalanche/1.10.19 (1 peers)

💡 Recommendations:
   - 1 peers need to be updated from avalanche/1.10.19 to avalanche/1.11.11
   - Consider notifying subnet validators about version updates
```

## Current Status: Proof of Concept

**Important Note**: This is currently a proof of concept using mock data for demonstration purposes.

### What's Implemented
- ✅ Basic application structure
- ✅ Subnet ID parsing and validation
- ✅ Mock peer data generation
- ✅ Version analysis and reporting
- ✅ User-friendly output formatting

### What's Needed for Production Use

To connect to real Avalanche peers, the following components need to be implemented:

1. **Full Network Stack Initialization**
   - TLS certificate generation and management
   - Network listener setup
   - Message routing and handling

2. **Bootstrap Connection**
   - Connect to Avalanche bootstrap nodes
   - Implement peer discovery protocol
   - Handle network handshakes

3. **Real Peer Discovery**
   - Use actual Avalanche networking protocols
   - Query connected peers for their tracked subnets
   - Collect real version information

4. **Configuration Options**
   - Command-line flags for different subnets
   - Network selection (mainnet/fuji/local)
   - Connection timeout and retry logic

## Extending the Tool

To analyze different subnets, modify the `targetSubnet` constant in `main.go`:

```go
const (
    targetSubnet = "YOUR_SUBNET_ID_HERE"
)
```

## Architecture Notes

The application uses the existing AvalancheGo networking and peer management libraries:
- `github.com/ava-labs/avalanchego/network/peer` - Peer information structures
- `github.com/ava-labs/avalanchego/ids` - ID parsing and management  
- `github.com/ava-labs/avalanchego/utils/logging` - Logging infrastructure

## Future Improvements

- Add support for multiple subnet analysis
- Implement JSON output format
- Add historical version tracking
- Create web dashboard for visualization
- Add alerting for version mismatches