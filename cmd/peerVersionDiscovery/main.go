package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ava-labs/avalanchego/genesis"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/message"
	"github.com/ava-labs/avalanchego/proto/pb/p2p"
	"github.com/ava-labs/avalanchego/snow/networking/router"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/ips"
	"github.com/ava-labs/avalanchego/utils/set"
)

var (
	targetSubnet  = flag.String("subnet", "", "Subnet ID to search for")
	targetNode    = flag.String("node", "", "Node ID to search for")
	interactive   = flag.Bool("interactive", false, "Run in interactive mode")
	listenPort    = flag.Uint("listen", 0, "Port to listen on for incoming connections (0 = disabled)")
	maxPeers      = flag.Int("max-peers", 200, "Maximum number of peers to discover")
	discoveryTime = flag.Duration("time", 2*time.Minute, "Time to spend discovering peers")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Create peer database
	peerDB := NewSearchablePeerDB()
	network := newSmartDiscoveryNetwork(peerDB)

	// Start listener if requested
	if *listenPort > 0 {
		if err := StartListener(uint16(*listenPort), constants.MainnetID, peerDB); err != nil {
			log.Printf("Warning: Failed to start listener: %v", err)
		}
	}

	// Run discovery phase
	fmt.Println("Starting peer discovery...")
	runDiscovery(ctx, network, peerDB, *maxPeers, *discoveryTime)

	// Show results based on flags
	if *targetSubnet != "" {
		searchForSubnet(peerDB, *targetSubnet)
	} else if *targetNode != "" {
		searchForNode(peerDB, *targetNode)
	} else if *interactive {
		RunInteractiveMode(peerDB)
	} else {
		// Default: show summary
		showSummary(peerDB)
	}
}

func runDiscovery(ctx context.Context, network *smartDiscoveryNetwork, peerDB *SearchablePeerDB, maxPeers int, duration time.Duration) {
	networkID := constants.MainnetID

	// Create timeout context
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, duration)
	defer timeoutCancel()

	// Get bootstrap peers
	bootstrappers := genesis.SampleBootstrappers(networkID, 5)
	ipQueue := make([]netip.AddrPort, 0, len(bootstrappers))
	for _, b := range bootstrappers {
		ipQueue = append(ipQueue, b.IP)
	}

	seenIPs := set.Set[netip.AddrPort]{}
	discoveredCount := 0

	fmt.Printf("Discovering peers for %v...\n", duration)

	for len(ipQueue) > 0 && discoveredCount < maxPeers {
		select {
		case <-timeoutCtx.Done():
			fmt.Printf("\nDiscovery phase completed (timeout)\n")
			return
		default:
		}

		ip := ipQueue[0]
		ipQueue = ipQueue[1:]

		if seenIPs.Contains(ip) {
			continue
		}
		seenIPs.Add(ip)

		// Try to connect
		connectCtx, connectCancel := context.WithTimeout(ctx, 10*time.Second)
		handler := createDiscoveryHandler(peerDB, &ipQueue, seenIPs)

		p, err := startDiscoveryPeer(connectCtx, ip, networkID, network, handler)
		connectCancel()

		if err != nil {
			continue
		}

		discoveredCount++
		if discoveredCount%10 == 0 {
			fmt.Printf("Discovered %d peers so far...\n", discoveredCount)
		}

		// Give time for message exchange
		time.Sleep(2 * time.Second)

		// Request peer list
		p.StartSendGetPeerList()
		time.Sleep(5 * time.Second)

		p.StartClose()
		p.AwaitClosed(context.Background())
	}

	fmt.Printf("\nDiscovery completed. Discovered %d peers.\n", peerDB.GetPeerCount())
}

func createDiscoveryHandler(peerDB *SearchablePeerDB, ipQueue *[]netip.AddrPort, seenIPs set.Set[netip.AddrPort]) router.InboundHandler {
	return router.InboundHandlerFunc(func(ctx context.Context, msg message.InboundMessage) {
		defer msg.OnFinishedHandling()

		switch msg.Op() {
		case message.HandshakeOp:
			// This typically won't happen for outbound connections
			if hs, ok := msg.Message().(*p2p.Handshake); ok {
				handleHandshake(peerDB, msg.NodeID(), hs)
			}
		case message.PeerListOp:
			if pl, ok := msg.Message().(*p2p.PeerList); ok {
				for _, c := range pl.ClaimedIpPorts {
					if addr, ok := netip.AddrFromSlice(c.IpAddr); ok && c.IpPort > 0 {
						peerAddr := netip.AddrPortFrom(addr, uint16(c.IpPort))
						if !seenIPs.Contains(peerAddr) {
							*ipQueue = append(*ipQueue, peerAddr)
						}
					}
				}
			}
		}
	})
}

