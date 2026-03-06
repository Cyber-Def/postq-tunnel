package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/Cyber-Def/postq-tunnel/pkg/ratelimit"
)

// IPWhitelistMiddleware enforces connection only from allowed CIDR blocks/IPs
func IPWhitelistMiddleware(allowedCIDRs []*net.IPNet, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(allowedCIDRs) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// Extract client IP correctly considering potential reverse proxies in front
		userIPStr, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			userIPStr = r.RemoteAddr
		}
		
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			userIPStr = strings.Split(forwarded, ",")[0]
		}
		
		userIP := net.ParseIP(strings.TrimSpace(userIPStr))
		
		allowed := false
		for _, network := range allowedCIDRs {
			if network.Contains(userIP) {
				allowed = true
				break
			}
		}

		if !allowed {
			http.Error(w, "Forbidden: IP is not within the allowed VPN/Corporate whitelist", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// BasicAuthMiddleware enforces HTTP Basic Auth with brute-force lockout.
// The AuthFailLimiter blocks IPs that exceed the failure threshold.
func BasicAuthMiddleware(expectedUser, expectedPass string, limiter *ratelimit.AuthFailLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if expectedUser == "" && expectedPass == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract client IP for rate limiting
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			clientIP = r.RemoteAddr
		}

		// Check if this IP is currently locked out from too many failures
		if limiter.IsBlocked(clientIP) {
			slog.Warn("Basic Auth: IP blocked due to too many failures", "ip", clientIP)
			http.Error(w, "Too Many Requests: temporarily locked out", http.StatusTooManyRequests)
			return
		}

		user, pass, ok := r.BasicAuth()
		// Prevent timing attacks using ConstantTimeCompare
		userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(expectedUser)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(expectedPass)) == 1

		if !ok || !userMatch || !passMatch {
			limiter.RecordFailure(clientIP)
			w.Header().Set("WWW-Authenticate", `Basic realm="QTUN Protected Tunnel"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SSOMiddleware enforces an Identity-Aware Proxy (IAP) workflow.
// This is a stub for future OIDC / GitHub OAuth integration.
func SSOMiddleware(ssoEnabled bool, loginURL string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !ssoEnabled {
			next.ServeHTTP(w, r)
			return
		}

		// Check for the SSO session cookie (Identity verification)
		cookie, err := r.Cookie("qtunnel_sso_session")
		if err != nil || cookie.Value == "" {
			// Redirect unauthenticated developer/user to the Enterprise IDP (GitHub/Google)
			http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
			return
		}

		// Proceed to local agent
		next.ServeHTTP(w, r)
	})
}
