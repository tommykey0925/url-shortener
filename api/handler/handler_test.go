package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tommykey-apps/url-shortener-api/model"
	"github.com/tommykey-apps/url-shortener-api/safety"
	"github.com/tommykey-apps/url-shortener-api/store"
)

// --- Mock Store ---

type mockStore struct {
	urls   map[string]*model.URL
	putErr error
}

func newMockStore() *mockStore {
	return &mockStore{urls: make(map[string]*model.URL)}
}

func (m *mockStore) Put(ctx context.Context, code, originalURL, safeStatus string) (*model.URL, error) {
	if m.putErr != nil {
		return nil, m.putErr
	}
	u := &model.URL{Code: code, Original: originalURL, SafeStatus: safeStatus}
	m.urls[code] = u
	return u, nil
}

func (m *mockStore) Get(ctx context.Context, code string) (*model.URL, error) {
	u, ok := m.urls[code]
	if !ok {
		return nil, store.ErrNotFound
	}
	return u, nil
}

func (m *mockStore) List(ctx context.Context) ([]model.URL, error) {
	var urls []model.URL
	for _, u := range m.urls {
		urls = append(urls, *u)
	}
	return urls, nil
}

func (m *mockStore) Delete(ctx context.Context, code string) error {
	delete(m.urls, code)
	return nil
}

func (m *mockStore) IncrementClicks(ctx context.Context, code string) error {
	if u, ok := m.urls[code]; ok {
		u.Clicks++
	}
	return nil
}

func (m *mockStore) GetClickStats(ctx context.Context, code string, days int) ([]model.DailyClicks, error) {
	if _, ok := m.urls[code]; !ok {
		return nil, nil
	}
	return []model.DailyClicks{
		{Date: "2026-04-17", Clicks: 5},
		{Date: "2026-04-16", Clicks: 3},
	}, nil
}

// --- Mock Checker ---

type mockChecker struct {
	result  safety.CheckResult
	summary string
	sumErr  error
}

func (m *mockChecker) Check(targetURL string) safety.CheckResult {
	return m.result
}

func (m *mockChecker) Summarize(targetURL string) (string, error) {
	return m.summary, m.sumErr
}

// --- Tests ---

