package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

var req int

func Test_http2_request(t *testing.T) {
	// Set up an http.Client for HTTP/2 requests.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // For testing with self-signed certs.
			},
		},
	}

	// Replace with the actual IP address of your server.
	url := "https://192.168.6.239:8442/"
	req = 1_00_0
	start := time.Now()
	for i := 0; i < req; i++ {
		resp, err := client.Get(url)
		if err != nil {
			log.Fatalf("HTTP/2 request %d failed: %v", i, err)
		}
		// Read and discard response body
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
	elapsed := time.Since(start)
	fmt.Printf("HTTP/2: Completed %d requests in %s (avg %s per request)\n", req, elapsed, elapsed/time.Duration(req))
}
