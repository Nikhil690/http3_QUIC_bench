package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	http3 "github.com/quic-go/quic-go/http3"
)

const numRequests = 1000

func benchmarkHTTP2() {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	url := "https://localhost:8443"

	startTime := time.Now()

	for i := 0; i < numRequests; i++ {
		resp, err := client.Get(url)
		if err != nil {
			log.Println("HTTP/2 Error:", err)
			return
		}
		resp.Body.Close()
	}

	elapsed := time.Since(startTime)
	fmt.Printf("HTTP/2: %d requests took %v\n", numRequests, elapsed)
}

func TestHTTP3(t *testing.T) {
	transport := &http3.Transport{
		TLSClientConfig: generateTLSConfigB(),
	}
	defer transport.Close()

	client := &http.Client{
		Transport: transport,
	}
	url := "https://127.0.0.2:8444/nsmf-oam/v1/"
	start := time.Now()
	for range 1 {
		resp, err := client.Get(url)
		if err != nil {
			log.Println("HTTP/3 Error:", err)
			return
		}
		fmt.Printf("response: %v\n", resp.Request.URL)
	}
	elapsed := time.Since(start)
	fmt.Printf("HTTP/3: %d requests took %v\n", numRequests, elapsed)
}


func generateTLSConfigB() *tls.Config {
	// Load your certificate and key, same as the server
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		InsecureSkipVerify: true, // For testing with self-signed certs
		Certificates:       []tls.Certificate{cert},
		NextProtos:         []string{"h3-29", "h3", "hq-29"},
	}
}

// func Test(b *testing.T) {
// 	fmt.Println("Starting Benchmark...")

// 	// benchmarkHTTP2()
// 	TestHTTP3(b)
// }
