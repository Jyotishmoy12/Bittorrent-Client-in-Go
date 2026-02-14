// The Handshake Format
// A handshake is a fixed-length message (68 bytes) consisting of:

// Length of the protocol identifier (1 byte): Always 19.

// Protocol identifier (19 bytes): The string "BitTorrent protocol".

// Reserved bytes (8 bytes): Mostly zeros, used for extensions.

// Info Hash (20 bytes): The hash we calculated earlier.

// Peer ID (20 bytes): Your random ID.

package peer

import (
	"fmt"
	"io"
)

// Handshake represents the message used to start a connection with a peer
type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

// Serialize converts the Handshake struct into a byte slice that can be sent over the network

func (h *Handshake) Serialize() []byte {
	buf := make([]byte, 49+len(h.Pstr))
	buf[0] = byte(len(h.Pstr)) // first byte: length of the protocol string
	curr := 1
	curr += copy(buf[curr:], h.Pstr)          // The string "BitTorrent protocol
	curr += copy(buf[curr:], make([]byte, 8)) // reserved bytes (8 zeros)
	curr += copy(buf[curr:], h.InfoHash[:])   // info hash (20 bytes)
	curr += copy(buf[curr:], h.PeerID[:])     // peer ID (20 bytes)
	return buf
}

func ReadHandshake(r io.Reader) (*Handshake, error) {
	// read the length of the protocol string (1 byte)
	lengthBuf := make([]byte, 1)
	//io.ReadFull: This is safer than r.Read.
	// It guarantees that we wait until the buffer is completely filled before moving on. Network packets are often fragmented,
	//  so this ensures we don't get a "half-handshake."
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	pstrLen := int(lengthBuf[0])
	if pstrLen == 0 {
		return nil, fmt.Errorf("invalid protocol string length: %d", pstrLen)
	}
	// read the rest of the handshake (pstrLen + 48 bytes)
	handshakeBuf := make([]byte, pstrLen+48)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}
	//pstrLen + 48: Since the protocol string is usually 19 bytes, $19 + 48 = 67$.
	// Plus the 1-byte length at the start makes the total 68 bytes.
	var infohash, peerID [20]byte
	copy(infohash[:], handshakeBuf[pstrLen+8:pstrLen+28]) // info hash starts after pstr and reserved bytes
	copy(peerID[:], handshakeBuf[pstrLen+28:])

	return &Handshake{
		Pstr:     string(handshakeBuf[0:pstrLen]),
		InfoHash: infohash,
		PeerID:   peerID,
	}, nil
}
