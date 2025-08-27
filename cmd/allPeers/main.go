package main

import (
	"context"
	"crypto"
	"crypto/tls"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/genesis"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/message"
	"github.com/ava-labs/avalanchego/network/peer"
	"github.com/ava-labs/avalanchego/network/throttling"
	"github.com/ava-labs/avalanchego/proto/pb/p2p"
	"github.com/ava-labs/avalanchego/snow/networking/router"
	"github.com/ava-labs/avalanchego/snow/networking/tracker"
	"github.com/ava-labs/avalanchego/snow/uptime"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/avalanchego/staking"
	"github.com/ava-labs/avalanchego/upgrade"
	"github.com/ava-labs/avalanchego/utils"
	"github.com/ava-labs/avalanchego/utils/bloom"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto/bls/signer/localsigner"
	"github.com/ava-labs/avalanchego/utils/ips"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/math/meter"
	"github.com/ava-labs/avalanchego/utils/resource"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/version"
)

type DiscoveredPeer struct {
	NodeID         ids.NodeID
	IP             netip.AddrPort
	Version        string
	TrackedSubnets []ids.ID
	LastSeen       time.Time
}

// getOutboundIP attempts to get our external IP by dialing a well-known address
func getOutboundIP() netip.Addr {

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return netip.IPv4Unspecified()
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	if addr, ok := netip.AddrFromSlice(localAddr.IP); ok {
		return addr
	}
	return netip.IPv4Unspecified()
}

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
		peerSalt:   []byte("discovery i2yb2b323e3"),
		maxPeers:   3000,
	}
}

func (n *discoveryNetwork) Connected(peerID ids.NodeID) {
	fmt.Printf("  ✅ Network: Connected to %s\n", peerID)
}

func (n *discoveryNetwork) AllowConnection(peerID ids.NodeID) bool {
	return true
}

func (n *discoveryNetwork) Track(peers []*ips.ClaimedIPPort) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	newPeers := 0
	for _, peer := range peers {
		// fmt.Printf("peer %+v\n", peer)
		nodeID := ids.NodeIDFromCert(peer.Cert)
		if _, exists := n.knownPeers[nodeID]; !exists && len(n.knownPeers) < n.maxPeers {
			n.knownPeers[nodeID] = peer
			bloom.Add(n.peerFilter, nodeID[:], n.peerSalt)
			// fmt.Printf("  📍 Network: Tracked new peer %s at %s\n", nodeID, peer.AddrPort)
			newPeers++
		}
	}
	fmt.Printf("  📍 Network: Tracked %d new peers\n", newPeers)

	fmt.Printf("  📊 Network: Now tracking %d peers\n", len(n.knownPeers))
	return nil
}

func (n *discoveryNetwork) Disconnected(peerID ids.NodeID) {
	fmt.Printf("  ❌ Network: Disconnected from %s\n", peerID)
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

	fmt.Printf("  📤 Network: Returning %d peers to %s\n", len(result), peerID)
	return result
}

