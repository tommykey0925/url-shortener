//go:build integration

package safety

import (
	"os"
	"testing"
)

func TestSafeBrowsingIntegration(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_SAFE_BROWSING_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_SAFE_BROWSING_API_KEY not set")
	}

	client := NewSafeBrowsingClient(apiKey)

	tests := []struct {
		name       string
		url        string
		wantSafe   bool
		wantDetail string
	}{
		{
			name:       "malware test URL",
			url:        "http://testsafebrowsing.appspot.com/s/malware.html",
			wantSafe:   false,
			wantDetail: "blocked: MALWARE",
		},
		{
			name:       "phishing test URL",
			url:        "http://testsafebrowsing.appspot.com/s/phishing.html",
			wantSafe:   false,
			wantDetail: "blocked: SOCIAL_ENGINEERING",
		},
		{
			name:       "unwanted software test URL",
			url:        "http://testsafebrowsing.appspot.com/s/unwanted.html",
			wantSafe:   false,
			wantDetail: "blocked: UNWANTED_SOFTWARE",
		},
		{
			name:       "safe URL",
			url:        "https://example.com",
			wantSafe:   true,
			wantDetail: "safe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safe, detail, err := client.Check(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if safe != tt.wantSafe {
				t.Errorf("safe = %v, want %v (detail: %s)", safe, tt.wantSafe, detail)
			}
			if detail != tt.wantDetail {
				t.Errorf("detail = %q, want %q", detail, tt.wantDetail)
			}
		})
	}
}
