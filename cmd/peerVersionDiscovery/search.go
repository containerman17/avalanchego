package main

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/message"
	"github.com/ava-labs/avalanchego/network/peer"
	"github.com/ava-labs/avalanchego/proto/pb/p2p"
	"github.com/ava-labs/avalanchego/snow/networking/router"
	"github.com/ava-labs/avalanchego/staking"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/prometheus/client_golang/prometheus"
)

// SearchablePeerDB extends SubnetDiscovery with search capabilities
type SearchablePeerDB struct {
	*SubnetDiscovery
}

func NewSearchablePeerDB() *SearchablePeerDB {
	return &SearchablePeerDB{
		SubnetDiscovery: NewSubnetDiscovery(),
	}
}

// SearchByNodeID finds a peer by partial or full NodeID
func (db *SearchablePeerDB) SearchByNodeID(nodeIDStr string) []*PeerInfo {
	db.mu.RLock()
	defer db.mu.RUnlock()

	nodeIDStr = strings.ToLower(nodeIDStr)
	var results []*PeerInfo

	for nodeID, peer := range db.peers {
		nodeIDString := nodeID.String()
		if strings.Contains(strings.ToLower(nodeIDString), nodeIDStr) {
			results = append(results, peer)
		}
	}

	return results
}

// SearchBySubnet finds all peers tracking a specific subnet (by ID or partial ID)
func (db *SearchablePeerDB) SearchBySubnet(subnetStr string) []*PeerInfo {
	db.mu.RLock()
	defer db.mu.RUnlock()

	subnetStr = strings.ToLower(subnetStr)
	var results []*PeerInfo

	// Try exact match first
	if subnetID, err := ids.FromString(subnetStr); err == nil {
		return db.GetPeersTrackingSubnet(subnetID)
	}

	// Otherwise do partial match
	for _, peer := range db.peers {
		for subnetID := range peer.TrackedSubnets {
			if subnetID != ids.Empty && strings.Contains(strings.ToLower(subnetID.String()), subnetStr) {
				results = append(results, peer)
				break
			}
		}
	}

	return results
}

// GetAllSubnets returns all unique subnets being tracked
func (db *SearchablePeerDB) GetAllSubnets() map[ids.ID]int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	subnets := make(map[ids.ID]int)
	for _, peer := range db.peers {
		for subnetID := range peer.TrackedSubnets {
			if subnetID != ids.Empty {
				subnets[subnetID]++
			}
		}
	}

	return subnets
}

// StartListener starts a P2P listener to accept incoming connections
func StartListener(port uint16, networkID uint32, peerDB *SearchablePeerDB) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	fmt.Printf("Started P2P listener on port %d\n", port)
	fmt.Println("Incoming peers will reveal their tracked subnets...")

	// Create TLS certificate
	tlsCert, err := staking.NewTLSCert()
	if err != nil {
		return err
	}

	tlsConfig := peer.TLSConfig(*tlsCert, nil)
	serverUpgrader := peer.NewTLSServerUpgrader(
		tlsConfig,
		prometheus.NewCounter(prometheus.CounterOpts{}),
	)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Printf("Error accepting connection: %v\n", err)
				continue
			}

			go handleIncomingPeer(conn, serverUpgrader, networkID, peerDB)
		}
	}()

	return nil
}

func handleIncomingPeer(conn net.Conn, upgrader peer.Upgrader, networkID uint32, peerDB *SearchablePeerDB) {
	// Upgrade connection
	peerID, conn, _, err := upgrader.Upgrade(conn)
	if err != nil {
		fmt.Printf("Failed to upgrade connection: %v\n", err)
		conn.Close()
		return
	}

	fmt.Printf("Incoming connection from %s\n", peerID)

	// Create handler to capture handshake
	_ = router.InboundHandlerFunc(func(ctx context.Context, msg message.InboundMessage) {
		if msg.Op() == message.HandshakeOp {
			if hs, ok := msg.Message().(*p2p.Handshake); ok {
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
				trackedSubnets.Add(ids.Empty) // Primary network

				for _, subnetBytes := range hs.TrackedSubnets {
					if subnetID, err := ids.ToID(subnetBytes); err == nil {
						trackedSubnets.Add(subnetID)
					}
				}

				// Extract IP
				if addr, ok := netip.AddrFromSlice(hs.IpAddr); ok && hs.IpPort > 0 {
					ip := netip.AddrPortFrom(addr, uint16(hs.IpPort))
					peerDB.AddPeer(peerID, ip, version, trackedSubnets)

					fmt.Printf("  Version: %s\n", version)
					if trackedSubnets.Len() > 1 {
						fmt.Printf("  Tracking %d subnets:\n", trackedSubnets.Len()-1)
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

	// Start peer (simplified - in real implementation would need full peer setup)
	// For now, just close the connection after receiving handshake
	conn.Close()
}