func startDiscoveryPeer(
	ctx context.Context,
	remoteIP netip.AddrPort,
	networkID uint32,
	network peer.Network,
	msgHandler router.InboundHandler,
	tlsCert *tls.Certificate,
) (peer.Peer, error) {
	// Connect to remote peer
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, constants.NetworkType, remoteIP.String())
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Setup TLS
	tlsConfig := peer.TLSConfig(*tlsCert, nil)
	clientUpgrader := peer.NewTLSClientUpgrader(
		tlsConfig,
		prometheus.NewCounter(prometheus.CounterOpts{}),
	)

	peerID, conn, cert, err := clientUpgrader.Upgrade(conn)
	if err != nil {
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	fmt.Printf("  🔐 TLS handshake successful with NodeID: %s\n", peerID)

	// Create message creator
	mc, err := message.NewCreator(
		prometheus.NewRegistry(),
		constants.DefaultNetworkCompressionType,
		10*time.Second,
	)
	if err != nil {
		return nil, err
	}

	// Create metrics
	metrics, err := peer.NewMetrics(prometheus.NewRegistry())
	if err != nil {
		return nil, err
	}

	// Create resource tracker
	resourceTracker, err := tracker.NewResourceTracker(
		prometheus.NewRegistry(),
		resource.NoUsage,
		meter.ContinuousFactory{},
		10*time.Second,
	)
	if err != nil {
		return nil, err
	}

	// Setup keys
	tlsKey := tlsCert.PrivateKey.(crypto.Signer)
	blsKey, err := localsigner.New()
	if err != nil {
		return nil, err
	}

	// Try to get our external IP (fallback to hardcoded if detection fails)
	ourIP := getOutboundIP()
	if ourIP == netip.IPv4Unspecified() {
		// Fallback to your EC2 instance IP
		ourIP = netip.MustParseAddr("54.95.191.28")
	}
	ourPort := uint16(9651) // Default Avalanche port

	parsedCert, err := staking.ParseCertificate(tlsCert.Leaf.Raw)
	if err != nil {
		return nil, err
	}
	myNodeID := ids.NodeIDFromCert(parsedCert)

	// Parse the subnet IDs we want to track
	subnet1, _ := ids.FromString("h7egyVb6fKHMDpVaEsTEcy7YaEnXrayxZS4A1AEU4pyBzmwGp")
	subnet2, _ := ids.FromString("nQCwF6V9y8VFjvMuPeQVWWYn6ba75518Dpf6ZMWZNb3NyTA94")

	// Create set of tracked subnets
	trackedSubnets := set.Set[ids.ID]{}
	trackedSubnets.Add(subnet1)
	trackedSubnets.Add(subnet2)

	// Create peer configuration
	config := &peer.Config{
		Metrics:              metrics,
		MessageCreator:       mc,
		Log:                  logging.NoLog{},
		InboundMsgThrottler:  throttling.NewNoInboundThrottler(),
		Network:              network,
		Router:               msgHandler,
		VersionCompatibility: version.GetCompatibility(upgrade.InitiallyActiveTime),
		MyNodeID:             myNodeID,
		MySubnets:            trackedSubnets, // Set our tracked subnets here
		Beacons:              validators.NewManager(),
		Validators:           validators.NewManager(),
		NetworkID:            networkID,
		PingFrequency:        constants.DefaultPingFrequency,
		PongTimeout:          constants.DefaultPingPongTimeout,
		MaxClockDifference:   time.Minute,
		ResourceTracker:      resourceTracker,
		UptimeCalculator:     uptime.NoOpCalculator,
		IPSigner: peer.NewIPSigner(
			utils.NewAtomic(netip.AddrPortFrom(ourIP, ourPort)),
			tlsKey,
			blsKey,
		),
	}

	// Create and start peer
	p := peer.Start(
		config,
		conn,
		cert,
		peerID,
		peer.NewBlockingMessageQueue(
			metrics,
			logging.NoLog{},
			1024,
		),
		false,
	)

	// Wait for peer to be ready
	if err := p.AwaitReady(ctx); err != nil {
		return nil, fmt.Errorf("peer failed to become ready: %w", err)
	}

	return p, nil
}

