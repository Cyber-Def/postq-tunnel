// Package ratelimit provides IP-based rate limiting for incoming agent connections.
//
// Design: Per-IP token bucket with a fixed burst window. Each unique IP address
// is allowed up to MaxHandshakesPerSecond new connection attempts per second.
// This is intentionally generous (default: 10/s) so no real user is ever blocked
// by normal reconnect or multi-tunnel workflows.
package ratelimit

import (
	"net"
	"sync"
	"time"
)

const (
	// MaxHandshakesPerSecond is the maximum number of new agent connections
	// allowed from a single IP address per second.
	MaxHandshakesPerSecond = 10

	// cleanupInterval is how often we evict idle/expired IP entries.
	cleanupInterval = 5 * time.Minute
)

// ipEntry tracks connection count and the timestamp of the current window.
type ipEntry struct {
	count     int
	windowEnd time.Time
}

// HandshakeLimiter limits incoming connections per IP.
type HandshakeLimiter struct {
	mu      sync.Mutex
	entries map[string]*ipEntry
}

// NewHandshakeLimiter creates a new HandshakeLimiter and starts background cleanup.
func NewHandshakeLimiter() *HandshakeLimiter {
	rl := &HandshakeLimiter{
		entries: make(map[string]*ipEntry),
	}
	go rl.cleanup()
	return rl
}

// Allow returns true if the request from the given address is within the rate limit.
func (rl *HandshakeLimiter) Allow(addr string) bool {
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		ip = addr // fallback: treat the raw string as ip
	}

	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.entries[ip]
	if !exists || now.After(entry.windowEnd) {
		// Start a fresh 1-second window
		rl.entries[ip] = &ipEntry{count: 1, windowEnd: now.Add(time.Second)}
		return true
	}

	entry.count++
	return entry.count <= MaxHandshakesPerSecond
}

// cleanup periodically removes stale entries to prevent unbounded memory growth.
func (rl *HandshakeLimiter) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		rl.mu.Lock()
		for ip, entry := range rl.entries {
			if now.After(entry.windowEnd) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// --- Basic Auth Lockout ---

const (
	// authFailWindow is the sliding window for counting failed auth attempts.
	authFailWindow = 30 * time.Second
	// authFailThreshold is the number of failures within the window that triggers lockout.
	authFailThreshold = 20
	// authLockoutDuration is how long a misbehaving IP is blocked after crossing the threshold.
	authLockoutDuration = 1 * time.Minute
)

// authEntry tracks failed auth attempts and whether the IP is currently blocked.
type authEntry struct {
	failures  int
	windowEnd time.Time
	blockedUntil time.Time
}

// AuthFailLimiter blocks IPs that repeatedly fail Basic Auth challenges.
type AuthFailLimiter struct {
	mu      sync.Mutex
	entries map[string]*authEntry
}

// NewAuthFailLimiter creates a new AuthFailLimiter and starts background cleanup.
func NewAuthFailLimiter() *AuthFailLimiter {
	al := &AuthFailLimiter{entries: make(map[string]*authEntry)}
	go al.cleanup()
	return al
}

// IsBlocked returns true if the given IP is currently locked out.
func (al *AuthFailLimiter) IsBlocked(ip string) bool {
	al.mu.Lock()
	defer al.mu.Unlock()
	e, ok := al.entries[ip]
	if !ok {
		return false
	}
	return time.Now().Before(e.blockedUntil)
}

// RecordFailure records a failed auth attempt for the given IP.
// If the failure count crosses authFailThreshold within authFailWindow,
// the IP is locked out for authLockoutDuration.
func (al *AuthFailLimiter) RecordFailure(ip string) {
	now := time.Now()
	al.mu.Lock()
	defer al.mu.Unlock()

	e, ok := al.entries[ip]
	if !ok || now.After(e.windowEnd) {
		al.entries[ip] = &authEntry{failures: 1, windowEnd: now.Add(authFailWindow)}
		return
	}
	e.failures++
	if e.failures >= authFailThreshold {
		e.blockedUntil = now.Add(authLockoutDuration)
	}
}

// cleanup removes stale auth entries.
func (al *AuthFailLimiter) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		al.mu.Lock()
		for ip, e := range al.entries {
			if now.After(e.windowEnd) && now.After(e.blockedUntil) {
				delete(al.entries, ip)
			}
		}
		al.mu.Unlock()
	}
}
