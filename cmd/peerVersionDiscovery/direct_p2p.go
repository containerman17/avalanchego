package main

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/message"
	"github.com/ava-labs/avalanchego/network/peer"
	"github.com/ava-labs/avalanchego/staking"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/version"
	"github.com/prometheus/client_golang/prometheus"
)

const targetSubnetID = "23dqTMHK186m4Rzcn1ukJdmHy13nqido4LjTp5Kh9W6qBKaFib"

func main() {
	fmt.Println("Direct P2P Connection to Avalanche Network")
	fmt.Println("=========================================")
	fmt.Printf("Target Subnet: %s\n\n", targetSubnetID)

	// Parse subnet ID
	subnetID, err := ids.FromString(targetSubnetID)
	if err != nil {
		log.Fatalf("Failed to parse subnet ID: %v", err)
	}

	// Generate TLS certificate for our node
	tlsCert, err := staking.NewTLSCert()
	if err != nil {
		log.Fatalf("Failed to generate TLS cert: %v", err)
	}

	fmt.Println("Generated TLS certificate for P2P connection")

	// Try connecting to a known Avalanche node
	// Using localhost first, then can try public nodes
	targetIP := "127.0.0.1:9651"
	
	fmt.Printf("\nAttempting direct P2P connection to %s...\n", targetIP)

	// Parse IP
	ip, err := netip.ParseAddrPort(targetIP)
	if err != nil {
		log.Fatalf("Invalid IP: %v", err)
	}

	// Create raw TCP connection
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.Dial("tcp", ip.String())
	if err != nil {
		fmt.Printf("Failed to establish TCP connection: %v\n", err)
		fmt.Println("\nTo run this tool:")
		fmt.Println("1. You need an Avalanche node running locally, OR")
		fmt.Println("2. Network access to Avalanche mainnet nodes on port 9651")
		fmt.Println("3. The tool will perform TLS handshake and exchange version info via P2P protocol")
		return
	}
	defer conn.Close()

	fmt.Println("TCP connection established, upgrading to TLS...")

	// Create TLS config
	tlsConfig := peer.TLSConfig(*tlsCert, nil)
	
	// Create client upgrader
	clientUpgrader := peer.NewTLSClientUpgrader(
		tlsConfig,
		prometheus.NewCounter(prometheus.CounterOpts{
			Name: "tls_upgrades",
			Help: "TLS upgrade attempts",
		}),
	)

	// Upgrade connection to TLS
	peerID, tlsConn, _, err := clientUpgrader.Upgrade(conn)
	if err != nil {
		log.Fatalf("TLS upgrade failed: %v", err)
	}

	fmt.Printf("TLS handshake complete! Connected to peer: %s\n", peerID)

	// Create message creator
	mc, err := message.NewCreator(
		prometheus.NewRegistry(),
		constants.DefaultNetworkCompressionType,
		10*time.Second,
	)
	if err != nil {
		log.Fatalf("Failed to create message creator: %v", err)
	}

	// Prepare handshake message
	myTime := uint64(time.Now().Unix())
	
	// Create IP signature (simplified - in production this would be proper)
	ipSig := make([]byte, 64)
	
	handshakeMsg, err := mc.Handshake(
		constants.MainnetID,
		myTime,
		netip.AddrPortFrom(netip.IPv4Unspecified(), 0),
		version.CurrentApp.String(),
		uint32(version.CurrentApp.Major),
		uint32(version.CurrentApp.Minor),
		uint32(version.CurrentApp.Patch),
		myTime,
		ipSig,
		[]byte{}, // BLS sig
		[]ids.ID{subnetID}, // We're interested in this subnet
		[]uint32{}, // supportedACPs
		[]uint32{}, // objectedACPs
		nil, // knownPeersFilter
		nil, // knownPeersSalt
		false, // requestAllSubnetIPs
	)
	if err != nil {
		log.Fatalf("Failed to create handshake: %v", err)
	}

	fmt.Println("\nSending Avalanche P2P handshake...")
	
	// Send handshake
	_, err = tlsConn.Write(handshakeMsg.Bytes())
	if err != nil {
		log.Fatalf("Failed to send handshake: %v", err)
	}

	fmt.Println("Handshake sent, waiting for response...")

	// Read response
	buf := make([]byte, 65536)
	tlsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := tlsConn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("\nReceived %d bytes response\n", n)

	// In a full implementation, we would:
	// 1. Parse the protobuf response properly
	// 2. Handle PeerList messages
	// 3. Connect to subnet-specific validators
	// 4. Build a complete picture of all validator versions

	fmt.Println("\nNext steps for complete implementation:")
	fmt.Println("1. Parse protobuf messages from response")
	fmt.Println("2. Handle PeerList to discover more subnet validators")
	fmt.Println("3. Connect to each validator and get their version")
	fmt.Println("4. Aggregate results into a complete report")
}