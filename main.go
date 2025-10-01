package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	conf, err := GetConfig()
	if err != nil {
		slog.Error("Error getting config", "error", err)
		os.Exit(1)
	}
	validUrls := ValidateUrls(conf.Urls)
	conf.Urls = validUrls

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
		go HealthCheck(url, c, sigChan)
	}
}

func HealthCheck(url string, c Config, sigChan chan os.Signal) {
	for {
		select {
		case s := <-sigChan:
			slog.Info("stopping health check for URL", "url", url, "sig", s)
		default:
			slog.Info("Checking health of url", "url", url)
			time.Sleep(time.Duration(c.TimeoutSeconds) * time.Second)
		}
	}
}
