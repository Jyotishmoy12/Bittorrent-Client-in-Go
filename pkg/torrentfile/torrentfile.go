package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

// bencodeTorrent is the internak representation of a torrent file. It is used to unmarshal the bencoded data from the torrent file.
type bencodeTorrent struct {
	Announce string      `bencode:"announce"` // The URL of the tracker that coordinates the torrent swarm
	Info     bencodeInfo `bencode:"info"` // The file metadata, which includes the piece hashes, piece length, total length, and file name
}

// bencodeInfo contains the actual file metadata
type bencodeInfo struct {
	Pieces      string `bencode:"pieces"` // A string containing the concatenated SHA-1 hashes of each piece
	PieceLength int    `bencode:"piece length"` // The length of each piece
	Length      int    `bencode:"length"` // The total length of the file
	Name        string `bencode:"name"` // The name of the file being shared
}

func Open(path string) (bencodeTorrent, error) {
	// open the file from the disk
	file, err := os.Open(path)
	if err != nil {
		return bencodeTorrent{}, err
	}
	defer file.Close()

	//create an empty struct to hold our data
	bto := bencodeTorrent{}

	// unmarshal the bencoded data from the file into our struct
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return bencodeTorrent{}, err
	}

	return bto, nil
}

// InfoHash calculates the SHA-1 hash of the bencoded info dictionary.
// This is the unique ID used to identify the torrent to trackers and peers.
func (b *bencodeTorrent) InfoHash() ([20]byte, error) {
	var buf bytes.Buffer
	// We must marshal the Info section back to bencode to get the correct hash
	err := bencode.Marshal(&buf, b.Info)

	if err != nil {
		return [20]byte{}, err
	}
	// generates the 20-byte fingerprint of the file metadata
	return sha1.Sum(buf.Bytes()), nil
}

// SplitPieceHashes breaks the giant Pieces string into a slice of 20-byte hashes.

func (b *bencodeTorrent) SplitPieceHashes() ([][20]byte, error) {
	hashLen := 20 // Each piece hash is 20 bytes long
	buf := []byte(b.Info.Pieces)

	if len(buf)%hashLen != 0 {
		return nil, fmt.Errorf("invalid pieces length: must be a multiple of %d", hashLen)
	}
	// Calculate the number of piece hashes and create a slice to hold them
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)
	for i := 0; i < numHashes; i++ {
		// Copy each 20-byte segment into the corresponding hash slot
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}
