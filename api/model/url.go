package model

import "time"

type URL struct {
	Code       string    `json:"code" dynamodbav:"code"`
	Original   string    `json:"original_url" dynamodbav:"original_url"`
	CreatedAt  time.Time `json:"created_at" dynamodbav:"created_at"`
	Clicks     int64     `json:"clicks" dynamodbav:"clicks"`
	SafeStatus string    `json:"safe_status" dynamodbav:"safe_status"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Code       string `json:"code"`
	ShortURL   string `json:"short_url"`
	SafeStatus string `json:"safe_status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type DailyClicks struct {
	Date   string `json:"date" dynamodbav:"date"`
	Clicks int64  `json:"clicks" dynamodbav:"clicks"`
}

type ClickStats struct {
	Code        string        `json:"code"`
	TotalClicks int64         `json:"total_clicks"`
	Daily       []DailyClicks `json:"daily"`
}