func main() {
	fmt.Println("🚀 Starting minimal peer discovery...")
	fmt.Println("================================================")

	// Create or load certificate
	certPath := filepath.Join("/tmp", "avalanche_cert.pem")
	keyPath := filepath.Join("/tmp", "avalanche_key.pem")

	var tlsCert *tls.Certificate
	var stakingCert, stakingKey []byte

	// Try to load existing certificate
	if _, err := os.Stat(certPath); err == nil {
		fmt.Println("📂 Loading existing certificate from /tmp/...")
		stakingCert, err = os.ReadFile(certPath)
		if err == nil {
			stakingKey, err = os.ReadFile(keyPath)
			if err == nil {
				tlsCert, err = staking.LoadTLSCertFromBytes(stakingCert, stakingKey)
				if err != nil {
					fmt.Printf("⚠️  Failed to load existing cert, creating new one: %v\n", err)
					tlsCert = nil
				}
			}
		}
	}

	// Create new certificate if needed
	if tlsCert == nil {
		fmt.Println("🔐 Creating new certificate...")
		var err error
		stakingKey, stakingCert, err = staking.NewCertAndKeyBytes()
		if err != nil {
			fmt.Printf("❌ Failed to create staking cert: %v\n", err)
			return
		}

		// Save to /tmp for future use
		if err := os.WriteFile(certPath, stakingCert, 0600); err != nil {
			fmt.Printf("⚠️  Failed to save cert: %v\n", err)
		}
		if err := os.WriteFile(keyPath, stakingKey, 0600); err != nil {
			fmt.Printf("⚠️  Failed to save key: %v\n", err)
		}

		tlsCert, err = staking.LoadTLSCertFromBytes(stakingCert, stakingKey)
		if err != nil {
			fmt.Printf("❌ Failed to load TLS cert: %v\n", err)
			return
		}
	}

	// Print our NodeID
	parsedCert, err := staking.ParseCertificate(tlsCert.Leaf.Raw)
	if err != nil {
		fmt.Printf("❌ Failed to parse cert: %v\n", err)
		return
	}
	myNodeID := ids.NodeIDFromCert(parsedCert)
	fmt.Printf("🔑 Our consistent NodeID: %s\n", myNodeID)

	// Print the subnets we're tracking
	subnet1, _ := ids.FromString("h7egyVb6fKHMDpVaEsTEcy7YaEnXrayxZS4A1AEU4pyBzmwGp")
	subnet2, _ := ids.FromString("nQCwF6V9y8VFjvMuPeQVWWYn6ba75518Dpf6ZMWZNb3NyTA94")
	fmt.Printf("🌐 We are tracking subnets:\n")
	fmt.Printf("   - %s\n", subnet1)
	fmt.Printf("   - %s\n", subnet2)

	// Track discovered peers
	discoveredPeers := make(map[ids.NodeID]*DiscoveredPeer)
	var discoveredMutex sync.Mutex

	ipQueue := []netip.AddrPort{}
	seenIPs := set.Set[netip.AddrPort]{}

	// Create network implementation
	network := newDiscoveryNetwork()

	// Get bootstrap nodes
	fmt.Println("📡 Loading bootstrap nodes...")
	bootstrappers := genesis.SampleBootstrappers(constants.MainnetID, 20) // Get Fuji testnet bootstrappers
	for _, b := range bootstrappers {
		ipQueue = append(ipQueue, b.IP)
		fmt.Printf("  📍 Bootstrap: %s at %s\n", b.ID, b.IP)
	}

	// Create message handler
	msgHandler := router.InboundHandlerFunc(func(ctx context.Context, msg message.InboundMessage) {
		defer msg.OnFinishedHandling()

		switch msg.Op() {
		case message.HandshakeOp:
			// This shouldn't happen for outbound connections normally
			fmt.Printf("  📨 Received handshake from %s\n", msg.NodeID())

		case message.PeerListOp:
			if pl, ok := msg.Message().(*p2p.PeerList); ok {
				fmt.Printf("  📋 Received PeerList with %d peers\n", len(pl.ClaimedIpPorts))
				for _, c := range pl.ClaimedIpPorts {
					if addr, ok := netip.AddrFromSlice(c.IpAddr); ok && c.IpPort > 0 {
						peerAddr := netip.AddrPortFrom(addr, uint16(c.IpPort))
						if !seenIPs.Contains(peerAddr) {
							ipQueue = append(ipQueue, peerAddr)
						}
					}
				}
			}
		}
	})

	fmt.Println("\n🔄 Starting endless discovery loop...")
	fmt.Println("================================================\n")

	// Discovery loop
	successCount := 0
	failureCount := 0
	round := 0

	for {
		round++
		// Take up to 10 IPs per round
		maxPerRound := 10
		if len(ipQueue) < maxPerRound {
			maxPerRound = len(ipQueue)
		}

		if maxPerRound == 0 {
			fmt.Printf("\n⚠️  No more IPs in queue. Waiting for more peers... (Round %d)\n", round)
			time.Sleep(10 * time.Second)
			continue
		}

		currentQueue := ipQueue[:maxPerRound]
		ipQueue = ipQueue[maxPerRound:]

		fmt.Printf("\n🔄 Round %d: Processing %d IPs (%d remaining in queue)\n", round, len(currentQueue), len(ipQueue))

		// Create a wait group for parallel connections
		var wg sync.WaitGroup

		// Create channels for results
		type connectionResult struct {
			ip      netip.AddrPort
			success bool
			peer    *DiscoveredPeer
			newIPs  []netip.AddrPort
		}
		results := make(chan connectionResult, len(currentQueue))

		// Process connections in parallel
		for i, ip := range currentQueue {
			if seenIPs.Contains(ip) {
				continue
			}
			seenIPs.Add(ip)

			wg.Add(1)
			go func(idx int, addr netip.AddrPort) {
				defer wg.Done()

				fmt.Printf("🔗 [%d/%d] Starting connection to %s...\n", idx+1, len(currentQueue), addr)

				// Try to connect
				connectCtx, connectCancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer connectCancel()

				p, err := startDiscoveryPeer(connectCtx, addr, constants.MainnetID, network, msgHandler, tlsCert)
				if err != nil {
					fmt.Printf("  ❌ [%s] Failed: %v\n", addr, err)
					results <- connectionResult{ip: addr, success: false}
					return
				}

				// Get tracked subnets from peer
				trackedSubnets := p.TrackedSubnets()
				subnetIDs := make([]ids.ID, 0, trackedSubnets.Len())
				for subnet := range trackedSubnets {
					subnetIDs = append(subnetIDs, subnet)
				}

				peer := &DiscoveredPeer{
					NodeID:         p.ID(),
					IP:             addr,
					Version:        p.Version().String(),
					TrackedSubnets: subnetIDs,
					LastSeen:       time.Now(),
				}

				fmt.Printf("  ✅ [%s] Connected! Version: %s\n", addr, p.Version())
				if len(subnetIDs) > 0 {
					// Check if peer tracks our subnets
					tracksOurSubnets := false
					for _, subnet := range subnetIDs {
						if subnet.String() == "h7egyVb6fKHMDpVaEsTEcy7YaEnXrayxZS4A1AEU4pyBzmwGp" ||
							subnet.String() == "nQCwF6V9y8VFjvMuPeQVWWYn6ba75518Dpf6ZMWZNb3NyTA94" {
							tracksOurSubnets = true
							break
						}
					}

					if tracksOurSubnets {
						fmt.Printf("  🎯 [%s] TRACKS OUR SUBNETS! Subnets: %v\n", addr, subnetIDs)
					} else if len(subnetIDs) > 1 {
						fmt.Printf("  🌐 [%s] Tracking %d other subnet(s)\n", addr, len(subnetIDs))
					}
				}

				// Give time for message exchange
				time.Sleep(2 * time.Second)

				// Request peer list
				p.StartSendGetPeerList()

				// Wait for response
				time.Sleep(3 * time.Second)

				// Collect new IPs before closing
				// Note: In real implementation, we'd capture these from the PeerList message
				newIPs := []netip.AddrPort{}

				// Close connection
				p.StartClose()
				p.AwaitClosed(context.Background())

				results <- connectionResult{
					ip:      addr,
					success: true,
					peer:    peer,
					newIPs:  newIPs,
				}
			}(i, ip)
		}

		// Wait for all connections to complete
		go func() {
			wg.Wait()
			close(results)
		}()

		// Process results as they come in
		roundSuccess := 0
		roundFailure := 0
		for result := range results {
			if result.success {
				roundSuccess++
				successCount++

				// Store peer info
				discoveredMutex.Lock()
				discoveredPeers[result.peer.NodeID] = result.peer
				discoveredMutex.Unlock()

				// Add new IPs to queue
				for _, newIP := range result.newIPs {
					if !seenIPs.Contains(newIP) {
						ipQueue = append(ipQueue, newIP)
					}
				}
			} else {
				roundFailure++
				failureCount++
			}
		}

		fmt.Printf("\n📊 Round %d complete - Success: %d, Failures: %d\n", round, roundSuccess, roundFailure)

		// Print stats
		totalAttempts := successCount + failureCount
		successRate := float64(successCount) / float64(totalAttempts) * 100
		fmt.Printf("\n📊 Stats - Success: %d, Failures: %d, Success Rate: %.1f%%\n",
			successCount, failureCount, successRate)
		fmt.Printf("📊 Discovered peers: %d, IPs in queue: %d\n", len(discoveredPeers), len(ipQueue))

		// Take a break between rounds
		fmt.Println("\n⏳ Waiting 10 seconds before next round...")
		time.Sleep(10 * time.Second)
	}
}
