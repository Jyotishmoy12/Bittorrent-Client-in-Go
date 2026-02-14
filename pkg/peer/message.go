package peer

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Message holds the ID and payload of a bit torrent message.
// The ID is used to identify the type of message (e.g., request, piece, etc.)
// and the payload contains the actual data being sent.

type Message struct {
	ID      uint8
	Payload []byte
}

// serialize converts a msg into a bitstream (Lenght+ID+Payload) that can be sent over the network.

func (m *Message) Serialize() []byte {
	if m == nil { // keep alive msg (length =0)
		return make([]byte, 4)
	}
	length := uint32(len(m.Payload) + 1)         // +1 for the ID byte
	buf := make([]byte, 4+length)                // 4 bytes for length prefix + message length
	binary.BigEndian.PutUint32(buf[0:4], length) // write the length prefix
	buf[4] = byte(m.ID)                          // write the message ID
	copy(buf[5:], m.Payload)                     // write the payload after the ID
	return buf
}

//Read parses a msg from a stream

func ReadMessage(r io.Reader) (*Message, error) {
	// read the length prefix (4 bytes)
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lenBuf)
	if err != nil {
		return nil, err
	}
	//BitTorrent uses Big Endian for all integers.
	//This ensures our 4-byte length prefix is written in the correct network byte order.
	length := binary.BigEndian.Uint32(lenBuf)
	if length == 0 {
		// This is a keep-alive message (no ID, no payload)
		return nil, nil
	}
	// read the message ID (1 byte)
	msgBuf := make([]byte, length)
	_, err = io.ReadFull(r, msgBuf)
	if err != nil {
		return nil, err
	}
	return &Message{
		ID:      uint8(msgBuf[0]), // first byte is the message ID
		Payload: msgBuf[1:],       // the rest is the payload
	}, nil
}

// Bitfield represents the pieces a peer has
// The Bitfield you received is 377 bytes long.
// Since 1 byte = 8 bits, this peer is describing $377 \times 8 = 3016$ pieces. Each bit represents a piece:
// 1 means they have it, 0 means they don't.
type Bitfield []byte

func (b Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	bitIndex := index % 8
	if byteIndex < 0 || byteIndex >= len(b) {
		return false
	}
	// BitTorrent uses bit order where the high bit of the first byte is index 0
	return b[byteIndex]>>(7-bitIndex)&1 != 0
}

// ParsePiece validates a PIECE message and returns the offset and the raw block data.
func ParsePiece(index int, buf []byte, msg *Message) (int, []byte, error) {
	// A Piece message payload must at least have 8 bytes (4 for index, 4 for begin)
	if len(msg.Payload) < 8 {
		return 0, nil, fmt.Errorf("payload too short: %d", len(msg.Payload))
	}

	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, nil, fmt.Errorf("expected index %d, got %d", index, parsedIndex)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, nil, fmt.Errorf("begin offset %d out of bounds", begin)
	}

	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, nil, fmt.Errorf("data too long for offset %d with length %d", begin, len(data))
	}

	return begin, data, nil
}

// Read parses a BitTorrent message from a stream
func Read(r io.Reader) (*Message, error) {
	// 1. Read the 4-byte length prefix
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// 2. Handle Keep-Alive messages (length 0)
	if length == 0 {
		return nil, nil
	}

	// 3. Read the rest of the message (ID + Payload)
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:      messageBuf[0],
		Payload: messageBuf[1:],
	}, nil
}
