package safety

import (
	"net"
	"net/url"
)

type CheckResult struct {
	Safe   bool   `json:"safe"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// SafeBrowsingChecker is the interface for URL threat checking.
type SafeBrowsingChecker interface {
	Check(targetURL string) (bool, string, error)
}

// AISummarizer is the interface for AI-based URL summarization.
type AISummarizer interface {
	Summarize(targetURL string) (string, error)
}

type Checker struct {
	sb SafeBrowsingChecker
	ai AISummarizer
}

func NewChecker(safeBrowsingKey, groqKey string) *Checker {
	return &Checker{
		sb: NewSafeBrowsingClient(safeBrowsingKey),
		ai: NewAIClient(groqKey),
	}
}

func NewCheckerWithDeps(sb SafeBrowsingChecker, ai AISummarizer) *Checker {
	return &Checker{sb: sb, ai: ai}
}

func (c *Checker) Check(targetURL string) CheckResult {
	// 1. DNS check
	parsed, err := url.Parse(targetURL)
	if err != nil || parsed.Hostname() == "" {
		return CheckResult{Safe: false, Status: "unsafe", Detail: "invalid URL"}
	}
	if _, err := net.LookupHost(parsed.Hostname()); err != nil {
		return CheckResult{Safe: false, Status: "unsafe", Detail: "ドメインが存在しません"}
	}

	// 2. Google Safe Browsing check
	sbSafe, sbDetail, _ := c.sb.Check(targetURL)
	if !sbSafe {
		return CheckResult{Safe: false, Status: "unsafe", Detail: sbDetail}
	}

	return CheckResult{Safe: true, Status: "safe", Detail: sbDetail}
}

func (c *Checker) Summarize(targetURL string) (string, error) {
	return c.ai.Summarize(targetURL)
}
