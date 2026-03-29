package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/tommykey0925/url-shortener-api/model"
	"github.com/tommykey0925/url-shortener-api/safety"
	"github.com/tommykey0925/url-shortener-api/store"
)

type Handler struct {
	store   *store.Store
	checker *safety.Checker
}

func New(s *store.Store, c *safety.Checker) *Handler {
	return &Handler{store: s, checker: c}
}

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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	urls, err := h.store.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, urls)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if err := h.store.Delete(r.Context(), code); err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
