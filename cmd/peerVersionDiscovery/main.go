package main

import (
	"context"
	"fmt"
	"log"
	"net/netip"
	"time"

	"github.com/ava-labs/avalanchego/genesis"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/message"
	"github.com/ava-labs/avalanchego/proto/pb/p2p"
	"github.com/ava-labs/avalanchego/snow/networking/router"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/set"
)

func main() {
	ctx := context.Background()
	networkID := constants.MainnetID

	// Parse the target subnet ID
	targetSubnetStr := "h7egyVb6fKHMDpVaEsTEcy7YaEnXrayxZS4A1AEU4pyBzmwGp"
	targetSubnetID, err := ids.FromString(targetSubnetStr)
	if err != nil {
		log.Fatalf("Failed to parse subnet ID: %v", err)
	}
	fmt.Printf("Looking for peers tracking subnet: %s\n\n", targetSubnetID)

	// Create our discovery network and subnet discovery
	network := newDiscoveryNetwork()
	subnetDiscovery := NewSubnetDiscovery()

	// seed with default bootstrappers
	bootstrappers := genesis.SampleBootstrappers(networkID, 2)
	ipQueue := []netip.AddrPort{}
	fmt.Println("Bootstrap peers:")
	for _, b := range bootstrappers {
		ipQueue = append(ipQueue, b.IP)
		fmt.Printf("  %s\n", b.IP)
	}
	fmt.Println()

	seenIPs := set.Set[netip.AddrPort]{}
	seenNodes := set.Set[netip.AddrPort]{}
	discoveredPeers := []netip.AddrPort{}

	// First phase: discover peers and their subnet information
	for len(ipQueue) > 0 && seenNodes.Len() < 100 {
		ip := ipQueue[0]
		ipQueue = ipQueue[1:]
		if seenIPs.Contains(ip) {
			continue
		}
		seenIPs.Add(ip)

		fmt.Printf("Connecting to %s...\n", ip)

		peerCh := make(chan netip.AddrPort, 32)
		// Create a combined handler for subnet discovery and peer list
		subnetHandler := CreateHandshakeHandler(subnetDiscovery)
		handler := router.InboundHandlerFunc(func(ctx context.Context, msg message.InboundMessage) {
			// Handle subnet discovery first
			subnetHandler.HandleInbound(ctx, msg)

			// Then handle peer list for further discovery
			if msg.Op() == message.PeerListOp {
				if pl, ok := msg.Message().(*p2p.PeerList); ok {
					fmt.Printf("  PeerList contains %d peers\n", len(pl.ClaimedIpPorts))
					for _, c := range pl.ClaimedIpPorts {
						if addr, ok := netip.AddrFromSlice(c.IpAddr); ok {
							peerAddr := netip.AddrPortFrom(addr, uint16(c.IpPort))
							select {
							case peerCh <- peerAddr:
							default:
							}
						}
					}
				}
			}
		})

		connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		p, err := startDiscoveryPeer(connectCtx, ip, networkID, network, handler)
		cancel()
		if err != nil {
			log.Printf("Failed to connect to %s: %v", ip, err)
			continue
		}

		if seenNodes.Contains(ip) {
			p.StartClose()
			p.AwaitClosed(ctx)
			continue
		}
		seenNodes.Add(ip)

		fmt.Printf("Connected to %s (Node ID: %s, Version: %s)\n", ip, p.ID(), p.Version())

		// Wait a bit for handshake to complete and receive messages
		time.Sleep(3 * time.Second)

		fmt.Printf("  Requesting peer list...\n")
		p.StartSendGetPeerList()
		timeout := time.After(10 * time.Second)
	loop:
		for {
			select {
			case addr := <-peerCh:
				if !seenIPs.Contains(addr) {
					fmt.Printf("  Discovered new peer: %s\n", addr)
					discoveredPeers = append(discoveredPeers, addr)
					ipQueue = append(ipQueue, addr)
				} else {
					fmt.Printf("  Already seen peer: %s\n", addr)
				}
			case <-timeout:
				fmt.Printf("  Timeout waiting for more peers from %s\n", ip)
				break loop
			}
		}

		p.StartClose()
		p.AwaitClosed(ctx)
	}

	fmt.Printf("\nDiscovered %d peers total\n", len(discoveredPeers))

	// Show peers tracking the target subnet
	subnetPeers := subnetDiscovery.GetPeersTrackingSubnet(targetSubnetID)
	fmt.Printf("\nFound %d peers tracking subnet %s:\n", len(subnetPeers), targetSubnetID)
	for i, peer := range subnetPeers {
		fmt.Printf("%d. Node: %s\n", i+1, peer.NodeID)
		fmt.Printf("   IP: %s\n", peer.IP)
		fmt.Printf("   Version: %s\n", peer.Version)
		fmt.Printf("   Tracking %d subnets\n", peer.TrackedSubnets.Len())
	}

	if len(subnetPeers) == 0 {
		fmt.Printf("\nNo peers found tracking subnet %s yet.\n", targetSubnetID)
		fmt.Printf("You may need to connect to more peers or the subnet may not have many validators.\n")
	}

	// Connect directly to subnet peers to verify their versions
	fmt.Printf("\nConnecting directly to subnet peers to verify versions...\n")
	for _, peer := range subnetPeers {
		fmt.Printf("\nConnecting to %s at %s...\n", peer.NodeID, peer.IP)

		connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		verifyHandler := router.InboundHandlerFunc(func(_ context.Context, msg message.InboundMessage) {
			if msg.Op() == message.HandshakeOp {
				if hs, ok := msg.Message().(*p2p.Handshake); ok && hs.Client != nil {
					version := fmt.Sprintf("%s/%d.%d.%d",
						hs.Client.Name,
						hs.Client.Major,
						hs.Client.Minor,
						hs.Client.Patch)
					fmt.Printf("  Verified version: %s\n", version)
				}
			}
			msg.OnFinishedHandling()
		})

		p, err := startDiscoveryPeer(connectCtx, peer.IP, networkID, network, verifyHandler)
		cancel()
		if err != nil {
			fmt.Printf("  Failed to connect: %v\n", err)
			continue
		}

		// Give time for handshake
		time.Sleep(1 * time.Second)
		p.StartClose()
		p.AwaitClosed(context.Background())
	}
}
