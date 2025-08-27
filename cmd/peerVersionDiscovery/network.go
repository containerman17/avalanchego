package main

import (
	"fmt"
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/bloom"
	"github.com/ava-labs/avalanchego/utils/ips"
	"github.com/ava-labs/avalanchego/utils/set"
)

// discoveryNetwork implements the Network interface for peer discovery
type discoveryNetwork struct {
	mu         sync.RWMutex
	knownPeers map[ids.NodeID]*ips.ClaimedIPPort
	peerFilter *bloom.Filter
	peerSalt   []byte
	maxPeers   int
}

func newDiscoveryNetwork() *discoveryNetwork {
	filter, _ := bloom.New(3, 256) // 3 hash functions, 256 bytes (2048 bits)
	return &discoveryNetwork{
		knownPeers: make(map[ids.NodeID]*ips.ClaimedIPPort),
		peerFilter: filter,
		peerSalt:   []byte("discovery"),
		maxPeers:   100,
	}
}

func (n *discoveryNetwork) Connected(peerID ids.NodeID) {
	fmt.Printf("  Network: Connected to %s\n", peerID)
}

func (n *discoveryNetwork) AllowConnection(peerID ids.NodeID) bool {
	return true
}

func (n *discoveryNetwork) Track(peers []*ips.ClaimedIPPort) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, peer := range peers {
		nodeID := ids.NodeIDFromCert(peer.Cert)
		if _, exists := n.knownPeers[nodeID]; !exists && len(n.knownPeers) < n.maxPeers {
			n.knownPeers[nodeID] = peer
			bloom.Add(n.peerFilter, nodeID[:], n.peerSalt)
			fmt.Printf("  Network: Tracked new peer %s at %s\n", nodeID, peer.AddrPort)
		}
	}

	fmt.Printf("  Network: Now tracking %d peers\n", len(n.knownPeers))
	return nil
}

func (n *discoveryNetwork) Disconnected(peerID ids.NodeID) {
	fmt.Printf("  Network: Disconnected from %s\n", peerID)
}

func (n *discoveryNetwork) KnownPeers() ([]byte, []byte) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.peerFilter.Marshal(), n.peerSalt
}

func (n *discoveryNetwork) Peers(
	peerID ids.NodeID,
	trackedSubnets set.Set[ids.ID],
	requestAllPeers bool,
	knownPeers *bloom.ReadFilter,
	peerSalt []byte,
) []*ips.ClaimedIPPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Return peers that the requester doesn't know about
	result := make([]*ips.ClaimedIPPort, 0)
	for nodeID, peer := range n.knownPeers {
		if nodeID == peerID {
			continue // Don't send the peer info about itself
		}

		// Check if peer already knows about this one
		if knownPeers != nil {
			known := bloom.Contains(knownPeers, nodeID[:], peerSalt)
			if known {
				continue
			}
		}

		result = append(result, peer)
		if len(result) >= 10 { // Limit response size
			break
		}
	}

	fmt.Printf("  Network: Returning %d peers to %s\n", len(result), peerID)
	return result
}
