package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/message"
	"github.com/ava-labs/avalanchego/network/peer"
	"github.com/ava-labs/avalanchego/staking"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/version"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	targetSubnetID = "23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib"
	defaultTimeout = 30 * time.Second
)

type PeerInfo struct {
	NodeID  string
	IP      string
	Version string
}

func main() {
	fmt.Println("Avalanche P2P Subnet Peer Version Discovery")
	fmt.Println("==========================================")
	fmt.Printf("Target Subnet: %s\n\n", targetSubnetID)

	// Parse subnet ID
	subnetID, err := ids.FromString(targetSubnetID)
	if err != nil {
		log.Fatalf("Failed to parse subnet ID: %v", err)
	}

	// First, get validator IPs from P-Chain
	nodeURL := os.Getenv("AVALANCHE_NODE_URL")
	if nodeURL == "" {
		nodeURL = "http://127.0.0.1:9650"
	}

	fmt.Printf("Getting subnet validators from: %s\n", nodeURL)
	validators, err := getSubnetValidators(nodeURL, targetSubnetID)
	if err != nil {
		log.Fatalf("Failed to get subnet validators: %v", err)
	}

	fmt.Printf("Found %d validators for subnet\n", len(validators))

	// Try to get peer list to find IPs
	fmt.Println("\nAttempting to discover peer IPs...")
	peerIPs := discoverPeerIPs(nodeURL, validators)
	
	if len(peerIPs) == 0 {
		fmt.Println("\nNo peer IPs discovered. Using bootstrap nodes...")
		// Use mainnet bootstrap nodes
		peerIPs = getMainnetBootstrapNodes()
	}

	// Create TLS certificate for our client
	tlsCert, err := staking.NewTLSCert()
	if err != nil {
		log.Fatalf("Failed to create TLS cert: %v", err)
	}

	// Connect to peers and get versions
	fmt.Printf("\nConnecting to %d peers via P2P...\n", len(peerIPs))
	
	var wg sync.WaitGroup
	resultsChan := make(chan PeerInfo, len(peerIPs))
	
	for ip, nodeID := range peerIPs {
		wg.Add(1)
		go func(ip string, nodeID string) {
			defer wg.Done()
			
			version, err := connectAndGetVersion(ip, nodeID, tlsCert, subnetID)
			if err != nil {
				fmt.Printf("Failed to connect to %s: %v\n", ip, err)
				return
			}
			
			resultsChan <- PeerInfo{
				NodeID:  nodeID,
				IP:      ip,
				Version: version,
			}
		}(ip, nodeID)
	}
	
	// Wait for all connections to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()
	
	// Collect results
	var peers []PeerInfo
	for peer := range resultsChan {
		peers = append(peers, peer)
	}
	
	// Display results
	fmt.Printf("\n\nSuccessfully connected to %d peers\n", len(peers))
	fmt.Println("\nSubnet Validator Versions:")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("%-44s %-20s %s\n", "Node ID", "IP", "Version")
	fmt.Println(strings.Repeat("-", 80))
	
	versionCount := make(map[string]int)
	for _, peer := range peers {
		fmt.Printf("%-44s %-20s %s\n", peer.NodeID, peer.IP, peer.Version)
		versionCount[peer.Version]++
	}
	
	// Summary
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("\nVersion distribution:\n")
	for version, count := range versionCount {
		fmt.Printf("  %s: %d nodes\n", version, count)
	}
}

