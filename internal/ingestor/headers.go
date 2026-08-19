package ingestor

import "math/rand"

// UserAgentPool holds realistic desktop browser User-Agents.
var UserAgentPool = []string{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
}

// GetRandomUserAgent selects a User-Agent string randomly from the pool.
func GetRandomUserAgent() string {
	return UserAgentPool[rand.Intn(len(UserAgentPool))]
}

// BuildHumanizedHeaders returns a set of headers standard for modern desktop browsers.
func BuildHumanizedHeaders(ua string) map[string]string {
	return map[string]string{
		"User-Agent":                ua,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.9",
		"Accept-Encoding":           "gzip, deflate, br",
		"Cache-Control":             "max-age=0",
		"Sec-Ch-Ua":                 `"Not/A)Brand";v="8", "Chromium";v="127", "Google Chrome";v="127"`,
		"Sec-Ch-Ua-Mobile":          "?0",
		"Sec-Ch-Ua-Platform":        `"macOS"`,
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"Upgrade-Insecure-Requests": "1",
	}
}
