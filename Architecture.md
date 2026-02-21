# BitTorrent Architecture

## Overview
BitTorrent is a decentralized file-sharing protocol that breaks large files into pieces and downloads them from multiple peers simultaneously. The process involves metadata parsing, tracker communication, peer discovery, and distributed downloads with integrity verification.

---

## 1. Torrent File Structure (Bencode Format)

A `.torrent` file is encoded in **Bencode format** (not JSON). It contains:

```
Root Dictionary:
├── announce: Tracker URL (used to discover peers)
├── info: Contains file metadata
│   ├── name: File/directory name
│   ├── piece length: Size of each piece (typically 256KB)
│   ├── pieces: Concatenated SHA-1 hashes of all pieces
│   └── length: Total file size
└── Optional fields: comment, created by, creation date
```

**Key Difference:** Unlike REST APIs expecting JSON, BitTorrent uses Bencode—a binary format for efficient serialization.

---

## 2. Tracker Communication

**Purpose:** Bootstrap the download by discovering peers who have the file.

**Flow:**
```
Client → Tracker HTTP GET Request with:
  - info_hash: SHA-1 hash of the info dictionary
  - peer_id: Unique 20-byte ID for this client
  - port: Port listening for incoming connections
  - uploaded: Bytes uploaded so far (for statistics)
  - downloaded: Bytes downloaded so far (for statistics)
  - left: Remaining bytes to download
  - compact: 1 (compact peer list format)

Tracker → Client Response (Bencode):
  - interval: How often to contact tracker
  - peers: Compact list of peers (6 bytes each: 4 IP + 2 port)
```

---

## 3. Peer Connection & Handshake

**Purpose:** Verify that peers are downloading the same torrent and establish a connection.

**Handshake Process:**
```
Client Window                          Peer Window
     │                                    │
     ├─ Create Handshake Message ──────→ │
     │  (19 bytes "BitTorrent protocol"   │
     │   + 8 reserved bytes               │
     │   + 20 bytes info_hash             │
     │   + 20 bytes peer_id)              │
     │                                    │ (Validates info_hash)
     │ ←─────── Respond with Handshake ──┤
     │                                    │
     ✓ Handshake Successful               ✓
```

---

## 4. Interest Signaling

**Purpose:** Tell peers that you want to download pieces (required before requesting data).

**Message:** Send `Interested` (Message ID: 2)

```
Client sends: "I'm interested in downloading pieces from you"
Peer receives: Understands client wants something
```

---

## 5. Choke/Unchoke State Machine

**Purpose:** Implement tit-for-tat reciprocation to prevent free-riding.

**States:**
```
┌─────────────────────────┐
│  Default: CHOKED        │ (Peer refuses all requests)
│  (Peer ID 0)            │
└───────────┬─────────────┘
            │ (Peer decides to reward this client)
            ↓
┌─────────────────────────┐
│  UNCHOKED               │ (Peer allows block requests)
│  (Peer ID 1)            │
│  ✓ Can request blocks   │
└─────────────────────────┘
```

**Reciprocation Logic:**
- Peer tracks upload/download ratio with each client
- Periodically unchokes top uploaders (up to 3-5 clients)
- Unchokes one random peer (for discovery)
- Chokes non-reciprocating peers to prevent free-riding

---

## 6. Bitfield Exchange

**Purpose:** Peers communicate which pieces they have.

**Message:** `Bitfield` (Message ID: 5)

```
Bitfield Format: Array of bytes where each bit = 1 piece
Example (377 bytes = 3016 pieces):
- Byte 0, Bit 0 (MSB) → Piece 0
- Byte 0, Bit 1 → Piece 1
- Byte 0, Bit 7 (LSB) → Piece 7
- Byte 1, Bit 0 → Piece 8
... etc

Value 1 = "I have this piece"
Value 0 = "I don't have this piece"
```

---

## 7. Block-Level Downloading (Pipelining)

**Piece vs. Block:**
- **Piece:** Unit of verification (e.g., 256KB), verified with SHA-1
- **Block:** Unit of transmission (e.g., 16KB), requested individually

**Pipeline Strategy:**
```
Instead of:
Request Block 1 → Wait → Receive Block 1 → Request Block 2 → Wait → Receive Block 2

Do this (Pipeline Depth = 5):
Request Block 1 ─┐
Request Block 2 ├─→ (5 requests in flight)
Request Block 3 ├─→
Request Block 4 ├─→
Request Block 5 ─┘
               ↓
Receive Block 1 → Request Block 6
Receive Block 2 → Request Block 7
Receive Block 3 → Request Block 8
... (maintains 5 concurrent requests)
```

**Benefits:** Masks network latency, maximizes throughput.

---

## 8. Piece Integrity Verification

**Purpose:** Detect corrupted or malicious data.

**Process:**
```
Client receives all blocks for a piece
    ↓
Concatenate blocks in order
    ↓
Calculate SHA-1 hash of concatenated data
    ↓
Compare with expected hash from .torrent file
    ↓
✓ Hash matches → Piece verified, save to disk
✗ Hash fails   → Discard piece, re-queue download
```

---

## 9. File Assembly

**After all pieces downloaded and verified:**
```
Piece 0 (verified) ──┐
Piece 1 (verified) ──┤
Piece 2 (verified) ──├─→ Concatenate in order → Final File
...                  ├─→
Piece N (verified) ──┘
```

---

## Complete Download Flow

```
1. PARSE TORRENT
   Load .torrent file → Bencode decode → Extract metadata

2. TRACKER DISCOVERY
   Contact tracker → Receive peer list

3. PEER HANDSHAKE (per peer)
   Connect to peer → Exchange handshake → Verify info_hash

4. SIGNAL INTEREST
   Send "Interested" message → Indicate intent to download

5. AWAIT UNCHOKE
   Wait for peer to decide to serve us (choke/unchoke decision)

6. RECEIVE BITFIELD
   Peer sends which pieces they have

7. REQUEST BLOCKS
   Pipelined requests (5 concurrent) → Stream blocks from peer
   
8. VERIFY PIECES
   Accumulate blocks → Check SHA-1 hash → Save to disk

9. FETCH & ASSEMBLE
   Parallel downloads from multiple peers → Concatenate pieces → Complete file
```

---

## Key Concepts

| Concept | Purpose |
|---------|---------|
| **Bencode** | Efficient binary encoding for metadata |
| **Info Hash** | SHA-1 of torrent info; identifies the file uniquely |
| **Peer ID** | 20-byte identifier for this client instance |
| **Handshake** | Verify peers want the same file |
| **Choke/Unchoke** | Enforce reciprocation; prevent free-riding |
| **Bitfield** | Communicate piece availability |
| **Pipelining** | Request multiple blocks concurrently |
| **SHA-1 Verification** | Ensure data integrity; detect poisoning |
| **Block Size** | 16KB; unit of transmission |
| **Piece Size** | 256KB (typical); unit of verification |