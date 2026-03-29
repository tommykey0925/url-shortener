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

type Checker struct {
	sb *SafeBrowsingClient
	ai *AIClient
}

func NewChecker(safeBrowsingKey, groqKey string) *Checker {
	return &Checker{
		sb: NewSafeBrowsingClient(safeBrowsingKey),
		ai: NewAIClient(groqKey),
	}
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
