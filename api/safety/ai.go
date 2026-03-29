package safety

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AIClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewAIClient(apiKey string) *AIClient {
	return &AIClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
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

func (c *AIClient) Summarize(targetURL string) (string, error) {
	if c.apiKey == "" {
		return "AI要約は利用できません（APIキー未設定）", nil
	}

	prompt := `以下のURLの遷移先について、日本語で簡潔に要約してください（3〜5行程度）。

以下の観点で分析してください：
- そのサイトが何のサービス・ページか
- 主なコンテンツは何か
- フィッシング、マルウェア、詐欺、不審なリダイレクトなどの疑いがある場合は、最初に「⚠️ 警告:」と記載して危険性を説明してください

URL: ` + targetURL

	body := groqRequest{
		Model: "llama-3.1-8b-instant",
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
		return "", fmt.Errorf("AI API request failed: %w", err)
	}
	defer resp.Body.Close()

	var result groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result.Choices) == 0 {
		return "", fmt.Errorf("AI API response parse failed")
	}

	return result.Choices[0].Message.Content, nil
}
