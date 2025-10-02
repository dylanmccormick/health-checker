package main

import (
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

func main() {
	w := os.Stderr
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level: slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	conf, err := GetConfig()
	if err != nil {
		slog.Error("Error getting config", "error", err)
		os.Exit(1)
	}
	validUrls := ValidateUrls(conf.Urls)
	conf.Urls = validUrls

	httpClient := &http.Client{
		Timeout: time.Duration(conf.TimeoutSeconds) * time.Second,
	}
	conf.httpClient = httpClient
	go Run(conf, sigChan)
	// Check health
	select {
	case s := <-sigChan:
		fmt.Printf("Received stop signal %v. Shutting down...\n", s)
		// HERE is where I will put the cleanup process. Shutting down other channels or whatever. Saving state.

		time.Sleep(2 * time.Millisecond)
		fmt.Printf("Cleanup complete... exiting")
		os.Exit(0)
	}
}

func Run(c Config, sigChan chan os.Signal) {
	var wg sync.WaitGroup
	for i, url := range c.Urls {
		wg.Add(i)
		go HealthCheck(url, c, sigChan, &wg)
	}
	wg.Wait()
}

func HealthCheck(url string, c Config, sigChan chan os.Signal, wg *sync.WaitGroup) {
	// ticker := time.NewTicker(time.Duration(c.IntervalSeconds) * time.Second)
	ticker := time.NewTicker(time.Duration(1) * time.Second)
	for {
		select {
		case s := <-sigChan:
			slog.Info("stopping health check for URL", "url", url, "sig", s)
			wg.Done()
		case <-ticker.C:
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Println(err)
				return
			}
			resp, err := c.httpClient.Do(req)
			if err != nil {
				panic(err)
			}
			switch resp.StatusCode {
			case http.StatusInternalServerError:
				slog.Error("", "url", url, "status", resp.StatusCode, "healthy", false)

			case http.StatusOK:
				slog.Info("", "url", url, "status", resp.StatusCode, "healthy", true)
			}
		}
	}
}
