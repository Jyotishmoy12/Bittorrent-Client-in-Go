package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/jyotishmoy12/bittorrent-go/pkg/peer"
)

type Torrent struct {
	Peers       []peer.Peer
	PeerId      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

func (t *Torrent) startDownloadWorker(p peer.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	// Log the connection attempt
	log.Printf("Connecting to peer: %s", p.String())

	conn, err := net.DialTimeout("tcp", p.String(), 5*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()

	// 1. Handshake
	hs := peer.Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: t.InfoHash,
		PeerID:   t.PeerId,
	}
	_, err = conn.Write(hs.Serialize())
	if err != nil {
		log.Printf("Handshake failed with %s: %v", p.String(), err)
		return
	}

	res, err := peer.ReadHandshake(conn)
	if err != nil {
		log.Printf("Error reading handshake from %s: %v", p.String(), err)
		return
	}
	log.Printf("Handshake successful with %s | PeerID: %x", p.String(), res.PeerID[:8])

	// 2. Signal Interest
	interested := peer.Message{ID: 2}
	conn.Write(interested.Serialize())

	// 3. Wait for Unchoke
	log.Printf("Waiting for unchoke from %s...", p.String())
	unchoked := false
	for !unchoked {
		msg, err := peer.Read(conn)
		if err != nil {
			log.Printf("Peer %s disconnected while waiting for unchoke", p.String())
			return
		}
		if msg == nil {
			continue
		}
		if msg.ID == 1 {
			unchoked = true
			log.Printf("Peer %s UNCHOKED us. Starting download...", p.String())
		}
	}

	// 4. Process work
	for pw := range workQueue {
		time.Sleep(1 * time.Second)
		log.Printf("Requesting piece %d (%d bytes) from %s", pw.index, pw.length, p.String())

		buf, err := attemptDownloadPiece(conn, pw)
		if err != nil {
			log.Printf("Download failed for piece %d from %s: %v", pw.index, p.String(), err)
			workQueue <- pw
			return
		}

		hash := sha1.Sum(buf)
		if !bytes.Equal(hash[:], pw.hash[:]) {
			log.Printf("Integrity Check Failed: Piece %d from %s. Expected %x, got %x", pw.index, p.String(), pw.hash, hash)
			workQueue <- pw
			continue
		}

		// Log success
		log.Printf("Piece %d verified from %s", pw.index, p.String())
		results <- &pieceResult{pw.index, buf}
	}
}

func attemptDownloadPiece(c net.Conn, pw *pieceWork) ([]byte, error) {
	c.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.SetDeadline(time.Time{})

	buf := make([]byte, pw.length)
	var requested int
	var downloaded int
	const maxBacklog = 5
	const blockSize = 16384

	for downloaded < pw.length {
		// 1. PIPELINING: Keep 5 requests in flight
		for requested < pw.length && (requested-downloaded) < maxBacklog*blockSize {
			curBlockSize := blockSize
			if pw.length-requested < curBlockSize {
				curBlockSize = pw.length - requested
			}

			// FIXED: Using RequestMessage as defined in your peer package
			msg := peer.RequestMessage(pw.index, requested, curBlockSize)
			_, err := c.Write(msg.Serialize())
			if err != nil {
				return nil, err
			}
			requested += curBlockSize
		}

		// 2. RECEIVE: Read from wire
		msg, err := peer.Read(c)
		if err != nil {
			return nil, err
		}

		if msg == nil || msg.ID != 7 { // 7 is MsgPiece
			continue
		}

		// 3. PARSE & COPY
		begin, data, err := peer.ParsePiece(pw.index, buf, msg)
		if err != nil {
			return nil, err
		}

		copy(buf[begin:], data)
		downloaded += len(data)
	}
	return buf, nil
}

func (t *Torrent) Download() error {
	log.Printf("Starting download for %s (Total size: %d bytes)...", t.Name, t.Length)
	out, err := os.Create(t.Name)
	if err != nil {
		return fmt.Errorf("could not create file: %v", err)
	}
	defer out.Close()
	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)

	for index, hash := range t.PieceHashes {
		begin := index * t.PieceLength
		end := begin + t.PieceLength
		if end > t.Length {
			end = t.Length
		}
		workQueue <- &pieceWork{index, hash, end - begin}
	}

	for _, p := range t.Peers {
		go t.startDownloadWorker(p, workQueue, results)
	}
	doneCount := 0
	for doneCount < len(t.PieceHashes) {
		res := <-results
		time.Sleep(500 * time.Millisecond)
		begin := res.index * t.PieceLength
		_, err := out.WriteAt(res.buf, int64(begin))
		if err != nil {
			log.Printf("Failed to write piece %d to disk: %v", res.index, err)
			continue
		}
		doneCount++

		percent := float64(doneCount) / float64(len(t.PieceHashes)) * 100
		log.Printf("Overall Progress: %.2f%% (%d/%d pieces)", percent, doneCount, len(t.PieceHashes))
	}
	close(workQueue)

	log.Printf("Download complete! File saved as: %s", t.Name)
	return nil
}
