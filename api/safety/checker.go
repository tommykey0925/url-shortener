package safety

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
	sbSafe, sbDetail, _ := c.sb.Check(targetURL)
	if !sbSafe {
		return CheckResult{Safe: false, Status: "unsafe", Detail: sbDetail}
	}
	return CheckResult{Safe: true, Status: "safe", Detail: sbDetail}
}

func (c *Checker) Summarize(targetURL string) (string, error) {
	return c.ai.Summarize(targetURL)
}
