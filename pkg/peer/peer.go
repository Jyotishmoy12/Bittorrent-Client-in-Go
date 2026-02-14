package peer

import (
	"encoding/binary"
	"fmt"
	"net"
)

// Peer represents a peer in the BitTorrent network. It contains the IP address and port number of the peer.
type Peer struct {
	IP   net.IP
	Port uint16
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), fmt.Sprintf("%d", p.Port))
}

// unmarshal parses the compact peer list from the tracker response and returns a slice of Peer structs.

func Unmarshal(peersBin string) ([]Peer, error) {
	const peerSize = 6 // Each peer is represented by 6 bytes (4 for IP, 2 for Port)
	numPeers := len(peersBin) / peerSize
	if len(peersBin)%peerSize != 0 {
		return nil, fmt.Errorf("invalid peers binary length: must be a multiple of %d", peerSize)
	}
	peers := make([]Peer, numPeers)

	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		// Extract the 4 bytes for the IP address and 2 bytes for the port
		peers[i].IP = net.IP(peersBin[offset : offset+4])
		// Extract the 2-byte Port using Big Endian (standard for network byte order)
		peers[i].Port = binary.BigEndian.Uint16([]byte(peersBin[offset+4 : offset+6]))
	}

	return peers, nil
}
