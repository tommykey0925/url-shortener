package main

import (
	"log"
	"net/http"
	"os"

	"github.com/tommykey0925/url-shortener-api/handler"
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

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
