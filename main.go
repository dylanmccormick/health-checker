package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
)

type HealthChecker struct {
	// this will hold metrics and stuff later. Good place to store things for uptime and whatever else we may want
	httpClient *http.Client
	config Config
}

func main() {
	w := os.Stderr
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-sigChan
		slog.Info("received shutdown signal")
		cancel()
	}()

	conf, err := GetConfig()
	if err != nil {
		slog.Error("Error getting config", "error", err)
		os.Exit(1)
	}
	validUrls := ValidateUrls(conf.Urls)
	conf.Urls = validUrls

	h := NewHealthChecker(conf)
	h.Run(ctx, conf, sigChan)
	// Check health
	os.Exit(0)
}

func NewHealthChecker(c Config) *HealthChecker{
	return &HealthChecker{
		httpClient: &http.Client{
			Timeout: time.Duration(c.TimeoutSeconds) * time.Second,
		},
		config: c,
	}
}
func (h *HealthChecker) Run(ctx context.Context, c Config, sigChan chan os.Signal) {
	var wg sync.WaitGroup
	for _, url := range c.Urls {
		wg.Add(1)
		go func(url string){
			defer wg.Done()
			h.monitorUrl(ctx, url)
			healthCheck(ctx, url, c)
		}(url)
	}
	wg.Wait()
	slog.Info("All healthchecks stopped")
}

func (h *HealthChecker) monitorUrl(ctx context.Context, url string){
}

func healthCheck(ctx context.Context, url string, c Config) {
	// ticker := time.NewTicker(time.Duration(c.IntervalSeconds) * time.Second)
	ticker := time.NewTicker(time.Duration(1) * time.Second)
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping health check for URL", "url", url, "reason", ctx.Err())
			return
		case <-ticker.C:
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Println(err)
				return
			}
			start := time.Now()
		}
	}
}

func checkUrl(ctx context.Context, url string, client *http.Client) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		slog.Error("failed to create http request", "url", url, "error", err)
		return
		resp, err := c.httpClient.Do(req)
		if err != nil {
			slog.Error("", "url", url, "status", "NONE", "healthy", false, "response_time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
			continue
		}
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusInternalServerError:
			slog.Error("", "url", url, "status", resp.StatusCode, "healthy", false, "response_time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))

		case http.StatusOK:
			slog.Info("", "url", url, "status", resp.StatusCode, "healthy", true, "response_time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
		}
	}
}
