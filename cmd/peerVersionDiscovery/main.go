package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/ids"
)

const (
	// Target subnet ID
	targetSubnetID = "23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib"
	
	// Mainnet bootstrap nodes
	defaultMainnetHost = "127.0.0.1:9651" // Connect to local node first
)

// PeerVersion holds peer info and version
type PeerVersion struct {
	NodeID  string
	IP      string
	Version string
}

func main() {
	fmt.Println("Avalanche Subnet Peer Version Discovery")
	fmt.Println("======================================")
	fmt.Printf("Target Subnet: %s\n\n", targetSubnetID)

	// Parse subnet ID
	subnetID, err := ids.FromString(targetSubnetID)
	if err != nil {
		log.Fatalf("Failed to parse subnet ID: %v", err)
	}

	// Try to connect to local node first via RPC
	nodeURL := os.Getenv("AVALANCHE_NODE_URL")
	if nodeURL == "" {
		nodeURL = fmt.Sprintf("http://%s", defaultMainnetHost)
	}

	fmt.Printf("Connecting to node at: %s\n", nodeURL)

	// Create info client
	infoClient := info.NewClient(nodeURL)
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get node info first
	nodeInfoReply, err := infoClient.GetNodeID(ctx)
	if err != nil {
		log.Fatalf("Failed to get node ID: %v", err)
	}
	fmt.Printf("Connected to node: %s\n", nodeInfoReply.NodeID)

	// Get network info
	networkInfo, err := infoClient.GetNodeVersion(ctx)
	if err != nil {
		log.Fatalf("Failed to get node version: %v", err)
	}
	fmt.Printf("Node version: %s\n", networkInfo.Version)

	// Get all peers
	fmt.Println("\nFetching peer information...")
	peersReply, err := infoClient.Peers(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to get peers: %v", err)
	}

	fmt.Printf("Total peers: %d\n", len(peersReply))

	// Filter peers by subnet
	var subnetPeers []PeerVersion
	for _, p := range peersReply {
		// Check if peer tracks our target subnet
		trackedSubnets := p.TrackedSubnets
		if trackedSubnets.Contains(subnetID) {
			peerVersion := PeerVersion{
				NodeID:  p.ID.String(),
				IP:      p.IP.String(),
				Version: p.Version,
			}
			subnetPeers = append(subnetPeers, peerVersion)
		}
	}

	// Sort by version
	sort.Slice(subnetPeers, func(i, j int) bool {
		return subnetPeers[i].Version < subnetPeers[j].Version
	})

	// Display results
	fmt.Printf("\n\nSubnet %s Validators:\n", targetSubnetID)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("%-44s %-20s %s\n", "Node ID", "IP", "Version")
	fmt.Println(strings.Repeat("-", 80))

	// Group by version
	versionCount := make(map[string]int)
	for _, peer := range subnetPeers {
		fmt.Printf("%-44s %-20s %s\n", peer.NodeID, peer.IP, peer.Version)
		versionCount[peer.Version]++
	}

	// Summary
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("\nTotal subnet validators found: %d\n", len(subnetPeers))
	fmt.Println("\nVersion distribution:")
	
	// Sort versions for display
	var versions []string
	for v := range versionCount {
		versions = append(versions, v)
	}
	sort.Strings(versions)
	
	for _, v := range versions {
		fmt.Printf("  %s: %d nodes\n", v, versionCount[v])
	}

	// Check for outdated versions
	fmt.Println("\nVersion Analysis:")
	latestVersion := networkInfo.Version
	fmt.Printf("Current node version: %s\n", latestVersion)
	
	outdatedCount := 0
	for _, v := range versions {
		if v != latestVersion {
			outdatedCount += versionCount[v]
		}
	}
	
	if outdatedCount > 0 {
		fmt.Printf("\nWARNING: %d nodes (%d%%) are running versions different from %s\n", 
			outdatedCount, 
			(outdatedCount*100)/len(subnetPeers),
			latestVersion)
	} else {
		fmt.Println("\nAll subnet validators are running the same version!")
	}
}