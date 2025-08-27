package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
)

// RunInteractiveMode runs an interactive search mode
func RunInteractiveMode(peerDB *SearchablePeerDB) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("\n=== Interactive Peer Search Mode ===")
	fmt.Println("Commands:")
	fmt.Println("  node <nodeID>     - Search by Node ID (partial match supported)")
	fmt.Println("  subnet <subnetID> - Search by Subnet ID (partial match supported)")
	fmt.Println("  list subnets      - List all discovered subnets")
	fmt.Println("  list peers        - List all discovered peers")
	fmt.Println("  connect <nodeID>  - Connect to a specific peer by NodeID")
	fmt.Println("  stats             - Show discovery statistics")
	fmt.Println("  quit              - Exit")
	fmt.Println()

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return

		case "node":
			if len(parts) < 2 {
				fmt.Println("Usage: node <nodeID>")
				continue
			}
			searchNodeID(peerDB, parts[1])

		case "subnet":
			if len(parts) < 2 {
				fmt.Println("Usage: subnet <subnetID>")
				continue
			}
			searchSubnet(peerDB, parts[1])

		case "list":
			if len(parts) < 2 {
				fmt.Println("Usage: list [subnets|peers]")
				continue
			}
			switch parts[1] {
			case "subnets":
				listSubnets(peerDB)
			case "peers":
				listPeers(peerDB)
			default:
				fmt.Println("Unknown list command. Use 'list subnets' or 'list peers'")
			}

		case "connect":
			if len(parts) < 2 {
				fmt.Println("Usage: connect <nodeID>")
				continue
			}
			connectToPeer(peerDB, parts[1])

		case "stats":
			showStats(peerDB)

		default:
			fmt.Printf("Unknown command: %s\n", parts[0])
			fmt.Println("Type 'help' for available commands")
		}
	}
}

func searchNodeID(peerDB *SearchablePeerDB, nodeIDStr string) {
	peers := peerDB.SearchByNodeID(nodeIDStr)
	if len(peers) == 0 {
		fmt.Printf("No peers found matching NodeID: %s\n", nodeIDStr)
		return
	}

	fmt.Printf("Found %d peer(s) matching '%s':\n", len(peers), nodeIDStr)
	for _, peer := range peers {
		printPeerInfo(peer)
	}
}

func searchSubnet(peerDB *SearchablePeerDB, subnetStr string) {
	// Try to parse as full subnet ID first
	var searchingFor string
	if subnetID, err := ids.FromString(subnetStr); err == nil {
		searchingFor = subnetID.String()
	} else {
		searchingFor = subnetStr
	}

	peers := peerDB.SearchBySubnet(subnetStr)
	if len(peers) == 0 {
		fmt.Printf("No peers found tracking subnet: %s\n", searchingFor)
		return
	}

	fmt.Printf("Found %d peer(s) tracking subnet '%s':\n", len(peers), searchingFor)
	for _, peer := range peers {
		printPeerInfo(peer)
	}
}

func listSubnets(peerDB *SearchablePeerDB) {
	subnets := peerDB.GetAllSubnets()
	if len(subnets) == 0 {
		fmt.Println("No subnets discovered yet")
		return
	}

	fmt.Printf("Discovered %d subnet(s):\n", len(subnets))
	for subnetID, count := range subnets {
		fmt.Printf("  %s - tracked by %d peer(s)\n", subnetID, count)
	}
}

func listPeers(peerDB *SearchablePeerDB) {
	peers := peerDB.GetAllPeers()
	if len(peers) == 0 {
		fmt.Println("No peers discovered yet")
		return
	}

	fmt.Printf("Discovered %d peer(s):\n", len(peers))
	for i, peer := range peers {
		fmt.Printf("\n%d. ", i+1)
		printPeerInfo(peer)
	}
}

func printPeerInfo(peer *PeerInfo) {
	fmt.Printf("NodeID: %s\n", peer.NodeID)
	fmt.Printf("   IP: %s\n", peer.IP)
	fmt.Printf("   Version: %s\n", peer.Version)
	if peer.TrackedSubnets.Len() > 1 { // More than primary network
		fmt.Printf("   Subnets (%d):\n", peer.TrackedSubnets.Len()-1)
		for subnetID := range peer.TrackedSubnets {
			if subnetID != ids.Empty {
				fmt.Printf("     - %s\n", subnetID)
			}
		}
	}
}

func connectToPeer(peerDB *SearchablePeerDB, nodeIDStr string) {
	peers := peerDB.SearchByNodeID(nodeIDStr)
	if len(peers) == 0 {
		fmt.Printf("No peer found with NodeID: %s\n", nodeIDStr)
		return
	}

	peer := peers[0]
	if len(peers) > 1 {
		fmt.Printf("Multiple peers match. Connecting to: %s\n", peer.NodeID)
	}

	fmt.Printf("Connecting to %s at %s...\n", peer.NodeID, peer.IP)

	// Create a simple connection to verify version
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	network := newDiscoveryNetwork()
	p, err := startDiscoveryPeer(ctx, peer.IP, constants.MainnetID, network, nil)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}

	fmt.Printf("Successfully connected! Version: %s\n", p.Version())
	p.StartClose()
	p.AwaitClosed(context.Background())
}

func showStats(peerDB *SearchablePeerDB) {
	allPeers := peerDB.GetAllPeers()
	subnets := peerDB.GetAllSubnets()

	// Count versions
	versions := make(map[string]int)
	for _, peer := range allPeers {
		versions[peer.Version]++
	}

	fmt.Println("\n=== Discovery Statistics ===")
	fmt.Printf("Total peers discovered: %d\n", len(allPeers))
	fmt.Printf("Total subnets discovered: %d\n", len(subnets))

	fmt.Println("\nVersion distribution:")
	for version, count := range versions {
		fmt.Printf("  %s: %d peers\n", version, count)
	}

	fmt.Println("\nTop subnets by peer count:")
	// Convert to slice for sorting
	type subnetCount struct {
		ID    ids.ID
		Count int
	}
	var subnetList []subnetCount
	for id, count := range subnets {
		subnetList = append(subnetList, subnetCount{id, count})
	}

	// Simple bubble sort for top 10
	for i := 0; i < len(subnetList); i++ {
		for j := i + 1; j < len(subnetList); j++ {
			if subnetList[j].Count > subnetList[i].Count {
				subnetList[i], subnetList[j] = subnetList[j], subnetList[i]
			}
		}
	}

	limit := 10
	if len(subnetList) < limit {
		limit = len(subnetList)
	}

	for i := 0; i < limit; i++ {
		fmt.Printf("  %s: %d peers\n", subnetList[i].ID, subnetList[i].Count)
	}
}
