package safety

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestSBClient(srv *httptest.Server, apiKey string) *SafeBrowsingClient {
	return &SafeBrowsingClient{
		apiKey:     apiKey,
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}
}

func TestSafeBrowsingCheck_NoAPIKey(t *testing.T) {
	c := NewSafeBrowsingClient("")
	safe, detail, err := c.Check("https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !safe {
		t.Error("expected safe=true when no API key")
	}
	if detail != "skipped (no API key)" {
		t.Errorf("expected skipped detail, got %q", detail)
	}
}

func TestSafeBrowsingCheck_Safe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(sbResponse{})
	}))
	defer srv.Close()

	c := newTestSBClient(srv, "test-key")
	safe, detail, err := c.Check("https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !safe {
		t.Error("expected safe=true")
	}
	if detail != "safe" {
		t.Errorf("expected detail 'safe', got %q", detail)
	}
}

func TestSafeBrowsingCheck_Unsafe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(sbResponse{
			Matches: []sbMatch{{ThreatType: "MALWARE"}},
		})
	}))
	defer srv.Close()

	c := newTestSBClient(srv, "test-key")
	safe, detail, err := c.Check("https://malware.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if safe {
		t.Error("expected safe=false for malware URL")
	}
	if detail != "blocked: MALWARE" {
		t.Errorf("expected 'blocked: MALWARE', got %q", detail)
	}
}

func TestSafeBrowsingCheck_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer srv.Close()

	c := newTestSBClient(srv, "test-key")
	safe, detail, _ := c.Check("https://example.com")
	if !safe {
		t.Error("expected safe=true on network error (fail open)")
	}
	if detail != "check failed (network error)" {
		t.Errorf("expected network error detail, got %q", detail)
	}
}

func TestSafeBrowsingCheck_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := newTestSBClient(srv, "test-key")
	safe, detail, _ := c.Check("https://example.com")
	if !safe {
		t.Error("expected safe=true on parse error (fail open)")
	}
	if detail != "check failed (parse error)" {
		t.Errorf("expected parse error detail, got %q", detail)
	}
}

func TestSafeBrowsingCheck_RequestBody(t *testing.T) {
	var received sbRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		json.NewEncoder(w).Encode(sbResponse{})
	}))
	defer srv.Close()

	c := newTestSBClient(srv, "test-key")
	c.Check("https://example.com")

	if received.Client.ClientID != "url-shortener" {
		t.Errorf("expected clientId 'url-shortener', got %q", received.Client.ClientID)
	}
	if len(received.ThreatInfo.ThreatEntries) != 1 || received.ThreatInfo.ThreatEntries[0].URL != "https://example.com" {
		t.Error("expected target URL in threat entries")
	}
	if len(received.ThreatInfo.ThreatTypes) != 4 {
		t.Errorf("expected 4 threat types, got %d", len(received.ThreatInfo.ThreatTypes))
	}
}
