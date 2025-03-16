This document is a step-by-step guide that not only explains how to configure **netem** and benchmark **HTTP/3 vs. HTTP/2**, but also clarifies *how each netem impairment typically impacts TCP vs. QUIC* at a high level.  

Feel free to adjust specifics (interface names, ports, tools) to match your own environment.

---

# Benchmarking HTTP/3 vs. HTTP/2 Under Adverse Network Conditions

This document outlines how to use **netem** (the Traffic Control network emulator) to simulate degraded network conditions—latency, packet loss, jitter, etc.—and then benchmark HTTP/3 versus HTTP/2 to assess performance differences. It also includes notes on *how each netem rule can affect TCP-based (HTTP/2) vs. QUIC-based (HTTP/3) traffic.*

## Table of Contents

1. [Introduction](#introduction)  
2. [Prerequisites](#prerequisites)  
3. [Setting Up Netem](#setting-up-netem)  
   - [Adding Latency](#adding-latency)  
   - [Introducing Packet Loss](#introducing-packet-loss)  
   - [Adding Jitter](#adding-jitter)  
   - [Bandwidth Throttling / Rate Control](#bandwidth-throttling--rate-control)  
   - [Reordering and Corruption (Advanced)](#reordering-and-corruption-advanced)  
   - [Combining Multiple Conditions](#combining-multiple-conditions)  
   - [Clearing Netem Rules](#clearing-netem-rules)  
4. [Benchmarking Strategies](#benchmarking-strategies)  
   - [Tooling](#tooling)  
   - [Test Scenarios](#test-scenarios)  
   - [Metrics to Track](#metrics-to-track)  
5. [Example Workflow](#example-workflow)  
6. [Analysis Tips](#analysis-tips)  
7. [Further Resources](#further-resources)  

---

## 1. Introduction

**HTTP/3** (based on QUIC) is designed to improve upon some limitations of HTTP/2, particularly under unreliable or high-latency networks, by:
- Reducing head-of-line blocking.  
- Handling packet loss more gracefully.  
- Offering faster handshake/connection establishment.

Using **netem**, you can simulate different network impairments and see how HTTP/3 performance compares to HTTP/2 under the same conditions. Where relevant, we’ll note how impairments typically affect TCP vs. QUIC at a protocol level.

---

## 2. Prerequisites

- A Linux environment with `tc` (Traffic Control) installed (part of the `iproute2` package).  
- A test server capable of serving **HTTP/2** and **HTTP/3** (e.g., a Go server configured with both protocols, Nginx with QUIC patches, Caddy, or another HTTP/3-capable server).  
- A test client environment (could be the same machine or a separate one) where you can run HTTP/2/HTTP/3 requests.  
- Sufficient permissions (root or `sudo`) to configure `tc`.  

> **Note**: Some tools may need special flags for HTTP/3 (`curl --http3`, `h2load --h3`), while HTTP/2 can be tested with `curl --http2`, `h2load`, etc.

---

## 3. Setting Up Netem

Netem is configured via **qdiscs** (queueing disciplines). Typically you’ll apply these rules on one side of the connection (e.g., your test client) to degrade outgoing traffic. In some cases, you might apply it on both ends or on the server side—depends on your testing strategy.

Replace `eth0` with your actual network interface name (e.g., `ens33`, `wlan0`, etc.).

### a. Adding Latency

Simulate a constant delay (e.g., 100ms):

```bash
sudo tc qdisc add dev eth0 root netem delay 100ms
```

- All outbound packets will be delayed by 100ms.

#### How This Affects TCP (HTTP/2)

- TCP’s congestion control and retransmission timers will factor in the increased RTT.  
- High latency means slower ramp-up of TCP window size; any lost segment also takes longer to recover.  
- HTTP/2’s streams still share one underlying TCP connection, so if the single connection is delayed, *all* streams see the impact of that higher latency.

#### How This Affects QUIC (HTTP/3)

- QUIC also feels the higher RTT, but its initial handshake can be faster than TCP+TLS.  
- QUIC’s congestion control adapts similarly to TCP, but QUIC avoids some head-of-line blocking issues inherent to TCP.  
- Delays still reduce throughput overall, but each stream recovers more independently under QUIC.

---

### b. Introducing Packet Loss

Simulate a 5% packet loss:

```bash
sudo tc qdisc change dev eth0 root netem loss 5%
```

- 1 out of 20 packets (on average) will be dropped.

#### How This Affects TCP (HTTP/2)

- Loss triggers retransmissions at the TCP layer, impacting *all* streams within the HTTP/2 connection.  
- If a packet is lost, subsequent TCP segments can’t be processed by the receiver until the missing segment is recovered (head-of-line blocking at the TCP layer).  
- HTTP/2’s multiple streams share the single TCP connection, so one lost TCP segment can stall data for all streams in that connection.

#### How This Affects QUIC (HTTP/3)

- QUIC handles loss at the individual packet level on a per-stream basis.  
- While loss still reduces throughput, QUIC’s design mitigates the “all streams stall” scenario.  
- QUIC’s faster detection and repair of lost packets can improve performance under moderate loss compared to TCP.

---

### c. Adding Jitter

Simulate jitter around an average delay. For example, a 100ms average with a 20ms variation:

```bash
sudo tc qdisc change dev eth0 root netem delay 100ms 20ms distribution normal
```

- Netem uses a normal distribution centered at 100ms with ±20ms variation.

#### How This Affects TCP (HTTP/2)

- Variable RTT can disrupt TCP congestion control, which tries to estimate round-trip time for retransmissions and flow control.  
- Spiky jitter can lead TCP to occasionally assume congestion and back off.  
- HTTP/2 streams share the same TCP connection, so jitter-induced slowdown or retransmit events can affect all streams.

#### How This Affects QUIC (HTTP/3)

- QUIC’s congestion control also reacts to RTT variation, but QUIC can respond more flexibly to out-of-order packets.  
- If jitter causes out-of-order arrivals, QUIC can still process unaffected streams, reducing overall head-of-line blocking.

---

### d. Bandwidth Throttling / Rate Control

Limit the maximum throughput (e.g., 1Mbps):

```bash
sudo tc qdisc change dev eth0 root netem rate 1mbit
```

- Emulates a low-bandwidth environment.

#### How This Affects TCP (HTTP/2)

- TCP connection throughput is capped by the netem limit.  
- HTTP/2 streams share that capped bandwidth; each stream must contend within the same TCP pipe.  
- If you have many concurrent streams, they all compete for the single 1Mbps limit.

#### How This Affects QUIC (HTTP/3)

- Similar 1Mbps cap, but QUIC can schedule data among streams more flexibly.  
- However, if the cap is the primary bottleneck, QUIC and TCP may end up showing similar throughput—both are limited by raw bandwidth.  
- In scenarios with concurrent streams, QUIC can sometimes more effectively avoid per-stream blocking overhead.

---

### e. Reordering and Corruption (Advanced)

**Packet Reordering**:

```bash
sudo tc qdisc change dev eth0 root netem delay 100ms reorder 25% 50%
```
- 25% of packets are reordered; of those, 50% are re-queued, causing out-of-order delivery.

**Packet Corruption**:

```bash
sudo tc qdisc change dev eth0 root netem corrupt 1%
```
- 1% chance of flipping bits in packets.

#### How This Affects TCP (HTTP/2)

- **Reordering**: TCP must still deliver data in order, so out-of-order packets can cause the receiver to wait for the missing segments.  
- **Corruption**: Corrupted segments fail checksum and trigger retransmissions, potentially stalling all streams.

#### How This Affects QUIC (HTTP/3)

- **Reordering**: QUIC can handle out-of-order packets more gracefully at the transport layer, but extreme reordering can still harm performance.  
- **Corruption**: QUIC also detects corruption (via its own checks) and retransmits. Similar to TCP, corruption leads to retransmissions, but QUIC’s per-stream approach can isolate the impact.

Use these advanced options with caution as they can cause unpredictable, hard-to-debug behavior.

---

### f. Combining Multiple Conditions

You can combine multiple parameters in a single command, e.g.:

```bash
sudo tc qdisc change dev eth0 root netem delay 100ms 20ms loss 2% corrupt 0.1% rate 2mbit
```

Meaning:  
1. Average delay = 100ms with ±20ms jitter.  
2. 2% packet loss.  
3. 0.1% chance of bit corruption.  
4. Bandwidth cap = 2 Mbit/s.

**TCP vs. QUIC**:  
- Multiple impairments often amplify each other, and the differences between TCP and QUIC become more apparent at moderate to high network degradation.  
- HTTP/3’s design typically helps mitigate head-of-line blocking, but extremely harsh conditions will affect *all* transport protocols.

---

### g. Clearing Netem Rules

After you finish testing, remove the netem qdisc to restore normal network conditions:

```bash
sudo tc qdisc del dev eth0 root netem
```

---

## 4. Benchmarking Strategies

### a. Tooling

- **curl**: Good for single-shot requests, can do HTTP/2 (`--http2`) or HTTP/3 (`--http3`).  
- **h2load** (from the nghttp2 project): Excellent for load testing HTTP/2. Some builds support HTTP/3 with `--h3`.  
- **wrk / hey**: Common HTTP benchmarking tools for concurrency tests (mostly for HTTP/1.1 or HTTP/2). For HTTP/3, you might need specialized forks.  
- **Go custom clients**: If you’re writing your own Go program, ensure you have libraries that support HTTP/3 (e.g., `qtls` or Go’s experimental HTTP/3 packages).

### b. Test Scenarios

1. **Baseline (No Netem)**  
   - Measure performance under ideal conditions.  
   - Compare HTTP/2 vs. HTTP/3 throughput, latency, etc.

2. **High Latency**  
   - Add 100–200ms delay.  
   - Evaluate how each protocol deals with large RTTs.

3. **Packet Loss**  
   - Introduce 1–5% loss.  
   - Observe how QUIC (HTTP/3) handles lost packets vs. HTTP/2’s TCP-based approach.

4. **Jitter**  
   - Introduce variable delay.  
   - Check if HTTP/3 experiences less performance degradation due to head-of-line blocking in HTTP/2.

5. **Combination (Latency + Loss + Jitter)**  
   - Simulate real mobile or congested networks.  
   - Evaluate the robustness of each protocol under multiple simultaneous impairments.

### c. Metrics to Track

- **Request Latency**: Time from request start to first byte / last byte received.  
- **Throughput / RPS**: Requests per second (under concurrency).  
- **Connection Establishment Time**: Handshake time (QUIC’s 0-RTT or 1-RTT handshake can outperform TCP+TLS in some conditions).  
- **Error Rates**: How many requests fail due to timeouts or other errors.  

---

## 5. Example Workflow

1. **Start Your Servers**  
   - HTTP/2 server on port 8443.  
   - HTTP/3 server on port 8444 (or the same server, different configuration).

2. **Apply Netem**  
   ```bash
   sudo tc qdisc add dev eth0 root netem delay 100ms loss 2%
   ```
   > Adds a 100ms delay and 2% packet loss for outbound traffic.

3. **Run Benchmarks**  
   - **HTTP/2** (example using `h2load`):  
     ```bash
     h2load --rate=100 --requests=1000 --threads=4 https://server-ip:8443/
     ```
   - **HTTP/3** (using a tool supporting HTTP/3, e.g., `curl --http3` or `h2load --h3` if compiled with QUIC):  
     ```bash
     curl --http3 --cacert server.crt https://server-ip:8444/
     ```
   - Record throughput, latency, error rates, etc.

4. **Collect Metrics**  
   - Note average latency, request success/failure, concurrency scaling, etc.  

5. **Change Netem Conditions**  
   ```bash
   sudo tc qdisc change dev eth0 root netem delay 300ms loss 2%
   ```
   - Increase delay to 300ms for a higher-latency test.  
   - Re-run benchmarks for HTTP/2 and HTTP/3.  

6. **Remove Netem**  
   ```bash
   sudo tc qdisc del dev eth0 root netem
   ```
   - Restore normal network conditions.

---

## 6. Analysis Tips

- **Plot Your Data**: Create graphs showing how throughput or latency changes as you vary loss/delay.  
- **Look for Inflection Points**: Where does HTTP/3 begin outperforming HTTP/2? Is there a threshold of loss/latency?  
- **Consider Real-World Use Cases**: If targeting mobile or WAN environments, match netem parameters to realistic jitter, loss, and latency.  

---

## 7. Further Resources

- **Netem Documentation**:  
  [Linux Foundation Netem](https://www.linux.org/docs/man8/tc-netem.html) (man pages and examples)
- **HTTP/3 and QUIC**:  
  - [QUIC RFC 9000](https://datatracker.ietf.org/doc/html/rfc9000)  
  - [IETF HTTP/3 RFC 9114](https://datatracker.ietf.org/doc/html/rfc9114)  
- **HTTP/2**:  
  - [HTTP/2 RFC 7540](https://datatracker.ietf.org/doc/html/rfc7540)  

---

## Final Notes

By combining **netem** impairments with solid **benchmarking tools**, you can gain a clear picture of **HTTP/3**’s resilience and performance benefits over **HTTP/2** under various real-world-like network conditions. Understanding *how TCP vs. QUIC reacts* to latency, jitter, and packet loss helps you make data-driven decisions on protocol adoption and optimization.