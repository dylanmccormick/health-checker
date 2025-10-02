package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/dylanmccormick/health-checker/assert"
)

type Config struct {
	IntervalSeconds int      `json:"check_interval_seconds"`
	TimeoutSeconds  int      `json:"timeout_seconds"`
	Urls            []string `json:"Urls"`
	httpClient      *http.Client
}

func GetConfig() (Config, error) {
	var err error

	dat, err := os.ReadFile("./config.json")
	if err != nil {
		fmt.Println(err)
		return Config{}, err
	}

	conf := Config{}

	err = json.Unmarshal(dat, &conf)
	if err != nil {
		return Config{}, err
	}

	return validateConfig(conf)
}

func validateConfig(c Config) (Config, error) {
	// At some point we should have default values which are set if something doesn't exist
	assert.Assert(c.IntervalSeconds > 0, "Interval Seconds Must Be Positive")
	assert.Assert(c.TimeoutSeconds > 0, "Timeout Seconds Must Be Positive")
	c.Urls = ValidateUrls(c.Urls)
	return c, nil
}

func ValidateUrls(urls []string) []string {
	assert.Assert(len(urls) > 0, "No URLs provided")

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
