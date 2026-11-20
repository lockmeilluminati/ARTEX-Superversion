package integrations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxHealthResponseBytes = 64 * 1024

type HTTPProbe struct {
	Endpoint string
	Client   *http.Client
}

type healthPayload struct {
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}

// Check performs a read-only GET against an explicitly configured loopback URL.
// No endpoint path, port, or response schema is assumed by ARTEX.
func (p HTTPProbe) Check(ctx context.Context, id, name string) Health {
	result := Health{IntegrationID: id, Name: name, State: StateUnknown, CheckedAt: time.Now().UTC()}
	u, err := url.Parse(strings.TrimSpace(p.Endpoint))
	if err != nil || u == nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		result.State, result.Diagnostic = StateMisconfigured, "health endpoint must be an absolute HTTP(S) URL without embedded credentials"
		return result
	}
	if !isLoopbackHost(u.Hostname()) {
		result.State, result.Diagnostic = StateMisconfigured, "health endpoint must resolve to localhost or a loopback IP"
		return result
	}

	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	// Do not follow redirects: a local service must not redirect ARTEX to another host.
	safeClient := *client
	safeClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		result.State, result.Diagnostic = StateMisconfigured, "could not construct health request"
		return result
	}
	resp, err := safeClient.Do(req)
	if err != nil {
		result.State, result.Diagnostic = StateUnavailable, "local service did not respond: " + safeDiagnostic(err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		result.State, result.Diagnostic = StateMisconfigured, "local service requires authentication; configure credentials after verifying its documented interface"
		return result
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.State, result.Diagnostic = StateUnavailable, fmt.Sprintf("health endpoint returned HTTP %d", resp.StatusCode)
		return result
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxHealthResponseBytes+1))
	if err != nil {
		result.State, result.Diagnostic = StateUnavailable, "could not read health response"
		return result
	}
	if len(body) > maxHealthResponseBytes {
		result.State, result.Diagnostic = StateUnavailable, "health response exceeded the 64 KiB safety limit"
		return result
	}
	var payload healthPayload
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			result.State, result.Diagnostic = StateAvailable, "endpoint responded successfully; response was not the optional ARTEX health JSON format"
			return result
		}
	}
	result.State = StateAvailable
	result.Version = payload.Version
	result.Capabilities = payload.Capabilities
	result.Diagnostic = "local endpoint responded successfully"
	return result
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func safeDiagnostic(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "request timed out"
	}
	if errors.Is(err, context.Canceled) {
		return "request canceled"
	}
	// Avoid returning raw URLs or request details that may contain local configuration.
	return "connection failed"
}
