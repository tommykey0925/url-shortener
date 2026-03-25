package model

import "time"

type URL struct {
	Code      string    `json:"code" dynamodbav:"code"`
	Original  string    `json:"original_url" dynamodbav:"original_url"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
	Clicks    int64     `json:"clicks" dynamodbav:"clicks"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
