package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tommykey0925/url-shortener-api/model"
	"github.com/tommykey0925/url-shortener-api/safety"
	"github.com/tommykey0925/url-shortener-api/store"
)

// MonitorStore is the interface for store operations used by the monitor.
type MonitorStore interface {
	List(ctx context.Context) ([]model.URL, error)
	UpdateSafeStatus(ctx context.Context, code, status string) error
}

// RunMonitor checks all URLs and flags unsafe ones.
func RunMonitor(ctx context.Context, s MonitorStore, sb safety.SafeBrowsingChecker) error {
	urls, err := s.List(ctx)
	if err != nil {
		log.Printf("ERROR: failed to list URLs: %v", err)
		return err
	}

	log.Printf("Checking %d URLs...", len(urls))
	unsafeCount := 0

	for _, u := range urls {
		safe, detail, _ := sb.Check(u.Original)
		if !safe && u.SafeStatus != "unsafe" {
			if err := s.UpdateSafeStatus(ctx, u.Code, "unsafe"); err != nil {
				log.Printf("ERROR: failed to update %s: %v", u.Code, err)
				continue
			}
			unsafeCount++
			log.Printf("UNSAFE: %s (%s) - %s", u.Code, u.Original, detail)
		}
	}

	log.Printf("Done. %d URLs flagged as unsafe.", unsafeCount)
	return nil
}

func main() {
	s := store.New()
	sb := safety.NewSafeBrowsingClient(os.Getenv("GOOGLE_SAFE_BROWSING_API_KEY"))

	lambda.Start(func(ctx context.Context) error {
		return RunMonitor(ctx, s, sb)
	})
}
