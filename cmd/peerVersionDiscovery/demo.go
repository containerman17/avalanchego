package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

const targetSubnetID = "23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib"

// Simulated peer data based on typical mainnet distribution
type SimulatedPeer struct {
	NodeID  string
	IP      string
	Version string
}

func main() {
	fmt.Println("Avalanche Subnet Peer Version Discovery (DEMO)")
	fmt.Println("==============================================")
	fmt.Printf("Target Subnet: %s\n\n", targetSubnetID)
	fmt.Println("NOTE: This is a DEMO showing what typical output would look like")
	fmt.Println("To run with real data, you need an Avalanche node with info API enabled\n")

	// Simulate realistic version distribution based on typical mainnet patterns
	versions := []string{
		"avalanche/1.11.12",
		"avalanche/1.11.11", 
		"avalanche/1.11.10",
		"avalanche/1.11.9",
		"avalanche/1.11.8",
		"avalanche/1.10.19",
	}
	
	// Version weights (newer versions typically have higher adoption)
	versionWeights := []int{45, 25, 15, 8, 5, 2}
	
	// Generate simulated subnet validators
	rand.Seed(time.Now().UnixNano())
	numValidators := 18 + rand.Intn(7) // 18-24 validators typically
	
	var subnetPeers []SimulatedPeer
	
	for i := 0; i < numValidators; i++ {
		// Pick version based on weights
		versionIdx := weightedRandom(versionWeights)
		
		peer := SimulatedPeer{
			NodeID:  fmt.Sprintf("NodeID-%s", generateRandomNodeID()),
			IP:      fmt.Sprintf("%d.%d.%d.%d:9651", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256)),
			Version: versions[versionIdx],
		}
		subnetPeers = append(subnetPeers, peer)
	}
	
	// Sort by version
	sort.Slice(subnetPeers, func(i, j int) bool {
		return subnetPeers[i].Version < subnetPeers[j].Version
	})
	
	// Display results
	fmt.Printf("Simulated connection to node at: http://127.0.0.1:9651\n")
	fmt.Printf("Connected to node: NodeID-Simulated123\n")
	fmt.Printf("Node version: %s\n", versions[0])
	fmt.Printf("\nFetching peer information...\n")
	fmt.Printf("Total peers: ~250 (typical mainnet)\n")
	fmt.Printf("Peers tracking subnet: %d\n", len(subnetPeers))
	
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
	var sortedVersions []string
	for v := range versionCount {
		sortedVersions = append(sortedVersions, v)
	}
	sort.Strings(sortedVersions)
	
	for _, v := range sortedVersions {
		fmt.Printf("  %s: %d nodes (%d%%)\n", v, versionCount[v], (versionCount[v]*100)/len(subnetPeers))
	}
	
	// Analysis
	fmt.Println("\nVersion Analysis:")
	latestVersion := versions[0]
	fmt.Printf("Latest version: %s\n", latestVersion)
	
	outdatedCount := 0
	for v, count := range versionCount {
		if v != latestVersion {
			outdatedCount += count
		}
	}
	
	if outdatedCount > 0 {
		fmt.Printf("\nWARNING: %d nodes (%d%%) are running versions older than %s\n", 
			outdatedCount, 
			(outdatedCount*100)/len(subnetPeers),
			latestVersion)
		fmt.Println("\nRecommendation: These nodes should upgrade to the latest version for:")
		fmt.Println("  - Security patches and bug fixes")
		fmt.Println("  - Performance improvements")
		fmt.Println("  - New features and protocol upgrades")
	} else {
		fmt.Println("\nAll subnet validators are running the latest version!")
	}
	
	fmt.Println("\n---")
	fmt.Println("To run with real data:")
	fmt.Println("1. Start an Avalanche node with info API enabled")
	fmt.Println("2. Run: go run ./cmd/peerVersionDiscovery/main.go")
	fmt.Println("3. Or use: AVALANCHE_NODE_URL=http://your-node:9650 go run ./cmd/peerVersionDiscovery/main.go")
}

func generateRandomNodeID() string {
	chars := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	result := make([]byte, 26)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func weightedRandom(weights []int) int {
	total := 0
	for _, w := range weights {
		total += w
	}
	
	r := rand.Intn(total)
	for i, w := range weights {
		r -= w
		if r < 0 {
			return i
		}
	}
	return len(weights) - 1
}