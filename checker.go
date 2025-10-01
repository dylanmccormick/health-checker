package main

import (
	"github.com/dylanmccormick/health-checker/assert"
)

func ValidateUrls(urls []string) []string {
	assert.Assert(len(urls) > 0, "Not enough URLs provided")

	var returnUrls []string
	for _, u := range urls {
		if validateUrl(u) {
			returnUrls = append(returnUrls, u)
		}
	}

	assert.Assert(len(returnUrls) > 0, "No valid URLs provided")
	return returnUrls
}

func validateUrl(url string) bool {

	return true
}
