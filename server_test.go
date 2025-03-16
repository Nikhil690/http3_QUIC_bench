package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quic-go/quic-go/http3"
)

// generateTLSConfig loads the TLS certificate and key, and sets the proper ALPN tokens.
func generateTLSConfig(proto string) *tls.Config {
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("failed to load TLS key pair: %v", err)
	}
	if proto == "http3" {
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h3-29", "h3", "hq-29"},
		}
	} else if proto == "http2" {
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h2", "http/1.1"},
		}
	} else {
		panic("unknown protocol")
	}
}

func Test_http2Server(T *testing.T) {
	// Create a Gin router with logging and recovery middleware.
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Define a simple endpoint.
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, HTTP/2 with Gin!")
	})

	// Configure the HTTP/2 server.
	port := 8442
	addr := fmt.Sprintf("192.168.6.239:%d", port)
	server := &http.Server{
		Addr:      addr,
		Handler:   router,
		TLSConfig: generateTLSConfig("http3"),
	}

	log.Printf("Starting HTTP/2 server on https://192.168.6.239:%d", port)
	if err := server.ListenAndServeTLS("server.crt", "server.key"); err != nil {
		log.Fatalf("HTTP/2 server failed: %v", err)
	}
}

func Test_http3server(T *testing.T) {
	// Create a Gin router with logging and recovery middleware.
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Define a simple endpoint.
	router.GET("/", func(c *gin.Context) {
		c.String(200, "Hello, HTTP/3 with Gin!")
	})

	// Configure the HTTP/3 server.
	port := 8443
	addr := fmt.Sprintf("192.168.6.239:%d", port)
	server := http3.Server{
		Addr:      addr,
		Handler:   router,
		TLSConfig: generateTLSConfig("http3"),
	}

	log.Printf("Starting HTTP/3 server on https://192.168.6.239:%d", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("HTTP/3 server failed: %v", err)
	}
}

func Test_server(T *testing.T) {
	// Create a Gin router
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Define a simple root endpoint
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, Gin benchmarking!")
	})

	// Define a benchmarking endpoint that simulates processing delay
	router.GET("/benchmark", func(c *gin.Context) {
		// Simulate some processing delay (adjust as needed)
		time.Sleep(50 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{
			"message": "Benchmark endpoint",
			"time":    time.Now().Format(time.RFC3339Nano),
		})
	})

	// Define ports for HTTP/2 and HTTP/3
	http2Port := 8442
	http3Port := 8443

	// Create the HTTP/2 server using Go's standard library
	http2Server := &http.Server{
		Addr:      fmt.Sprintf(":%d", http2Port),
		Handler:   router,
		TLSConfig: generateTLSConfig("http2"),
	}

	// Create the HTTP/3 server using quic-go's http3 package
	http3Server := http3.Server{
		Addr:      fmt.Sprintf(":%d", http3Port),
		Handler:   router,
		TLSConfig: generateTLSConfig("http3"),
	}

	// Channel to listen for termination signals (Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Run the HTTP/2 server in a goroutine
	go func() {
		log.Printf("Starting HTTP/2 server on https://localhost:%d\n", http2Port)
		// ListenAndServeTLS is used for HTTP/2 over TLS
		if err := http2Server.ListenAndServeTLS("server.crt", "server.key"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP/2 server failed: %v", err)
		}
	}()

	// Run the HTTP/3 server in a goroutine
	go func() {
		log.Printf("Starting HTTP/3 server on https://localhost:%d\n", http3Port)
		if err := http3Server.ListenAndServe(); err != nil {
			log.Fatalf("HTTP/3 server failed: %v", err)
		}
	}()

	// Wait for termination signal
	<-quit
	log.Println("Shutdown signal received, shutting down servers...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown HTTP/2 server
	if err := http2Server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP/2 server forced to shutdown: %v", err)
	}
	// Shutdown HTTP/3 server
	if err := http3Server.Close(); err != nil {
		log.Fatalf("HTTP/3 server forced to shutdown: %v", err)
	}

	log.Println("Servers stopped gracefully")
}
