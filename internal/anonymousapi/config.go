// Package anonymousapi implements the bounded, intentionally credential-free API contract.
package anonymousapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Profile struct {
	ID               string   `json:"id"`
	AgentID          string   `json:"agent_id"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids"`
	MaxTokens        int      `json:"max_tokens"`
}
type Config struct {
	Enabled           bool      `json:"enabled"`
	TenantID          uint64    `json:"tenant_id"`
	UserID            string    `json:"user_id"`
	AllowedCIDRs      []string  `json:"allowed_cidrs"`
	TrustedProxyCIDRs []string  `json:"trusted_proxy_cidrs"`
	GlobalConcurrency int       `json:"global_concurrency"`
	IPConcurrency     int       `json:"ip_concurrency"`
	RequestsPerMinute int       `json:"requests_per_minute"`
	TimeoutSeconds    int       `json:"timeout_seconds"`
	Profiles          []Profile `json:"profiles"`
}

var alias = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$`)

func Load(path string) (*Config, error) {
	if path == "" {
		return &Config{}, nil
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	var c Config
	d := json.NewDecoder(io.LimitReader(f, 1<<20))
	d.DisallowUnknownFields()
	if e = d.Decode(&c); e != nil {
		return nil, e
	}
	if d.Decode(new(any)) != io.EOF {
		return nil, errors.New("invalid trailing config data")
	}
	if e = c.Validate(); e != nil {
		return nil, e
	}
	return &c, nil
}
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.TenantID == 0 || c.UserID == "" || len(c.Profiles) == 0 {
		return errors.New("tenant, dedicated user and profiles are required")
	}
	if c.GlobalConcurrency == 0 {
		c.GlobalConcurrency = 4
	}
	if c.IPConcurrency == 0 {
		c.IPConcurrency = 2
	}
	if c.RequestsPerMinute == 0 {
		c.RequestsPerMinute = 10
	}
	if c.TimeoutSeconds == 0 {
		c.TimeoutSeconds = 300
	}
	if c.GlobalConcurrency < 1 || c.GlobalConcurrency > 64 || c.IPConcurrency < 1 || c.IPConcurrency > c.GlobalConcurrency || c.RequestsPerMinute < 1 || c.RequestsPerMinute > 10000 || c.TimeoutSeconds < 1 || c.TimeoutSeconds > 300 {
		return errors.New("invalid execution limits")
	}
	for _, s := range append(append([]string{}, c.AllowedCIDRs...), c.TrustedProxyCIDRs...) {
		if _, _, e := net.ParseCIDR(s); e != nil {
			return fmt.Errorf("invalid CIDR: %s", s)
		}
	}
	seen := map[string]bool{}
	for i := range c.Profiles {
		p := &c.Profiles[i]
		if !alias.MatchString(p.ID) || seen[p.ID] || p.AgentID == "" || len(p.KnowledgeBaseIDs) == 0 {
			return errors.New("invalid or duplicate profile")
		}
		seen[p.ID] = true
		if p.MaxTokens == 0 {
			p.MaxTokens = 2048
		}
		if p.MaxTokens < 1 || p.MaxTokens > 12288 {
			return errors.New("invalid output cap")
		}
		ks := map[string]bool{}
		for _, k := range p.KnowledgeBaseIDs {
			if k == "" || ks[k] {
				return errors.New("invalid knowledge scope")
			}
			ks[k] = true
		}
	}
	return nil
}
func inCIDRs(ip net.IP, cidrs []string) bool {
	for _, s := range cidrs {
		_, n, e := net.ParseCIDR(s)
		if e == nil && n.Contains(ip) {
			return true
		}
	}
	return false
}

// ClientIP only trusts forwarded headers when every traversed proxy is configured.
func (c *Config) ClientIP(r *http.Request) (net.IP, error) {
	host, _, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, errors.New("invalid peer")
	}
	if inCIDRs(ip, c.TrustedProxyCIDRs) {
		parts := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		for i := len(parts) - 1; i >= 0 && inCIDRs(ip, c.TrustedProxyCIDRs); i-- {
			next := net.ParseIP(strings.TrimSpace(parts[i]))
			if next == nil {
				return nil, errors.New("invalid proxy chain")
			}
			ip = next
		}
	}
	if len(c.AllowedCIDRs) == 0 {
		if !ip.IsLoopback() {
			return nil, errors.New("loopback only")
		}
	} else if !inCIDRs(ip, c.AllowedCIDRs) {
		return nil, errors.New("network not allowed")
	}
	return ip, nil
}
func (c *Config) Profile(id string) (Profile, bool) {
	for _, p := range c.Profiles {
		if p.ID == id {
			return p, true
		}
	}
	return Profile{}, false
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type Request struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
	MaxTokens *int      `json:"max_tokens,omitempty"`
}

func Decode(r io.Reader) (Request, error) {
	var v Request
	d := json.NewDecoder(io.LimitReader(r, 512*1024+1))
	d.DisallowUnknownFields()
	if e := d.Decode(&v); e != nil {
		return v, e
	}
	if d.Decode(new(any)) != io.EOF {
		return v, errors.New("trailing request data")
	}
	if v.Model == "" || len(v.Messages) == 0 || len(v.Messages) > 100 {
		return v, errors.New("model and 1..100 messages required")
	}
	size := 0
	for i, m := range v.Messages {
		size += len(m.Content)
		if m.Role != "user" && m.Role != "assistant" {
			return v, errors.New("only user/assistant text messages supported")
		}
		if strings.TrimSpace(m.Content) == "" {
			return v, errors.New("empty message")
		}
		if i == 0 && m.Role != "user" {
			return v, errors.New("history must begin with user")
		}
		if i > 0 && m.Role == v.Messages[i-1].Role {
			return v, errors.New("history must alternate user and assistant")
		}
	}
	if size > 256*1024 {
		return v, errors.New("messages exceed 256 KiB")
	}
	if v.Messages[len(v.Messages)-1].Role != "user" {
		return v, errors.New("last message must be user")
	}
	if v.MaxTokens != nil && (*v.MaxTokens < 1 || *v.MaxTokens > 12288) {
		return v, errors.New("invalid max_tokens")
	}
	return v, nil
}
