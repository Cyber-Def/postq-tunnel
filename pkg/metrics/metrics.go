package metrics

import (
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
)

var (
	ActiveTunnels     int64
	TotalBytesRead    uint64
	TotalBytesWritten uint64
	TotalHandshakes   uint64
	TotalHandshakesMs uint64
	TotalNetErrors    uint64
)

// AddActiveTunnel increments or decrements the active tunnel gauge.
func AddActiveTunnel(delta int) {
	atomic.AddInt64(&ActiveTunnels, int64(delta))
}

// AddBytesRead adds to the total bytes read counter
func AddBytesRead(n int) {
	atomic.AddUint64(&TotalBytesRead, uint64(n))
}

// AddBytesWritten adds to the total bytes written counter
func AddBytesWritten(n int) {
	atomic.AddUint64(&TotalBytesWritten, uint64(n))
}

// ObserveHandshake records a handshake and its latency
func ObserveHandshake(latencyMs int64) {
	atomic.AddUint64(&TotalHandshakes, 1)
	atomic.AddUint64(&TotalHandshakesMs, uint64(latencyMs))
}

// AddNetError increments the network error counter
func AddNetError() {
	atomic.AddUint64(&TotalNetErrors, 1)
}

// ProbeMetrics returns a string with Prometheus-formatted metrics
func ProbeMetrics() string {
	active := atomic.LoadInt64(&ActiveTunnels)
	read := atomic.LoadUint64(&TotalBytesRead)
	written := atomic.LoadUint64(&TotalBytesWritten)
	handshakes := atomic.LoadUint64(&TotalHandshakes)
	handshakesMs := atomic.LoadUint64(&TotalHandshakesMs)
	errors := atomic.LoadUint64(&TotalNetErrors)

	return fmt.Sprintf(`
# HELP qtun_active_tunnels The total number of active agent tunnels.
# TYPE qtun_active_tunnels gauge
qtun_active_tunnels %d

# HELP qtun_bytes_read_total Total bytes read from multiplexed streams.
# TYPE qtun_bytes_read_total counter
qtun_bytes_read_total %d

# HELP qtun_bytes_written_total Total bytes written to multiplexed streams.
# TYPE qtun_bytes_written_total counter
qtun_bytes_written_total %d

# HELP qtun_handshakes_total Total number of agent handshakes.
# TYPE qtun_handshakes_total counter
qtun_handshakes_total %d

# HELP qtun_handshake_latency_ms_total Total latency of all handshakes in milliseconds.
# TYPE qtun_handshake_latency_ms_total counter
qtun_handshake_latency_ms_total %d

# HELP qtun_network_errors_total Total network or proxying errors.
# TYPE qtun_network_errors_total counter
qtun_network_errors_total %d
`, active, read, written, handshakes, handshakesMs, errors)
}

// MetricsHandler serves Prometheus metrics
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(ProbeMetrics()))
}

// HealthzHandler is used for liveness probes
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// ReadyzHandler is used for readiness probes
func ReadyzHandler(w http.ResponseWriter, r *http.Request) {
	// Add logic if there's any readiness dependency, otherwise return ok
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

// TrackedConn wraps a net.Conn to track I/O metrics
type TrackedConn struct {
	net.Conn
}

func (c *TrackedConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if n > 0 {
		AddBytesRead(n)
	}
	return
}

func (c *TrackedConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if n > 0 {
		AddBytesWritten(n)
	}
	return
}
