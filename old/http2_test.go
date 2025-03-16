package main

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// Simulate some processing delay
	time.Sleep(50 * time.Millisecond)
	fmt.Fprintln(w, "Hello, HTTP/2!")
}

func TestStartHTTP2(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	server := &http.Server{
		Addr:    ":8443",
		Handler: mux,
	}
	
	log.Println("HTTP/2 server running on https://localhost:8443")
	log.Fatal(server.ListenAndServeTLS("cert.pem", "key.pem"))
}
