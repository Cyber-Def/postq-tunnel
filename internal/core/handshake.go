package core

import (
	"encoding/json"
	"io"
)

// ProtocolVersion is embedded in all handshakes to ensure client-server compatibility.
// If the structural communication model changes significantly, we bump this.
const ProtocolVersion = "v1"

// HandshakeReq is sent by the qtun agent to the server over the designated auth stream.
type HandshakeReq struct {
	Version   string `json:"version"`
	Subdomain string `json:"subdomain"`
	Token     string `json:"token"`
	
	// Security constraints injected by the Edge Proxy before reaching the agent
	AllowCIDRs   []string `json:"allow_cidrs,omitempty"`
	BasicAuthUsr string   `json:"basic_auth_usr,omitempty"`
	BasicAuthPwd string   `json:"basic_auth_pwd,omitempty"`
}

// HandshakeResp is the server's reply.
type HandshakeResp struct {
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	AssignedURL string `json:"assigned_url,omitempty"`
}

func WriteHandshake(w io.Writer, req HandshakeReq) error {
	return json.NewEncoder(w).Encode(req)
}

func ReadHandshake(r io.Reader) (HandshakeReq, error) {
	var req HandshakeReq
	// Cap handshake reads at 4KB to prevent JSON memory exhaustion attacks
	limitedReader := io.LimitReader(r, 4096)
	err := json.NewDecoder(limitedReader).Decode(&req)
	return req, err
}

func WriteHandshakeResp(w io.Writer, resp HandshakeResp) error {
	return json.NewEncoder(w).Encode(resp)
}

func ReadHandshakeResp(r io.Reader) (HandshakeResp, error) {
	var resp HandshakeResp
	err := json.NewDecoder(r).Decode(&resp)
	return resp, err
}
