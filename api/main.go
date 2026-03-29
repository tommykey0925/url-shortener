package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/tommykey0925/url-shortener-api/handler"
	"github.com/tommykey0925/url-shortener-api/middleware"
	"github.com/tommykey0925/url-shortener-api/safety"
	"github.com/tommykey0925/url-shortener-api/store"
)

func setupMux() http.Handler {
	s := store.New()
	checker := safety.NewChecker(os.Getenv("GOOGLE_SAFE_BROWSING_API_KEY"), os.Getenv("GROQ_API_KEY"))
	h := handler.New(s, checker)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/shorten", h.Shorten)
	mux.HandleFunc("GET /api/urls", h.List)
	mux.HandleFunc("GET /api/urls/{code}", h.Get)
	mux.HandleFunc("DELETE /api/urls/{code}", h.Delete)
	mux.HandleFunc("POST /api/urls/{code}/summarize", h.Summarize)
	mux.HandleFunc("GET /r/{code}", h.Redirect)
	mux.HandleFunc("GET /health", h.Health)

	rl := middleware.NewRateLimiter(10, time.Minute)
	return rl.Wrap(mux)
}

var adapter *httpadapter.HandlerAdapterV2

func lambdaHandler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return adapter.ProxyWithContext(ctx, req)
}

func main() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		adapter = httpadapter.NewV2(setupMux())
		lambda.Start(lambdaHandler)
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("Starting server on :%s", port)
		if err := http.ListenAndServe(":"+port, setupMux()); err != nil {
			log.Fatal(err)
		}
	}
}
