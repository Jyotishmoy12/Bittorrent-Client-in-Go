# Choke Algorithm

## Overview

The choke algorithm was introduced to guarantee reasonable upload and download reciprocation.

This algorithm is a variant of the **"tit-for-tat"** algorithm, which means: **if you give me something, then I will give you something in return.**

---

## The Free-Rider Problem

**Free-riders** are peers that never upload but always download. They should be penalized to maintain fairness and incentivize reciprocation.

---

## Choking and Interest

### What is Choking?

**Choking** is a temporary refusal to upload data.

We say that **peer A has choked peer B** because peer B is unable to download the file from A, even though peer A can download from peer B.

### Why is Choking Necessary?

1. **TCP Congestion:** Prevents network overload when sending data across many connections simultaneously
2. **Network Abuse Prevention:** Prevents starvation and resource exhaustion

### Periodic Nature

Choking and unchoking are **not perpetual**—they are periodic. Upload happens only while a peer is unchoked. Decisions are re-evaluated periodically.

### What is an Interested Peer?

An **interested peer** is someone who wants a piece that you (the current peer) have.

**Example:** If peer A wants piece `p1` that peer B has, we say that **A is interested in peer B**.

---

## How to Decide Who to Choke and Who to Unchoke?

### The Reciprocation Principle

**Scenario:** Peer 1 is choked by peer 2
- Peer 1 can't download anything from peer 2
- But peer 2 can download from peer 1
- Since peer 2 can download from peer 1, it unchokes peer 1
- Then peer 1 can also download from peer 2

**Core Principle:** 
> **Any peer will upload to peers who give the best download rate—unchoke them first.**

---

## Choke Algorithm for Leechers (Downloaders)

When in leecher state, the choke algorithm is called:

- **Every 10 seconds**
- **Every time a peer leaves the peer set**
- **Every time an unchoked peer becomes interested or loses interest**

### Decision Rules

1. **Top 3 Uploaders (Every 10 seconds)**
   - Order interested remote peers by their download rate to the local peer
   - Unchoke the fastest 3 peers

2. **Regular Unchoke (Active Peer Filtering)**
   - Peers are ordered by their download rate to the local peer
   - Only peers that have sent **at least one block in the last 30 seconds** are considered
   - This guarantees only active peers are unchoked

3. **Optimistic Unchoke (Every 30 seconds)**
   - One additional interested peer is unchoked **at random**
   - This promotes fairness to new/slow peers

4. **Redundancy Prevention**
   - If the optimistic unchoked peer is already in the top 3 fastest peers, another peer is chosen for unchoke

### Summary for Leechers

> We maintain **at most 4 unchoked peers:** 3 based on download rate + 1 random optimistic unchoke

**Analogy:** It's basically a love story! Let's say 3 boys are interested in a girl, and the girl only chooses whoever provides her safety, proper love, and proper internet! Hahaha!

---

## Choke Algorithm for Seeders (Complete File Holders)

When in seeder state (already have the complete file), the choke algorithm is called:

- **Every 10 seconds or every time a peer leaves the peer set**
- **Every time an unchoked peer becomes interested or loses interest**

### Decision Rules

1. **Prioritize Active Peers**
   - Order peers according to **the time they were last unchoked** (not just download speed)
   - **Why?** If we only consider current speed, free-riders with high download speed would dominate. By considering time, we penalize free-riders and rotate opportunities.
   - Also prioritize peers that have **pending block requests**
   - Higher upload rate is given priority

2. **Secondary Ordering**
   - Other unchoked peers are ordered by their **upload rate**

3. **Bootstrap Phase (First 20 seconds)**
   - **Unchoke first 3 peers** → Keep seed unchoked to them
   - **Unchoke 1 peer at random** → Seed random peer for the next 10 seconds

4. **Steady State (After 20 seconds)**
   - **Unchoke first 4 peers** → Keep seed unchoked to them

### Key Difference for Seeders

Seeders do **not** unchoke based purely on upload rate. Instead, they use **the time peers were last unchoked**. Peers in the active peer set are changed regularly, with random peers taking slots from highly-ordered peers to ensure fairness and discourage free-riding.

---

## Summary

| Aspect | Leechers | Seeders |
|--------|----------|---------|
| **Unchoke Evaluation** | Every 10 seconds or on peer events | Every 10 seconds or on peer events |
| **Primary Metric** | Download rate (to local peer) | Time since last unchoked (fairness) |
| **Max Unchoked Peers** | Up to 4 (3 + 1 random) | Up to 4 |
| **Fairness Mechanism** | 1 random optimistic unchoke every 30s | Rotation based on time; random selection |
| **Goal** | Fast downloads + fairness | Prevent free-riding + encourage reciprocation |
