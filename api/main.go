package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tommykey0925/url-shortener-api/handler"
	"github.com/tommykey0925/url-shortener-api/middleware"
	"github.com/tommykey0925/url-shortener-api/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := store.New()
	h := handler.New(s)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/shorten", h.Shorten)
	mux.HandleFunc("GET /api/urls", h.List)
	mux.HandleFunc("GET /api/urls/{code}", h.Get)
	mux.HandleFunc("DELETE /api/urls/{code}", h.Delete)
	mux.HandleFunc("GET /r/{code}", h.Redirect)
	mux.HandleFunc("GET /health", h.Health)

	// Rate limit: 10 requests per IP per minute
	rl := middleware.NewRateLimiter(10, time.Minute)

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, rl.Wrap(mux)); err != nil {
		log.Fatal(err)
	}
}
