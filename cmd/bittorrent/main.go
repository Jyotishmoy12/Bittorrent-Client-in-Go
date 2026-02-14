package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jyotishmoy12/bittorrent-go/pkg/peer"
	"github.com/jyotishmoy12/bittorrent-go/pkg/torrentfile"
	"github.com/jyotishmoy12/bittorrent-go/pkg/tracker"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/bittorrent/main.go <torrent-file>")
	}
	torrentPath := os.Args[1]

	// 1. Open and parse the .torrent file
	bto, err := torrentfile.Open(torrentPath)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Prepare the Torrent metadata
	infoHash, _ := bto.InfoHash()
	hashes, _ := bto.SplitPieceHashes()
	peerID, _ := tracker.GeneratePeerID()

	// 3. Get Peers from Tracker
	trackerURL, _ := tracker.BuildTrackerURL(bto.Announce, infoHash, peerID, 6881, bto.Info.Length)
	peersBin, err := tracker.GetPeers(trackerURL)
	if err != nil {
		log.Fatal(err)
	}
	peers, _ := peer.Unmarshal(peersBin)

	// 4. Create the Torrent orchestrator
	to := &torrentfile.Torrent{
		Peers:       peers,
		PeerId:      peerID,
		InfoHash:    infoHash,
		PieceHashes: hashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}

	err = to.Download()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nDone! %s has been saved to your current directory.\n", bto.Info.Name)
}
