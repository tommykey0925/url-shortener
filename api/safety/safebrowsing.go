package safety

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SafeBrowsingClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewSafeBrowsingClient(apiKey string) *SafeBrowsingClient {
	return &SafeBrowsingClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type sbRequest struct {
	Client     sbClient     `json:"client"`
	ThreatInfo sbThreatInfo `json:"threatInfo"`
}

type sbClient struct {
	ClientID      string `json:"clientId"`
	ClientVersion string `json:"clientVersion"`
}

type sbThreatInfo struct {
	ThreatTypes      []string       `json:"threatTypes"`
	PlatformTypes    []string       `json:"platformTypes"`
	ThreatEntryTypes []string       `json:"threatEntryTypes"`
	ThreatEntries    []sbThreatEntry `json:"threatEntries"`
}

type sbThreatEntry struct {
	URL string `json:"url"`
}

type sbResponse struct {
	Matches []sbMatch `json:"matches"`
}

type sbMatch struct {
	ThreatType string `json:"threatType"`
}

func (c *SafeBrowsingClient) Check(targetURL string) (bool, string, error) {
	if c.apiKey == "" {
		return true, "skipped (no API key)", nil
	}

	body := sbRequest{
		Client: sbClient{ClientID: "url-shortener", ClientVersion: "1.0"},
		ThreatInfo: sbThreatInfo{
			ThreatTypes:      []string{"MALWARE", "SOCIAL_ENGINEERING", "UNWANTED_SOFTWARE", "POTENTIALLY_HARMFUL_APPLICATION"},
			PlatformTypes:    []string{"ANY_PLATFORM"},
			ThreatEntryTypes: []string{"URL"},
			ThreatEntries:    []sbThreatEntry{{URL: targetURL}},
		},
	}

	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("https://safebrowsing.googleapis.com/v4/threatMatches:find?key=%s", c.apiKey)

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return true, "check failed (network error)", nil
	}
	defer resp.Body.Close()

	var result sbResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return true, "check failed (parse error)", nil
	}

	if len(result.Matches) > 0 {
		return false, fmt.Sprintf("blocked: %s", result.Matches[0].ThreatType), nil
	}

	return true, "safe", nil
}