func connectAndGetVersion(ipStr string, nodeIDStr string, tlsCert *tls.Certificate, subnetID ids.ID) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse IP
	ip, err := netip.ParseAddrPort(ipStr)
	if err != nil {
		// Try adding default port
		ip, err = netip.ParseAddrPort(ipStr + ":9651")
		if err != nil {
			return "", fmt.Errorf("invalid IP: %w", err)
		}
	}

	// Create connection
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, constants.NetworkType, ip.String())
	if err != nil {
		return "", fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	// Upgrade to TLS
	tlsConfig := peer.TLSConfig(*tlsCert, nil)
	clientUpgrader := peer.NewTLSClientUpgrader(
		tlsConfig,
		prometheus.NewCounter(prometheus.CounterOpts{}),
	)

	peerID, conn, _, err := clientUpgrader.Upgrade(conn)
	if err != nil {
		return "", fmt.Errorf("TLS upgrade failed: %w", err)
	}

	// Create message creator
	mc, err := message.NewCreator(
		prometheus.NewRegistry(),
		constants.DefaultNetworkCompressionType,
		10*time.Second,
	)
	if err != nil {
		return "", err
	}

	// Send handshake
	myTime := uint64(time.Now().Unix())
	sig, err := tlsCert.PrivateKey.(interface {
		Sign([]byte, tls.SignatureScheme) ([]byte, error)
	}).Sign([]byte{}, tls.Ed25519)
	if err != nil {
		return "", err
	}

	handshake, err := mc.Handshake(
		constants.MainnetID,
		myTime,
		netip.AddrPortFrom(netip.IPv4Unspecified(), 0),
		version.CurrentApp.String(),
		version.CurrentApp.Major,
		version.CurrentApp.Minor,
		version.CurrentApp.Patch,
		myTime,
		sig,
		[]byte{}, // No BLS sig for now
		[]ids.ID{subnetID}, // Track target subnet
		[]uint32{}, // supportedACPs
		[]uint32{}, // objectedACPs
		nil, // knownPeersFilter
		nil, // knownPeersSalt
		false, // requestAllSubnetIPs
	)
	if err != nil {
		return "", err
	}

	// Send handshake
	_, err = conn.Write(handshake.Bytes())
	if err != nil {
		return "", err
	}

	// Read response
	// This is simplified - in production you'd need proper message framing
	buf := make([]byte, 65536)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	// Parse response to get version
	// In a real implementation, you'd properly parse the p2p messages
	// For now, we'll look for version string in the response
	response := string(buf[:n])
	
	// Extract version from handshake response
	// This is a simplified extraction - real implementation would parse protobuf
	if strings.Contains(response, "avalanche/") {
		start := strings.Index(response, "avalanche/")
		end := start + 20 // Approximate version string length
		if end > len(response) {
			end = len(response)
		}
		versionStr := response[start:end]
		// Clean up version string
		for i, ch := range versionStr {
			if ch < 32 || ch > 126 {
				versionStr = versionStr[:i]
				break
			}
		}
		return versionStr, nil
	}

	return fmt.Sprintf("Unknown (peer: %s)", peerID), nil
}

func getSubnetValidators(nodeURL string, subnetID string) ([]string, error) {
	// Query P-Chain for validators
	type GetCurrentValidatorsArgs struct {
		SubnetID string `json:"subnetID"`
	}
	
	type Validator struct {
		NodeID string `json:"nodeID"`
	}
	
	type GetCurrentValidatorsReply struct {
		Validators []Validator `json:"validators"`
	}

	args := GetCurrentValidatorsArgs{
		SubnetID: subnetID,
	}
	
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "platform.getCurrentValidators",
		"params":  args,
		"id":      1,
	})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(nodeURL+"/ext/bc/P", "application/json", strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResponse struct {
		Result GetCurrentValidatorsReply `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
		return nil, err
	}
	
	if rpcResponse.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", rpcResponse.Error.Message)
	}

	var nodeIDs []string
	for _, v := range rpcResponse.Result.Validators {
		nodeIDs = append(nodeIDs, v.NodeID)
	}
	
	return nodeIDs, nil
}

func discoverPeerIPs(nodeURL string, validators []string) map[string]string {
	// Try to get peer info to map node IDs to IPs
	// This is a simplified version - in production you'd query multiple sources
	
	peerIPs := make(map[string]string)
	
	// You could query info.peers here to get some IPs
	// For now, returning empty to demonstrate bootstrap approach
	
	return peerIPs
}

func getMainnetBootstrapNodes() map[string]string {
	// These are example IPs - in production you'd use real bootstrap nodes
	return map[string]string{
		"bootstrap1.avax.network:9651": "NodeID-A6onFGyJjA37EZ7kYHANMR1PFRT8NmXrF",
		"bootstrap2.avax.network:9651": "NodeID-6SwnPJLH8cWfrJ162JjZekbmzaFpjPcf",
		"bootstrap3.avax.network:9651": "NodeID-GSgaA47umS1p8BYPv9bPg7Zw5TFaKqKh",
		"bootstrap4.avax.network:9651": "NodeID-BQEo5Fy1FRKLbX51ejqDd14cuWXJLbZk",
		"bootstrap5.avax.network:9651": "NodeID-Drv1Qh7iJvW3zGBBeRnYfCzk56VCRM2GQ",
	}
}