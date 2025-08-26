package main

import (
	"context"
	"fmt"
	"log"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/message"
	"github.com/ava-labs/avalanchego/network/peer"
	"github.com/ava-labs/avalanchego/proto/pb/p2p"
	"github.com/ava-labs/avalanchego/snow/networking/router"
	"github.com/ava-labs/avalanchego/utils/constants"
)

const targetSubnetID = "23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib"

type PeerVersionInfo struct {
	NodeID  ids.NodeID
	IP      string
	Version string
}

func main() {
	fmt.Println("Avalanche P2P Subnet Peer Version Discovery (Simple)")
	fmt.Println("==================================================")
	fmt.Printf("Target Subnet: %s\n\n", targetSubnetID)

	// Parse subnet ID
	subnetID, err := ids.FromString(targetSubnetID)
	if err != nil {
		log.Fatalf("Failed to parse subnet ID: %v", err)
	}

	// List of known mainnet bootstrap nodes
	bootstrapNodes := []string{
		"54.176.202.126:9651",   // bootstrap1.avax.network
		"13.58.161.155:9651",    // bootstrap2.avax.network
		"35.154.12.125:9651",    // bootstrap3.avax.network
		"52.54.230.211:9651",    // bootstrap4.avax.network
		"35.178.98.242:9651",    // bootstrap5.avax.network
	}

	fmt.Printf("Attempting to connect to %d bootstrap nodes...\n", len(bootstrapNodes))

	var wg sync.WaitGroup
	resultsChan := make(chan PeerVersionInfo, len(bootstrapNodes))
	
	for _, nodeAddr := range bootstrapNodes {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			
			info, err := connectToPeer(addr, subnetID)
			if err != nil {
				fmt.Printf("Failed to connect to %s: %v\n", addr, err)
				return
			}
			
			if info != nil {
				resultsChan <- *info
			}
		}(nodeAddr)
	}
	
	// Wait for all connections
	go func() {
		wg.Wait()
		close(resultsChan)
	}()
	
	// Collect results
	var connectedPeers []PeerVersionInfo
	for peer := range resultsChan {
		connectedPeers = append(connectedPeers, peer)
	}
	
	// Display results
	if len(connectedPeers) == 0 {
		fmt.Println("\nFailed to connect to any peers.")
		fmt.Println("Make sure you have network connectivity to Avalanche mainnet.")
		return
	}
	
	fmt.Printf("\n\nSuccessfully connected to %d peers\n", len(connectedPeers))
	fmt.Println("\nPeer Versions:")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("%-44s %-20s %s\n", "Node ID", "IP", "Version")
	fmt.Println(strings.Repeat("-", 80))
	
	for _, peer := range connectedPeers {
		fmt.Printf("%-44s %-20s %s\n", peer.NodeID, peer.IP, peer.Version)
	}
	
	fmt.Println("\nNOTE: To discover subnet-specific validators:")
	fmt.Println("1. These bootstrap nodes can provide peer lists")
	fmt.Println("2. You would need to request GetPeerList and filter by subnet")
	fmt.Println("3. Then connect to those specific subnet validators")
}

func connectToPeer(addr string, subnetID ids.ID) (*PeerVersionInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse address
	ip, err := netip.ParseAddrPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Create a simple handler to capture handshake info
	var peerVersion string
	var peerNodeID ids.NodeID
	handshakeReceived := make(chan struct{})
	
	handler := router.InboundHandlerFunc(func(_ context.Context, msg message.InboundMessage) {
		// Check if this is a handshake or peerlist message
		switch msg.Op() {
		case message.HandshakeOp:
			// Parse handshake to get version
			handshakeMsg, ok := msg.Message().(*p2p.Handshake)
			if ok && handshakeMsg.Client != nil {
				peerVersion = fmt.Sprintf("%s/%d.%d.%d", 
					handshakeMsg.Client.Name,
					handshakeMsg.Client.Major,
					handshakeMsg.Client.Minor,
					handshakeMsg.Client.Patch,
				)
				peerNodeID = msg.NodeID()
				close(handshakeReceived)
			}
		case message.PeerListOp:
			// Could parse peer list here to discover more peers
			fmt.Printf("Received peer list from %s\n", msg.NodeID())
		}
	})

	// Connect using test peer (simplified connection)
	testPeer, err := peer.StartTestPeer(
		ctx,
		ip,
		constants.MainnetID,
		handler,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		testPeer.StartClose()
		testPeer.AwaitClosed(context.Background())
	}()

	// Wait for handshake or timeout
	select {
	case <-handshakeReceived:
		return &PeerVersionInfo{
			NodeID:  peerNodeID,
			IP:      addr,
			Version: peerVersion,
		}, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for handshake")
	}
}