func TestShorten_Success(t *testing.T) {
	s := newMockStore()
	c := &mockChecker{result: safety.CheckResult{Safe: true, Status: "safe", Detail: "safe"}}
	h := NewWithDeps(s, c)

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.Shorten(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp model.ShortenResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code == "" {
		t.Error("expected non-empty code")
	}
	if resp.SafeStatus != "safe" {
		t.Errorf("expected safe status, got %q", resp.SafeStatus)
	}
}

func TestShorten_InvalidBody(t *testing.T) {
	h := NewWithDeps(newMockStore(), &mockChecker{})
	req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	h.Shorten(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestShorten_InvalidURL(t *testing.T) {
	h := NewWithDeps(newMockStore(), &mockChecker{})
	body := `{"url":"not-a-url"}`
	req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.Shorten(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestShorten_UnsafeURL(t *testing.T) {
	c := &mockChecker{result: safety.CheckResult{Safe: false, Status: "unsafe", Detail: "blocked: MALWARE"}}
	h := NewWithDeps(newMockStore(), c)

	body := `{"url":"https://malware.example.com"}`
	req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.Shorten(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	var resp model.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !strings.Contains(resp.Error, "MALWARE") {
		t.Errorf("expected MALWARE in error, got %q", resp.Error)
	}
}

func TestShorten_StorePutError(t *testing.T) {
	s := newMockStore()
	s.putErr = fmt.Errorf("dynamo error")
	c := &mockChecker{result: safety.CheckResult{Safe: true, Status: "safe", Detail: "safe"}}
	h := NewWithDeps(s, c)

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.Shorten(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGet_Found(t *testing.T) {
	s := newMockStore()
	s.urls["abc1234"] = &model.URL{Code: "abc1234", Original: "https://example.com", SafeStatus: "safe"}
	h := NewWithDeps(s, &mockChecker{})

	req := httptest.NewRequest("GET", "/api/urls/abc1234", nil)
	req.SetPathValue("code", "abc1234")
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGet_NotFound(t *testing.T) {
	h := NewWithDeps(newMockStore(), &mockChecker{})

	req := httptest.NewRequest("GET", "/api/urls/unknown", nil)
	req.SetPathValue("code", "unknown")
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestList(t *testing.T) {
	s := newMockStore()
	s.urls["a"] = &model.URL{Code: "a", Original: "https://a.com"}
	s.urls["b"] = &model.URL{Code: "b", Original: "https://b.com"}
	h := NewWithDeps(s, &mockChecker{})

	req := httptest.NewRequest("GET", "/api/urls", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var urls []model.URL
	json.NewDecoder(w.Body).Decode(&urls)
	if len(urls) != 2 {
		t.Errorf("expected 2 urls, got %d", len(urls))
	}
}

func TestDelete(t *testing.T) {
	s := newMockStore()
	s.urls["abc"] = &model.URL{Code: "abc"}
	h := NewWithDeps(s, &mockChecker{})

	req := httptest.NewRequest("DELETE", "/api/urls/abc", nil)
	req.SetPathValue("code", "abc")
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
	if _, ok := s.urls["abc"]; ok {
		t.Error("expected URL to be deleted")
	}
}

func TestRedirect_Safe(t *testing.T) {
	s := newMockStore()
	s.urls["abc"] = &model.URL{Code: "abc", Original: "https://example.com", SafeStatus: "safe"}
	h := NewWithDeps(s, &mockChecker{})

	req := httptest.NewRequest("GET", "/r/abc", nil)
	req.SetPathValue("code", "abc")
	w := httptest.NewRecorder()

	h.Redirect(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "https://example.com" {
		t.Errorf("expected redirect to https://example.com, got %q", loc)
	}
}

func TestRedirect_Unsafe(t *testing.T) {
	s := newMockStore()
	s.urls["bad"] = &model.URL{Code: "bad", Original: "https://malware.com", SafeStatus: "unsafe"}
	h := NewWithDeps(s, &mockChecker{})

	req := httptest.NewRequest("GET", "/r/bad", nil)
	req.SetPathValue("code", "bad")
	w := httptest.NewRecorder()

	h.Redirect(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (warning page), got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Warning") {
		t.Error("expected warning page content")
	}
	if !strings.Contains(body, "https://malware.com") {
		t.Error("expected original URL in warning page")
	}
}

func TestRedirect_NotFound(t *testing.T) {
	h := NewWithDeps(newMockStore(), &mockChecker{})

	req := httptest.NewRequest("GET", "/r/nope", nil)
	req.SetPathValue("code", "nope")
	w := httptest.NewRecorder()

	h.Redirect(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestSummarize_Success(t *testing.T) {
	s := newMockStore()
	s.urls["abc"] = &model.URL{Code: "abc", Original: "https://example.com"}
	c := &mockChecker{summary: "テストサマリー"}
	h := NewWithDeps(s, c)

	req := httptest.NewRequest("POST", "/api/urls/abc/summarize", nil)
	req.SetPathValue("code", "abc")
	w := httptest.NewRecorder()

	h.Summarize(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["summary"] != "テストサマリー" {
		t.Errorf("expected test summary, got %q", resp["summary"])
	}
}

func TestSummarize_NotFound(t *testing.T) {
	h := NewWithDeps(newMockStore(), &mockChecker{})

	req := httptest.NewRequest("POST", "/api/urls/nope/summarize", nil)
	req.SetPathValue("code", "nope")
	w := httptest.NewRecorder()

	h.Summarize(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestSummarize_AIError(t *testing.T) {
	s := newMockStore()
	s.urls["abc"] = &model.URL{Code: "abc", Original: "https://example.com"}
	c := &mockChecker{sumErr: fmt.Errorf("AI error")}
	h := NewWithDeps(s, c)

	req := httptest.NewRequest("POST", "/api/urls/abc/summarize", nil)
	req.SetPathValue("code", "abc")
	w := httptest.NewRecorder()

	h.Summarize(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestStats_Success(t *testing.T) {
	s := newMockStore()
	s.urls["abc"] = &model.URL{Code: "abc", Original: "https://example.com", Clicks: 8}
	h := NewWithDeps(s, &mockChecker{})

	req := httptest.NewRequest("GET", "/api/urls/abc/stats", nil)
	req.SetPathValue("code", "abc")
	w := httptest.NewRecorder()

	h.Stats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp model.ClickStats
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.TotalClicks != 8 {
		t.Errorf("expected 8 total clicks, got %d", resp.TotalClicks)
	}
	if len(resp.Daily) != 2 {
		t.Errorf("expected 2 daily entries, got %d", len(resp.Daily))
	}
}

func TestStats_NotFound(t *testing.T) {
	h := NewWithDeps(newMockStore(), &mockChecker{})

	req := httptest.NewRequest("GET", "/api/urls/nope/stats", nil)
	req.SetPathValue("code", "nope")
	w := httptest.NewRecorder()

	h.Stats(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHealth(t *testing.T) {
	h := NewWithDeps(newMockStore(), &mockChecker{})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %q", resp["status"])
	}
}
