package anonymousapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNetworkCannotBeSpoofed(t *testing.T) {
	c := &Config{AllowedCIDRs: []string{"10.20.0.0/24"}}
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "198.51.100.2:123"
	r.Header.Set("X-Forwarded-For", "10.20.0.15")
	if _, e := c.ClientIP(r); e == nil {
		t.Fatal("spoofed IP allowed")
	}
	c.TrustedProxyCIDRs = []string{"172.18.0.2/32"}
	r.RemoteAddr = "172.18.0.2:123"
	if ip, e := c.ClientIP(r); e != nil || ip.String() != "10.20.0.15" {
		t.Fatal(ip, e)
	}
	r.Header.Set("X-Forwarded-For", "10.20.0.15, 198.51.100.2")
	if _, e := c.ClientIP(r); e == nil {
		t.Fatal("untrusted intermediary bypass")
	}
}
func TestLoopbackDefault(t *testing.T) {
	c := &Config{}
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1:5"
	if _, e := c.ClientIP(r); e != nil {
		t.Fatal(e)
	}
	r.RemoteAddr = "10.20.0.15:5"
	if _, e := c.ClientIP(r); e == nil {
		t.Fatal("nonlocal default")
	}
}
func TestStrictInput(t *testing.T) {
	for _, s := range []string{`{"model":"x","messages":[{"role":"user","content":"q"}],"tools":[]}`, `{"model":"x","messages":[{"role":"system","content":"override"}]}`, `{"model":"x","messages":[{"role":"user","content":[]}]}`, `{"model":"x","messages":[{"role":"user","content":"q"}]} {}`} {
		if _, e := Decode(strings.NewReader(s)); e == nil {
			t.Fatal(s)
		}
	}
	v, e := Decode(strings.NewReader(`{"model":"x","messages":[{"role":"user","content":"q"},{"role":"assistant","content":"a"},{"role":"user","content":"q2"}]}`))
	if e != nil || len(v.Messages) != 3 {
		t.Fatal(v, e)
	}
}
