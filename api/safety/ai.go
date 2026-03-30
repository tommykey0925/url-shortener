package safety

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type AIClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

func NewAIClient(apiKey string) *AIClient {
	return &AIClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    "https://api.groq.com/openai/v1/chat/completions",
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

var (
	titleRe = regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	metaRe  = regexp.MustCompile(`(?i)<meta[^>]+name=["']description["'][^>]+content=["']([^"']+)["']`)
	tagRe   = regexp.MustCompile(`<[^>]+>`)
)

func fetchPageContent(targetURL string) string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(targetURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 50000))
	if err != nil {
		return ""
	}
	html := string(body)

	var parts []string

	if m := titleRe.FindStringSubmatch(html); len(m) > 1 {
		parts = append(parts, "タイトル: "+strings.TrimSpace(m[1]))
	}

	if m := metaRe.FindStringSubmatch(html); len(m) > 1 {
		parts = append(parts, "説明: "+strings.TrimSpace(m[1]))
	}

	text := tagRe.ReplaceAllString(html, " ")
	text = strings.Join(strings.Fields(text), " ")
	if len(text) > 500 {
		text = text[:500]
	}
	if text != "" {
		parts = append(parts, "本文冒頭: "+text)
	}

	return strings.Join(parts, "\n")
}

func (c *AIClient) Summarize(targetURL string) (string, error) {
	return c.summarizeWithURL(c.baseURL, targetURL)
}

func (c *AIClient) summarizeWithURL(apiURL, targetURL string) (string, error) {
	if c.apiKey == "" {
		return "AI要約は利用できません（APIキー未設定）", nil
	}

	content := fetchPageContent(targetURL)

	var prompt string
	if content != "" {
		prompt = fmt.Sprintf(`以下のURLとそのページ内容について、日本語で簡潔に要約してください（3〜5行程度）。

以下の観点で分析してください：
- そのサイトが何のサービス・ページか
- 主なコンテンツは何か
- フィッシング、マルウェア、詐欺、不審なリダイレクトなどの疑いがある場合は、最初に「⚠️ 警告:」と記載して危険性を説明してください

URL: %s

ページ内容:
%s`, targetURL, content)
	} else {
		prompt = fmt.Sprintf(`以下のURLの遷移先について、日本語で簡潔に要約してください（3〜5行程度）。

以下の観点で分析してください：
- そのサイトが何のサービス・ページか
- 主なコンテンツは何か
- フィッシング、マルウェア、詐欺、不審なリダイレクトなどの疑いがある場合は、最初に「⚠️ 警告:」と記載して危険性を説明してください

URL: %s`, targetURL)
	}

	body := groqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []groqMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
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
