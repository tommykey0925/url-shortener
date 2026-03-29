package safety

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type AIClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewAIClient(apiKey string) *AIClient {
	return &AIClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type groqRequest struct {
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponse struct {
	Choices []struct {
		Message groqMessage `json:"message"`
	} `json:"choices"`
}

func (c *AIClient) Predict(targetURL string) (bool, string, error) {
	if c.apiKey == "" {
		return true, "skipped (no API key)", nil
	}

	prompt := `Analyze the following URL for safety. Consider:
- Is the domain well-known and reputable?
- Does the URL path contain suspicious patterns (phishing, malware, scam)?
- Are there signs of URL obfuscation or deception?

Reply with ONLY one word: "SAFE" or "UNSAFE", followed by a brief reason.

URL: ` + targetURL

	body := groqRequest{
		Model: "gemma2-9b-it",
		Messages: []groqMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return true, "AI check failed (network error)", nil
	}
	defer resp.Body.Close()

	var result groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result.Choices) == 0 {
		return true, "AI check failed (parse error)", nil
	}

	answer := result.Choices[0].Message.Content
	isUnsafe := strings.HasPrefix(strings.ToUpper(strings.TrimSpace(answer)), "UNSAFE")

	if isUnsafe {
		return false, "AI prediction: " + answer, nil
	}
	return true, "AI prediction: " + answer, nil
}
