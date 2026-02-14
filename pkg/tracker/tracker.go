package tracker

import (
	"crypto/rand"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
)

// bencodeTrackerResponse matches the format the tracker sends back
type bencodeTrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

// To talk to the tracker, we need to identify ourselves. We do this with a Peer ID.
// In BitTorrent, this is a random 20-byte string
func GeneratePeerID() ([20]byte, error) {
	var id [20]byte
	_, err := rand.Read(id[:])
	if err != nil {
		log.Printf("Failed to generate peer ID: %v", err)
		return [20]byte{}, err
	}
	return id, err
}

// The tracker expects a GET request with several query parameters,
// including the info hash of the torrent, our peer ID, and the port we're listening on.

func BuildTrackerURL(announce string, infoHash [20]byte, peerID [20]byte, port uint16, length int) (string, error) {
	base, err := url.Parse(announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(infoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"}, 
		//compact=1: This tells the tracker to send the peer list in a "compact" binary format 
		//(6 bytes per peer: 4 for IP, 2 for Port) rather than a bulky list. This is standard for modern clients.
		"left":       []string{strconv.Itoa(length)},
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}

func GetPeers(url string) (string, error) {
	// Here we would make an HTTP GET request to the tracker URL and parse the response
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	trackerResp := bencodeTrackerResponse{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return "", err
	}

	return trackerResp.Peers, nil
}
