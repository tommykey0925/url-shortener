package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/tommykey-apps/url-shortener-api/model"
	"github.com/tommykey-apps/url-shortener-api/safety"
	"github.com/tommykey-apps/url-shortener-api/store"
)

// URLStore is the interface for URL persistence operations.
type URLStore interface {
	Put(ctx context.Context, code, originalURL, safeStatus string) (*model.URL, error)
	Get(ctx context.Context, code string) (*model.URL, error)
	List(ctx context.Context) ([]model.URL, error)
	Delete(ctx context.Context, code string) error
	IncrementClicks(ctx context.Context, code string) error
	GetClickStats(ctx context.Context, code string, days int) ([]model.DailyClicks, error)
}

type DailyClicks = model.DailyClicks

// SafetyChecker is the interface for URL safety checking.
type SafetyChecker interface {
	Check(targetURL string) safety.CheckResult
	Summarize(targetURL string) (string, error)
}

type Handler struct {
	store   URLStore
	checker SafetyChecker
}

func New(s *store.Store, c *safety.Checker) *Handler {
	return &Handler{store: s, checker: c}
}

func NewWithDeps(s URLStore, c SafetyChecker) *Handler {
	return &Handler{store: s, checker: c}
}

// Shorten godoc
// @Summary URLを短縮
// @Description URLを受け取り短縮コードを生成。DNS解決チェック + Google Safe Browsing APIで安全性を検証し、危険なURLは拒否する。
// @Tags URLs
// @Accept json
// @Produce json
// @Param request body model.ShortenRequest true "短縮したいURL"
// @Success 201 {object} model.ShortenResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/shorten [post]
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req model.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid url"})
		return
	}

	// Safety check
	result := h.checker.Check(req.URL)
	if !result.Safe {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "unsafe URL: " + result.Detail})
		return
	}

	code := generateCode()
	u, err := h.store.Put(r.Context(), code, req.URL, result.Status)
	if err != nil {
		log.Printf("ERROR: store.Put failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "failed to create short url"})
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://%s", r.Host)
	}

	writeJSON(w, http.StatusCreated, model.ShortenResponse{
		Code:       u.Code,
		ShortURL:   fmt.Sprintf("%s/r/%s", baseURL, u.Code),
		SafeStatus: u.SafeStatus,
	})
}

// Get godoc
// @Summary URL詳細取得
// @Tags URLs
// @Produce json
// @Param code path string true "短縮コード"
// @Success 200 {object} model.URL
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/urls/{code} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	u, err := h.store.Get(r.Context(), code)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// List godoc
// @Summary URL一覧取得
// @Tags URLs
// @Produce json
// @Success 200 {array} model.URL
// @Failure 500 {object} model.ErrorResponse
// @Router /api/urls [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	urls, err := h.store.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, urls)
}

// Delete godoc
// @Summary URL削除
// @Tags URLs
// @Param code path string true "短縮コード"
// @Success 204 "削除成功"
// @Failure 500 {object} model.ErrorResponse
// @Router /api/urls/{code} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if err := h.store.Delete(r.Context(), code); err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Redirect godoc
// @Summary リダイレクト
// @Description 短縮コードに対応する元URLへ301リダイレクト。クリック数をカウント。unsafe判定のURLは警告ページを表示。
// @Tags Redirect
// @Param code path string true "短縮コード"
// @Success 301 "元URLへリダイレクト"
// @Success 200 {string} string "unsafe URLの場合、警告ページを表示"
// @Failure 404 "該当コードなし"
// @Router /r/{code} [get]
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	u, err := h.store.Get(r.Context(), code)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if u.SafeStatus == "unsafe" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<html><body style="font-family:sans-serif;max-width:600px;margin:40px auto;padding:20px">
			<h1>⚠️ Warning</h1>
			<p>This URL has been flagged as potentially unsafe.</p>
			<p>Destination: <code>%s</code></p>
			<p><a href="%s">Continue anyway</a></p>
		</body></html>`, u.Original, u.Original)
		return
	}

	_ = h.store.IncrementClicks(r.Context(), code)
	http.Redirect(w, r, u.Original, http.StatusMovedPermanently)
}

// Summarize godoc
// @Summary AIによるURL遷移先の要約
// @Description 短縮URLの遷移先ページをAI (Groq / Llama 3.3 70B) で要約する。
// @Tags AI
// @Produce json
// @Param code path string true "短縮コード"
// @Success 200 {object} map[string]string
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/urls/{code}/summarize [post]
func (h *Handler) Summarize(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	u, err := h.store.Get(r.Context(), code)
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "not found"})
		return
	}

	summary, err := h.checker.Summarize(u.Original)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "AI summarization failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"summary": summary})
}

// Stats godoc
// @Summary クリック統計取得
// @Description 短縮URLの日別クリック統計を取得する。デフォルトは過去30日分。
// @Tags URLs
// @Produce json
// @Param code path string true "短縮コード"
// @Param days query int false "取得日数（デフォルト30）"
// @Success 200 {object} model.ClickStats
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/urls/{code}/stats [get]
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	u, err := h.store.Get(r.Context(), code)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	daily, err := h.store.GetClickStats(r.Context(), code, days)
	if err != nil {
		log.Printf("ERROR: GetClickStats failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get stats"})
		return
	}

	writeJSON(w, http.StatusOK, model.ClickStats{
		Code:        code,
		TotalClicks: u.Clicks,
		Daily:       daily,
	})
}

// Health godoc
// @Summary ヘルスチェック
// @Tags System
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "1.0.0"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func generateCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)[:7]
}
