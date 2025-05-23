package internal

import (
	"strings"

	"github.com/playwright-community/playwright-go"
)

func ClassifyAPI(req playwright.Request) bool {
	u := strings.ToLower(req.URL())
	rt := req.ResourceType()
	return rt == "xhr" || rt == "fetch" || strings.Contains(u, "/api/") || req.Method() == "POST"
}
