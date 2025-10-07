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
	config     Config
	metrics    map[string]*URLMetrics
}

type URLMetrics struct {
	TotalChecks       int
	SuccessfulChecks  int
	TotalResponseTime time.Duration
	Mutex             *sync.RWMutex
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
	h.Run(ctx)
	// Check health
	os.Exit(0)
}

func NewHealthChecker(c Config) *HealthChecker {
	return &HealthChecker{
		httpClient: &http.Client{
			Timeout: time.Duration(c.TimeoutSeconds) * time.Second,
		},
		config:  c,
		metrics: make(map[string]*URLMetrics),
	}
}

func (h *HealthChecker) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		h.logMetrics(ctx)
	})
	for _, url := range h.config.Urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			h.createUrlEntry(ctx, url)
			h.monitorUrl(ctx, url)
		}(url)
	}
	wg.Wait()
	slog.Info("All healthchecks stopped")
}

func (h *HealthChecker) createUrlEntry(ctx context.Context, url string) {
	h.metrics[url] = &URLMetrics{
		TotalChecks:       0,
		SuccessfulChecks:  0,
		TotalResponseTime: time.Duration(0),
		Mutex:             new(sync.RWMutex),
	}
}

func (h *HealthChecker) logMetrics(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(h.config.IntervalSeconds) * time.Second)
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping health check metrics", "reason", ctx.Err())
			return
		case <-ticker.C:
			for _, u := range h.config.Urls {
				m := h.metrics[u]
				m.Mutex.RLock()
				if m.TotalChecks == 0 {
					m.Mutex.RUnlock()
					continue
				}
				slog.Info(
					"metrics",
					"URL", u,
					"TotalChecks", m.TotalChecks,
					"SuccessfulChecks", m.SuccessfulChecks,
					"AvgResponseTime", fmt.Sprintf("%dms", int(m.TotalResponseTime.Milliseconds())/m.TotalChecks),
				)
				m.Mutex.RUnlock()
			}
		}
	}
}

func (h *HealthChecker) monitorUrl(ctx context.Context, url string) {
	ticker := time.NewTicker(time.Duration(h.config.IntervalSeconds) * time.Second)
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping health check for URL", "url", url, "reason", ctx.Err())
			return
		case <-ticker.C:
			h.checkUrl(ctx, url)
		}
	}
}

func (h *HealthChecker) checkUrl(ctx context.Context, url string) {
	start := time.Now()
	m := h.metrics[url]
	m.Mutex.Lock()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		slog.Error("failed to create http request", "url", url, "error", err)
		return
	}
	m.TotalChecks += 1
	resp, err := h.httpClient.Do(req)
	if err != nil {
		responseTime := time.Since(start)
		slog.Error("",
			"url", url,
			"status", "NONE",
			"healthy", false,
			"response_time", fmt.Sprintf("%dms", responseTime.Milliseconds()))
		m.Mutex.Unlock()
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		m.SuccessfulChecks += 1
		responseTime := time.Since(start)
		m.TotalResponseTime += responseTime
		slog.Info("",
			"url", url,
			"status", resp.StatusCode,
			"healthy", true,
			"response_time", fmt.Sprintf("%dms", responseTime.Milliseconds()))
	} else {
		responseTime := time.Since(start)
		m.TotalResponseTime += responseTime
		slog.Error("",
			"url", url,
			"status", resp.StatusCode,
			"healthy", false,
			"response_time", fmt.Sprintf("%dms", responseTime.Milliseconds()))
	}
	m.Mutex.Unlock()
}
