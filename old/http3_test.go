package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/quic-go/quic-go/http3"
)

func handler3(w http.ResponseWriter, r *http.Request) {
	// Simulate some processing delay
	time.Sleep(50 * time.Millisecond)
	fmt.Fprintln(w, "Hello, HTTP/3!")
}

func generateTLSConfig() *tls.Config {
	// Load your certificate and key or generate a self-signed cert for testing.
	// For production, use proper certificates.
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h3-29", "h3", "hq-29"}, // ALPN identifiers for HTTP/3
	}
}

func TestStartHTTP3(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler3)

	server := http3.Server{
		Addr:      ":8444",
		Handler:   mux,
		TLSConfig: generateTLSConfig(),
	}

	log.Println("HTTP/3 server running on https://localhost:8444")
	log.Fatal(server.ListenAndServe())
}
