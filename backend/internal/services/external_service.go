package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
	"pt-dana-sejahtera/internal/security"
)

// AllowedDomains is the whitelist for SSRF protection.
var AllowedDomains = []string{
	"jsonplaceholder.typicode.com",
	"api.github.com",
	"httpbin.org",
}

type FetchRequest struct {
	URL     string            `json:"url"     binding:"required"`
	Headers map[string]string `json:"headers"`
}

type FetchResponse struct {
	StatusCode int             `json:"status_code"`
	Body       json.RawMessage `json:"body"`
	Headers    http.Header     `json:"headers,omitempty"`
}

type ExternalService interface {
	Fetch(req FetchRequest) (*FetchResponse, error)
}

type externalService struct {
	log        *zap.Logger
	httpClient *http.Client
	// Vulnerable client: no timeout, trusts any cert
	unsafeClient *http.Client
}

func NewExternalService(log *zap.Logger) ExternalService {
	return &externalService{
		log: log,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		// TODO: Vulnerability Injection Point — OWASP API #10 (Unsafe Consumption of APIs)
		// Vulnerable client has no timeout and would trust any response
		unsafeClient: &http.Client{},
	}
}

func (s *externalService) Fetch(req FetchRequest) (*FetchResponse, error) {
	if security.IsSecureFor(security.CategoryA10) {
		return s.secureFetch(req)
	}
	// TODO: Vulnerability Injection Point — OWASP API7 / A10 (SSRF)
	// A10 enabled: no URL allowlist — arbitrary internal/external URLs reachable.
	return s.vulnerableFetch(req)
}

// ─── Secure fetch ─────────────────────────────────────────────────────────────

func (s *externalService) secureFetch(req FetchRequest) (*FetchResponse, error) {
	// OWASP #7 Secure: validate URL domain against whitelist
	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow https in production-like scenarios
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return nil, fmt.Errorf("unsupported scheme: %s", parsedURL.Scheme)
	}

	// Block private / internal IP ranges
	hostname := parsedURL.Hostname()
	if isPrivateHost(hostname) {
		s.log.Warn("SSRF attempt blocked — private host", zap.String("host", hostname))
		return nil, fmt.Errorf("access to internal network is forbidden")
	}

	// Whitelist domain check
	if !isAllowedDomain(hostname) {
		s.log.Warn("SSRF attempt blocked — domain not whitelisted", zap.String("host", hostname))
		return nil, fmt.Errorf("domain '%s' is not in the allowed list: %v", hostname, AllowedDomains)
	}

	// OWASP #10 Secure: use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, req.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// OWASP #10 Secure: validate response content-type
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		return nil, fmt.Errorf("unexpected content-type from external API: %s", ct)
	}

	// OWASP #10 Secure: limit response body size (1 MB)
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// OWASP #10 Secure: validate response is valid JSON
	if !json.Valid(bodyBytes) {
		return nil, fmt.Errorf("external API returned invalid JSON")
	}

	s.log.Info("Secure external fetch completed",
		zap.String("url", req.URL),
		zap.Int("status", resp.StatusCode),
	)

	return &FetchResponse{
		StatusCode: resp.StatusCode,
		Body:       bodyBytes,
	}, nil
}

// ─── Vulnerable fetch ─────────────────────────────────────────────────────────

func (s *externalService) vulnerableFetch(req FetchRequest) (*FetchResponse, error) {
	// TODO: Vulnerability Injection Point — OWASP API #7 (SSRF)
	// No URL validation — allows fetching:
	// - Internal services: http://localhost:5432 (DB)
	// - AWS metadata: http://169.254.169.254/latest/meta-data/
	// - Internal microservices: http://internal-api/admin
	s.log.Warn("[VULNERABLE] SSRF: fetching arbitrary URL with no validation",
		zap.String("url", req.URL),
	)

	// TODO: Vulnerability Injection Point — OWASP API #10 (Unsafe Consumption of APIs)
	// No timeout — vulnerable to slowloris / resource exhaustion
	httpReq, err := http.NewRequest(http.MethodGet, req.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Apply any headers caller provides (trusting external input blindly)
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := s.unsafeClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// TODO: Vulnerability Injection Point — OWASP API #10
	// No response body size limit — could exhaust memory
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// TODO: Vulnerability Injection Point — OWASP API #10
	// No content-type or JSON schema validation — blindly returns raw response
	var rawBody json.RawMessage
	if json.Valid(bodyBytes) {
		rawBody = bodyBytes
	} else {
		rawBody, _ = json.Marshal(string(bodyBytes))
	}

	return &FetchResponse{
		StatusCode: resp.StatusCode,
		Body:       rawBody,
		Headers:    resp.Header, // VULN: exposes all response headers to caller
	}, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func isPrivateHost(hostname string) bool {
	// Block loopback and known cloud metadata endpoints
	privatePatterns := []string{
		"localhost", "127.", "0.0.0.0",
		"169.254.",       // AWS/GCP metadata
		"192.168.", "10.", "172.16.", "172.17.", "172.18.",
		"172.19.", "172.20.", "172.21.", "172.22.", "172.23.",
		"172.24.", "172.25.", "172.26.", "172.27.", "172.28.",
		"172.29.", "172.30.", "172.31.",
		"::1", "fc00:", "fd",
	}
	hostname = strings.ToLower(hostname)
	for _, p := range privatePatterns {
		if strings.HasPrefix(hostname, p) || hostname == strings.TrimSuffix(p, ".") {
			return true
		}
	}

	// Also resolve and check IPs
	ips, err := net.LookupHost(hostname)
	if err != nil {
		return false
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return true
		}
	}
	return false
}

func isAllowedDomain(hostname string) bool {
	hostname = strings.ToLower(hostname)
	for _, domain := range AllowedDomains {
		if hostname == domain || strings.HasSuffix(hostname, "."+domain) {
			return true
		}
	}
	return false
}
