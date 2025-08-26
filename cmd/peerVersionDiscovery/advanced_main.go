package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/ids"
)

// This is an advanced version that shows how to get more comprehensive data
// by querying the P-Chain for actual validators

const (
	targetSubnetID     = "23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib"
	defaultMainnetHost = "127.0.0.1:9651"
)

type ValidatorInfo struct {
	NodeID    string
	StartTime uint64
	EndTime   uint64
	Weight    uint64
}

func main() {
	fmt.Println("Advanced Subnet Validator Discovery (with P-Chain)")
	fmt.Println("=================================================")
	fmt.Printf("Target Subnet: %s\n\n", targetSubnetID)

	nodeURL := os.Getenv("AVALANCHE_NODE_URL")
	if nodeURL == "" {
		nodeURL = fmt.Sprintf("http://%s", defaultMainnetHost)
	}

	// Parse subnet ID
	subnetID, err := ids.FromString(targetSubnetID)
	if err != nil {
		log.Fatalf("Failed to parse subnet ID: %v", err)
	}

	// First, get peer info via info API
	infoClient := info.NewClient(nodeURL)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	peersReply, err := infoClient.Peers(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to get peers: %v", err)
	}

	// Create a map of connected peers
	connectedPeers := make(map[string]info.Peer)
	for _, p := range peersReply {
		connectedPeers[p.ID.String()] = p
	}

	// Now query P-Chain for actual validators
	fmt.Println("Querying P-Chain for subnet validators...")
	
	// Make RPC call to platform.getCurrentValidators
	type GetCurrentValidatorsArgs struct {
		SubnetID string `json:"subnetID"`
	}
	
	type Validator struct {
		NodeID    string `json:"nodeID"`
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
		Weight    string `json:"weight"`
	}
	
	type GetCurrentValidatorsReply struct {
		Validators []Validator `json:"validators"`
	}

	// Prepare RPC request
	args := GetCurrentValidatorsArgs{
		SubnetID: targetSubnetID,
	}
	
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "platform.getCurrentValidators",
		"params":  args,
		"id":      1,
	})
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	// Make HTTP request
	resp, err := http.Post(nodeURL+"/ext/bc/P", "application/json", strings.NewReader(string(requestBody)))
	if err != nil {
		log.Fatalf("Failed to make RPC call: %v", err)
	}
	defer resp.Body.Close()

	var rpcResponse struct {
		Result GetCurrentValidatorsReply `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}
	
	if rpcResponse.Error != nil {
		log.Fatalf("RPC error: %s", rpcResponse.Error.Message)
	}

	fmt.Printf("\nFound %d validators for subnet %s\n", len(rpcResponse.Result.Validators), targetSubnetID)
	
	// Display validator information
	fmt.Println("\nValidator Details:")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("%-44s %-20s %-20s %s\n", "Node ID", "Status", "IP", "Version")
	fmt.Println(strings.Repeat("-", 100))

	connectedCount := 0
	versionCount := make(map[string]int)
	
	for _, validator := range rpcResponse.Result.Validators {
		nodeID := validator.NodeID
		status := "Not Connected"
		ip := "Unknown"
		version := "Unknown"
		
		// Check if we're connected to this validator
		if peer, ok := connectedPeers[nodeID]; ok {
			status = "Connected"
			ip = peer.IP.String()
			version = peer.Version
			connectedCount++
			
			// Only count versions for connected peers
			versionCount[version]++
		}
		
		fmt.Printf("%-44s %-20s %-20s %s\n", nodeID, status, ip, version)
	}

	// Summary
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("\nTotal subnet validators: %d\n", len(rpcResponse.Result.Validators))
	fmt.Printf("Connected to: %d (%d%%)\n", connectedCount, (connectedCount*100)/len(rpcResponse.Result.Validators))
	fmt.Printf("Not connected to: %d\n", len(rpcResponse.Result.Validators)-connectedCount)
	
	if connectedCount > 0 {
		fmt.Println("\nVersion distribution (connected validators only):")
		
		var versions []string
		for v := range versionCount {
			versions = append(versions, v)
		}
		sort.Strings(versions)
		
		for _, v := range versions {
			fmt.Printf("  %s: %d nodes\n", v, versionCount[v])
		}
	}
	
	fmt.Println("\nNOTE: To get version info for ALL validators, you would need to:")
	fmt.Println("1. Implement P2P protocol to connect directly to each validator")
	fmt.Println("2. Or have your node connect to more peers by adjusting network settings")
	fmt.Println("3. Or query multiple nodes and aggregate the data")
}