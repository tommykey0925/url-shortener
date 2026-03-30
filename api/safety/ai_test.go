package safety

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchPageContent_WithTitleAndMeta(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head>
			<title>Test Page</title>
			<meta name="description" content="A test description">
		</head><body>Hello World</body></html>`))
	}))
	defer srv.Close()

	content := fetchPageContent(srv.URL)
	if !strings.Contains(content, "タイトル: Test Page") {
		t.Errorf("expected title in content, got %q", content)
	}
	if !strings.Contains(content, "説明: A test description") {
		t.Errorf("expected description in content, got %q", content)
	}
	if !strings.Contains(content, "本文冒頭:") {
		t.Errorf("expected body text in content, got %q", content)
	}
}

func TestFetchPageContent_EmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(""))
	}))
	defer srv.Close()

	content := fetchPageContent(srv.URL)
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
}

func TestFetchPageContent_Unreachable(t *testing.T) {
	content := fetchPageContent("http://127.0.0.1:1")
	if content != "" {
		t.Errorf("expected empty content for unreachable URL, got %q", content)
	}
}

func TestFetchPageContent_LongBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>" + strings.Repeat("a", 1000) + "</body></html>"))
	}))
	defer srv.Close()

	content := fetchPageContent(srv.URL)
	// Body text should be truncated to 500 chars
	if strings.Contains(content, "本文冒頭:") {
		bodyIdx := strings.Index(content, "本文冒頭: ")
		bodyText := content[bodyIdx+len("本文冒頭: "):]
		if len(bodyText) > 500 {
			t.Errorf("body text should be truncated to 500 chars, got %d", len(bodyText))
		}
	}
}

func TestAIClient_Summarize_NoAPIKey(t *testing.T) {
	c := NewAIClient("")
	summary, err := c.Summarize("https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != "AI要約は利用できません（APIキー未設定）" {
		t.Errorf("expected no-key message, got %q", summary)
	}
}

func TestAIClient_Summarize_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %q", r.Header.Get("Authorization"))
		}

		var req groqRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "llama-3.3-70b-versatile" {
			t.Errorf("expected llama model, got %q", req.Model)
		}

		json.NewEncoder(w).Encode(groqResponse{
			Choices: []struct {
				Message groqMessage `json:"message"`
			}{
				{Message: groqMessage{Role: "assistant", Content: "これはテストページです。"}},
			},
		})
	}))
	defer srv.Close()

	c := &AIClient{
		apiKey:     "test-key",
		httpClient: srv.Client(),
	}

	// We need to override the URL. Use a custom approach:
	// Create a page content server and an AI server
	pageSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><title>Test</title><body>Hello</body></html>"))
	}))
	defer pageSrv.Close()

	// For this test, we test the AI client directly by replacing the endpoint
	// Since the Groq URL is hardcoded, we need a different approach
	// Let's test the interface-level behavior instead
	summary, err := c.summarizeWithURL(srv.URL, pageSrv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != "これはテストページです。" {
		t.Errorf("expected test summary, got %q", summary)
	}
}

func TestAIClient_Summarize_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	c := &AIClient{
		apiKey:     "test-key",
		httpClient: srv.Client(),
	}

	_, err := c.summarizeWithURL(srv.URL, "http://127.0.0.1:1")
	if err == nil {
		t.Error("expected error on API failure")
	}
}
