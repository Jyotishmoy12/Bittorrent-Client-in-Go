package peer

import (
	"encoding/binary"
	"github.com/jyotishmoy12/bittorrent-go/pkg/pcode"
)

// MaxBlockSize is the standard size for a BitTorrent block (16KB)
const MaxBlockSize = 16384

// PieceProgress tracks the download of a single piece
type PieceProgress struct {
	Index      int
	Downloaded int
	Requested  int
	Buf        []byte
}

// RequestMessage builds a request for a specific block within a piece
func RequestMessage(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: pcode.MsgRequest, Payload: payload}
}
