# Piece Selection Algorithm

## Overview

The piece selection algorithm determines which pieces a peer should download from the network. This is critical for maximizing download speed, ensuring network resilience, and enabling early participation in the swarm.

---

## Data Hierarchy: File → Pieces → Blocks

### **File**
The original file shared in the network.

### **Pieces**
The file is split into **equal-sized pieces** (typically 256 KB to 1 MB).

- Each piece is identified by a **20-byte SHA-1 hash** stored in the `.torrent` file
- Pieces are the unit of verification
- Must be downloaded completely before verification

### **Blocks**
Each piece is further divided into **blocks** (typically 16 KB).

- Blocks are the unit of transmission
- A peer requests blocks individually from peers in the network
- Allows parallel downloading and pipelining

---

## Piece Verification Flow

```
1. Download Piece 1 
   ↓
   Calculate SHA-1(data) 
   ↓
   Compare with hash1 from .torrent
   ↓
   ✓ Match? → Save piece 1
   ✗ No match? → Discard and re-download

2. Download Piece 2 
   ↓
   Calculate SHA-1(data) 
   ↓
   Compare with hash2 from .torrent
   ↓
   ✓ Match? → Save piece 2
   ✗ No match? → Discard and re-download

3. ... (Repeat for all pieces)

4. Piece N 
   ↓
   Calculate SHA-1(data) 
   ↓
   Compare with hashN from .torrent
   ↓
   ✓ Match? → Save piece N

5. Concatenate all verified pieces → Complete file ✅
   (No additional verification needed—each piece already verified)
```

---

## The Seeder Dependency Problem

### **What if the seeder leaves the network?**
→ **The download of the entire network would stop!!!**

This creates a critical bottleneck. To prevent this, we must:

1. **Maximize download speed** (finish quickly)
2. **Minimize dependency on seeders** (get pieces from multiple peers)
3. **Diversify piece selection** (don't request the same pieces from one peer)

---

## Rarest-First Piece Selection Algorithm

### **Core Idea**
**Prioritize downloading pieces that are rarest in the network.**

By downloading rare pieces first, we:
- Reduce the risk of losing pieces when peers leave
- Ensure piece availability in the swarm
- Help distribute rare pieces back to the network

### **How Does a Peer Compute the Rarest Piece?**

1. **Collect peer information**
   - Every peer maintains a **peer set** obtained from the tracker
   - Query the bitfield/have messages from all peers in the set

2. **Two Ways Peers Announce Their Pieces**

   **a) "Have" Message (Incremental)**
   - A peer sends a series of "Have" messages, one for each piece they have
   - Useful for announcing newly acquired pieces
   - Example: Peer B and C send: "have piece 5", "have piece 12", etc.

   **b) "Bitfield" Message (Snapshot)**
   - At the start of connection, a peer sends a single bitfield message
   - Bits marked as 1 = peer has that piece
   - Bits marked as 0 = peer doesn't have that piece
   - More efficient than multiple "Have" messages

3. **Count piece availability**
   - Count how many peers have each piece
   - Piece with the lowest count = **rarest piece**
   - Download rarest piece first

---

## Random-First Policy

### **The Problem**
When a new peer joins the network, it has zero pieces. Peers use **choke/unchoke** to incentivize reciprocation—they only upload to peers who have something to upload back.

If we immediately request rare pieces:
- Rare pieces are slower to download (fewer peers have them)
- We take too long to accumulate initial pieces
- We can't reciprocate and get choked

### **The Solution: Random-First Policy**

**Rule:** If we have downloaded **fewer than 4 pieces**, request the next piece **at random**.

**Benefits:**
- Quickly accumulate initial pieces (in parallel from different peers)
- Enable early contribution to the network
- Demonstrate value for reciprocation (choke/unchoke)
- Once we have 4 pieces, switch to rarest-first

**Transition:**
```
Pieces Downloaded: [0, 3] → Random-First Selection
Pieces Downloaded: [4+] → Switch to Rarest-First Selection
```

---

## Strict Priority Policy

### **The Problem**
A peer **cannot contribute back to the network** until it has **at least one complete piece**.

If we download blocks haphazardly (jumping between pieces), we may never complete a single piece quickly.

### **The Solution: Strict Priority Policy**

**Rule:** Always prioritize completing one piece at a time.

**Process:**
1. Select a piece (using random-first or rarest-first)
2. Download **all blocks** of that piece from peers
3. **Complete and verify** the piece
4. **Only then** move to the next piece

**Why This Matters:**
- Ensures we complete pieces quickly
- Enables early reciprocation (upload pieces once verified)
- Prevents fragmentation across many incomplete pieces
- Peers will unchoke us once they see we're contributing

---

## End Game Mode

### **When Does It Start?**

End Game Mode activates at the **very end of the download**:
- Peer has requested **all remaining blocks** it needs
- Requests are **in transit** (awaiting responses)
- Download is about to complete
- Peer is just waiting for final responses to arrive

### **What Happens?**

**Aggressive request strategy:**
1. Send duplicate block requests to **all peers** in the peer set
2. Whoever responds first with the block "wins"
3. **Every time a block is received**, send a **Cancel** message to other peers
   - Prevents waste of bandwidth from redundant transfers

### **Why This Works:**
- **Minimize latency:** Don't wait for one slow peer
- **Maximize throughput:** Get blocks from whoever responds fastest
- **Broadcast cancellations:** Stop redundant downloads immediately
- **Complete download quickly:** Rush the final pieces

### **Example:**

```
Need blocks: [100, 101, 102, 103]

Send request for block 100 → Peer A, Peer B, Peer C, Peer D
Send request for block 101 → Peer A, Peer B, Peer C, Peer D
Send request for block 102 → Peer A, Peer B, Peer C, Peer D
Send request for block 103 → Peer A, Peer B, Peer C, Peer D

Peer B responds with block 100 first
  → Accept block 100
  → Send Cancel(100) to Peer A, C, D
  
Peer D responds with block 101 first
  → Accept block 101
  → Send Cancel(101) to Peer A, B, C

... continue until all blocks received
```

---

## Summary Table

| Algorithm | When Used | Goal | Strategy |
|-----------|-----------|------|----------|
| **Random-First** | First 4 pieces | Bootstrap; Enable reciprocation | Choose pieces at random |
| **Rarest-First** | After 4 pieces | Preserve rare pieces; Maximize network resilience | Download scarcest pieces first |
| **Strict Priority** | All time | Complete pieces quickly | Finish one piece before starting another |
| **End Game Mode** | Final pieces | Minimize latency on last blocks | Request from all peers; Cancel redundant requests |
