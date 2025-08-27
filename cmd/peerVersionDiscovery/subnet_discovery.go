package main

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/message"
	"github.com/ava-labs/avalanchego/proto/pb/p2p"
	"github.com/ava-labs/avalanchego/snow/networking/router"
	"github.com/ava-labs/avalanchego/utils/set"
)

// PeerInfo stores information about a discovered peer
type PeerInfo struct {
	IP             netip.AddrPort
	NodeID         ids.NodeID
	Version        string
	TrackedSubnets set.Set[ids.ID]
}

// SubnetDiscovery helps discover peers tracking specific subnets
type SubnetDiscovery struct {
	mu    sync.RWMutex
	peers map[ids.NodeID]*PeerInfo
}

func NewSubnetDiscovery() *SubnetDiscovery {
	return &SubnetDiscovery{
		peers: make(map[ids.NodeID]*PeerInfo),
	}
}

func (sd *SubnetDiscovery) AddPeer(nodeID ids.NodeID, ip netip.AddrPort, version string, trackedSubnets set.Set[ids.ID]) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	sd.peers[nodeID] = &PeerInfo{
		IP:             ip,
		NodeID:         nodeID,
		Version:        version,
		TrackedSubnets: trackedSubnets,
	}
}

func (sd *SubnetDiscovery) GetPeersTrackingSubnet(subnetID ids.ID) []*PeerInfo {
	sd.mu.RLock()
	defer sd.mu.RUnlock()

	var result []*PeerInfo
	for _, peer := range sd.peers {
		if peer.TrackedSubnets.Contains(subnetID) {
			result = append(result, peer)
		}
	}
	return result
}

func (sd *SubnetDiscovery) GetAllPeers() []*PeerInfo {
	sd.mu.RLock()
	defer sd.mu.RUnlock()

	result := make([]*PeerInfo, 0, len(sd.peers))
	for _, peer := range sd.peers {
		result = append(result, peer)
	}
	return result
}

func (sd *SubnetDiscovery) GetPeerCount() int {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	return len(sd.peers)
}

// CreateHandshakeHandler creates a message handler that captures peer information
func CreateHandshakeHandler(sd *SubnetDiscovery) router.InboundHandler {
	return router.InboundHandlerFunc(func(_ context.Context, msg message.InboundMessage) {
		fmt.Printf("  Received message: %s\n", msg.Op())

		// Capture handshake information
		if msg.Op() == message.HandshakeOp {
			if hs, ok := msg.Message().(*p2p.Handshake); ok {
				// Extract node info from the message
				nodeID := msg.NodeID()

				// Extract version
				var version string
				if hs.Client != nil {
					version = fmt.Sprintf("%s/%d.%d.%d",
						hs.Client.Name,
						hs.Client.Major,
						hs.Client.Minor,
						hs.Client.Patch)
				}

				// Extract tracked subnets
				trackedSubnets := set.Set[ids.ID]{}
				// Always add primary network
				trackedSubnets.Add(ids.Empty)

				for _, subnetBytes := range hs.TrackedSubnets {
					if subnetID, err := ids.ToID(subnetBytes); err == nil {
						trackedSubnets.Add(subnetID)
					}
				}

				// Extract IP
				if addr, ok := netip.AddrFromSlice(hs.IpAddr); ok && hs.IpPort > 0 {
					ip := netip.AddrPortFrom(addr, uint16(hs.IpPort))
					sd.AddPeer(nodeID, ip, version, trackedSubnets)

					fmt.Printf("  Handshake from %s (version: %s)\n", nodeID, version)
					if trackedSubnets.Len() > 1 { // More than just primary network
						fmt.Printf("  Tracking %d subnets (including primary):\n", trackedSubnets.Len())
						for subnetID := range trackedSubnets {
							if subnetID != ids.Empty {
								fmt.Printf("    - %s\n", subnetID)
							}
						}
					}
				}
			}
		}

		msg.OnFinishedHandling()
	})
}
