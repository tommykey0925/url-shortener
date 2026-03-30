package safety

import (
	"testing"
)

// mockSB implements SafeBrowsingChecker for testing.
type mockSB struct {
	safe   bool
	detail string
	err    error
}

func (m *mockSB) Check(targetURL string) (bool, string, error) {
	return m.safe, m.detail, m.err
}

// mockAI implements AISummarizer for testing.
type mockAI struct {
	summary string
	err     error
}

func (m *mockAI) Summarize(targetURL string) (string, error) {
	return m.summary, m.err
}

func TestChecker_Check_InvalidURL(t *testing.T) {
	c := NewCheckerWithDeps(&mockSB{safe: true, detail: "safe"}, &mockAI{})
	result := c.Check("://invalid")
	if result.Safe {
		t.Error("expected unsafe for invalid URL")
	}
	if result.Detail != "invalid URL" {
		t.Errorf("expected 'invalid URL', got %q", result.Detail)
	}
}

func TestChecker_Check_EmptyHostname(t *testing.T) {
	c := NewCheckerWithDeps(&mockSB{safe: true, detail: "safe"}, &mockAI{})
	result := c.Check("http://")
	if result.Safe {
		t.Error("expected unsafe for empty hostname")
	}
	if result.Detail != "invalid URL" {
		t.Errorf("expected 'invalid URL', got %q", result.Detail)
	}
}

func TestChecker_Check_DNSFailure(t *testing.T) {
	c := NewCheckerWithDeps(&mockSB{safe: true, detail: "safe"}, &mockAI{})
	result := c.Check("https://this-domain-does-not-exist-xyzzy.example")
	if result.Safe {
		t.Error("expected unsafe for non-existent domain")
	}
	if result.Status != "unsafe" {
		t.Errorf("expected status 'unsafe', got %q", result.Status)
	}
	if result.Detail != "ドメインが存在しません" {
		t.Errorf("expected DNS error detail, got %q", result.Detail)
	}
}

func TestChecker_Check_SafeBrowsingUnsafe(t *testing.T) {
	c := NewCheckerWithDeps(
		&mockSB{safe: false, detail: "blocked: MALWARE"},
		&mockAI{},
	)
	// Use a real domain that DNS resolves
	result := c.Check("https://example.com")
	if result.Safe {
		t.Error("expected unsafe when SafeBrowsing flags URL")
	}
	if result.Detail != "blocked: MALWARE" {
		t.Errorf("expected 'blocked: MALWARE', got %q", result.Detail)
	}
}

func TestChecker_Check_Safe(t *testing.T) {
	c := NewCheckerWithDeps(
		&mockSB{safe: true, detail: "safe"},
		&mockAI{},
	)
	result := c.Check("https://example.com")
	if !result.Safe {
		t.Error("expected safe=true")
	}
	if result.Status != "safe" {
		t.Errorf("expected status 'safe', got %q", result.Status)
	}
}

func TestChecker_Summarize(t *testing.T) {
	c := NewCheckerWithDeps(
		&mockSB{safe: true, detail: "safe"},
		&mockAI{summary: "テストサマリー", err: nil},
	)
	summary, err := c.Summarize("https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != "テストサマリー" {
		t.Errorf("expected 'テストサマリー', got %q", summary)
	}
}
