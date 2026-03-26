package profile

var builtinProfiles = map[string]Profile{
	"browser-like": {
		Name: "browser-like",
		Headers: map[string]string{
			"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			"Accept-Language": "ja,en-US;q=0.7,en;q=0.3",
			"Accept-Encoding": "gzip, deflate, br",
			"Cache-Control":   "no-cache",
			"Connection":      "keep-alive",
		},
	},
	"bot-like": {
		Name: "bot-like",
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (compatible; TestBot/1.0)",
			"Accept":     "*/*",
		},
	},
	"crawler-like": {
		Name: "crawler-like",
		Headers: map[string]string{
			"User-Agent": "TestCrawler/1.0 (+https://example.internal/test)",
			"Accept":     "text/html,application/xhtml+xml",
		},
	},
}
