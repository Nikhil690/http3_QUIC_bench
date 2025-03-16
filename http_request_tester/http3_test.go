package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/quic-go/quic-go/http3"
)

func Test_http3_request(t *testing.T) {
	// Set up an http3.RoundTripper for HTTP/3 requests.
	roundTripper := &http3.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // For testing with self-signed certs.
			NextProtos:         []string{"h3", "h3-29"},
		},
	}
	defer roundTripper.Close()

	client := &http.Client{
		Transport: roundTripper,
	}

	// Replace with the actual IP address of your server.
	url := "https://192.168.6.239:8443/"
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
