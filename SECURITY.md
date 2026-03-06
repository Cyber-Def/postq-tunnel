# PostQ-Tunnel Security & Threat Model

This document outlines the security architecture, accepted risks, and threat models for **PostQ-Tunnel**, an experimental post-quantum secure tunneling proxy.

## 1. System Overview

PostQ-Tunnel consists of two major moving parts:
- **Edge Server (qtunnel)**: A public-facing proxy exposed to the Internet, acting as a gateway.
- **Agent (qtun)**: A client deployed on a restricted / local network, dialing out to the Edge Server to proxy traffic securely. 

All traffic between the Agent and the Edge Server is secured at the transport layer using an experimental Post-Quantum TLS 1.3 implementation, leveraging **X25519MLKEM768**.

## 2. Threat Model

We assume the following attacker capabilities:
1. **Network Sniffer (Passive MITM)**: Can intercept, record, and log all ciphertext traveling between the Agent and the Server. The attacker may store the traffic to attempt decryption in the future via a cryptographically relevant quantum computer (CRQC), known as "Store Now, Decrypt Later" (SNDL).
2. **Active Network Attacker (Active MITM)**: Can drop, modify, inject, or delay packets between the Agent and Server.
3. **Malicious Client**: An arbitrary entity on the internet sending malformed HTTP payloads to the Edge Server in an attempt to fingerprint, crash, or exploit the underlying system.
4. **Malicious Agent**: A compromised internal agent trying to exhaust resources on the Edge Server (e.g., initiating infinite streams).

## 3. Security Properties & Mitigations

### 3.1 Post-Quantum Cryptography (PQC)
To defend against SNDL (Store Now, Decrypt Later) attacks, PostQ-Tunnel enforces:
- **Hybrid Key Exchange**: TLS connections are established using `X25519MLKEM768`. It combines the heavily audited classical `X25519` elliptic curve with the standardized lattice-based `ML-KEM-768` (Kyber).
- **Forward Secrecy**: Temporary session keys are derived using standard TLS 1.3 key derivation functions and are strictly meant to be zeroized from memory post-session.

### 3.2 DoS and Resource Exhaustion
- **Multiplexer Constraints**: We rely on Yamux for multiplexing. It is constrained directly by configuration blocks limiting `MaxStreamWindowSize = 1MB` and `StreamCloseTimeout = 5m`. 
- **Timeouts**: Handshakes enforce hard `SetDeadline(10s)` limits to prevent attackers from hanging TLS sockets and holding computational resources hostage.
- **Connection Rate-Limiting**: (Planned / In Development) - Token buckets at the TCP layer to prevent SYN/TLS handshake flooding.

### 3.3 Access Control & Protocol Defenses
- **Agent Mutual Authentication**: Agents must supply a verified authentication token during the multiplexed initial stream (`HandshakeReq`). If invalid, the connection proxy is immediately dropped.
- **IP Whitelisting**: Connections to mounted domains can optionally be restricted to supplied CIDRs by the agent.

## 4. Accepted Risks
- **Cryptographic Novelty**: `ML-KEM` relies on relatively young lattice-based mathematical hardness assumptions. If vulnerabilities are discovered in `ML-KEM`, the system correctly falls back to its `X25519` hybrid pair.
- **Endpoint Compromise**: If the Edge Server or Agent host is fully compromised (root access), all data flowing through them in plaintext is fully disclosed. PQC only secures data *in transit*. 
- **Side-Channel Attacks**: Go's experimental implementations of ML-KEM might eventually be found to be susceptible to certain timing/power-analysis attacks.

## 5. Reporting Vulnerabilities
If you discover a vulnerability, please DO NOT open a public issue. Reach out to the core maintainers privately.
