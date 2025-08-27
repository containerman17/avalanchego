package main

import (
	"context"
	"crypto"
	"net"
	"net/netip"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/message"
	"github.com/ava-labs/avalanchego/network/peer"
	"github.com/ava-labs/avalanchego/network/throttling"
	"github.com/ava-labs/avalanchego/snow/networking/router"
	"github.com/ava-labs/avalanchego/snow/networking/tracker"
	"github.com/ava-labs/avalanchego/snow/uptime"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/avalanchego/staking"
	"github.com/ava-labs/avalanchego/upgrade"
	"github.com/ava-labs/avalanchego/utils"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto/bls/signer/localsigner"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/math/meter"
	"github.com/ava-labs/avalanchego/utils/resource"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/version"
)

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

func startDiscoveryPeer(
	ctx context.Context,
	remoteIP netip.AddrPort,
	networkID uint32,
	network peer.Network,
	msgHandler router.InboundHandler,
) (peer.Peer, error) {
	// Connect to remote peer
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, constants.NetworkType, remoteIP.String())
	if err != nil {
		return nil, err
	}

	// Create TLS certificate
	tlsCert, err := staking.NewTLSCert()
	if err != nil {
		return nil, err
	}

	// Setup TLS
	tlsConfig := peer.TLSConfig(*tlsCert, nil)
	clientUpgrader := peer.NewTLSClientUpgrader(
		tlsConfig,
		prometheus.NewCounter(prometheus.CounterOpts{}),
	)

	peerID, conn, cert, err := clientUpgrader.Upgrade(conn)
	if err != nil {
		return nil, err
	}

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

	// Try to get our external IP
	ourIP := getOutboundIP()
	ourPort := uint16(9651) // Default Avalanche port

	parsedCert, err := staking.ParseCertificate(tlsCert.Leaf.Raw)
	if err != nil {
		return nil, err
	}
	myNodeID := ids.NodeIDFromCert(parsedCert)

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
		MySubnets:            set.Set[ids.ID]{},
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

	return p, p.AwaitReady(ctx)
}