func handleHandshake(peerDB *SearchablePeerDB, nodeID ids.NodeID, hs *p2p.Handshake) {
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
		peerDB.AddPeer(nodeID, ip, version, trackedSubnets)
	}
}

func searchForSubnet(peerDB *SearchablePeerDB, subnetStr string) {
	// Try to parse as full subnet ID
	var subnetID ids.ID
	var err error
	if subnetID, err = ids.FromString(subnetStr); err != nil {
		fmt.Printf("Invalid subnet ID: %s\n", subnetStr)
		return
	}

	peers := peerDB.GetPeersTrackingSubnet(subnetID)
	fmt.Printf("\nFound %d peers tracking subnet %s:\n", len(peers), subnetID)

	for i, peer := range peers {
		fmt.Printf("\n%d. NodeID: %s\n", i+1, peer.NodeID)
		fmt.Printf("   IP: %s\n", peer.IP)
		fmt.Printf("   Version: %s\n", peer.Version)
	}
}

func searchForNode(peerDB *SearchablePeerDB, nodeStr string) {
	peers := peerDB.SearchByNodeID(nodeStr)

	if len(peers) == 0 {
		fmt.Printf("No peers found matching NodeID: %s\n", nodeStr)
		return
	}

	fmt.Printf("\nFound %d peer(s) matching '%s':\n", len(peers), nodeStr)
	for i, peer := range peers {
		fmt.Printf("\n%d. NodeID: %s\n", i+1, peer.NodeID)
		fmt.Printf("   IP: %s\n", peer.IP)
		fmt.Printf("   Version: %s\n", peer.Version)
		if peer.TrackedSubnets.Len() > 1 {
			fmt.Printf("   Tracking %d subnets:\n", peer.TrackedSubnets.Len()-1)
			for subnetID := range peer.TrackedSubnets {
				if subnetID != ids.Empty {
					fmt.Printf("     - %s\n", subnetID)
				}
			}
		}
	}
}

func showSummary(peerDB *SearchablePeerDB) {
	allPeers := peerDB.GetAllPeers()
	subnets := peerDB.GetAllSubnets()

	fmt.Printf("\n=== Discovery Summary ===\n")
	fmt.Printf("Total peers discovered: %d\n", len(allPeers))
	fmt.Printf("Total subnets discovered: %d\n", len(subnets))

	// Show version distribution
	versions := make(map[string]int)
	for _, peer := range allPeers {
		versions[peer.Version]++
	}

	fmt.Println("\nVersion distribution:")
	for version, count := range versions {
		fmt.Printf("  %s: %d peers\n", version, count)
	}

	// Show subnets with validators
	if len(subnets) > 0 {
		fmt.Println("\nSubnets discovered:")
		for subnetID, count := range subnets {
			fmt.Printf("  %s: %d validators\n", subnetID, count)
		}
	}

	fmt.Println("\nUse --interactive flag for search mode")
	fmt.Println("Use --subnet <ID> to search for specific subnet")
	fmt.Println("Use --node <ID> to search for specific node")
}

// smartDiscoveryNetwork extends the basic network with peer tracking
type smartDiscoveryNetwork struct {
	*discoveryNetwork
	peerDB *SearchablePeerDB
}

func newSmartDiscoveryNetwork(peerDB *SearchablePeerDB) *smartDiscoveryNetwork {
	return &smartDiscoveryNetwork{
		discoveryNetwork: newDiscoveryNetwork(),
		peerDB:           peerDB,
	}
}

func (n *smartDiscoveryNetwork) Track(peers []*ips.ClaimedIPPort) error {
	// Call parent implementation
	if err := n.discoveryNetwork.Track(peers); err != nil {
		return err
	}

	// Also track in our database (though we won't have subnet info here)
	for _, peer := range peers {
		nodeID := ids.NodeIDFromCert(peer.Cert)
		// We only have basic info from PeerList, not subnet tracking info
		n.peerDB.AddPeer(nodeID, peer.AddrPort, "", set.Set[ids.ID]{})
	}

	return nil
}
