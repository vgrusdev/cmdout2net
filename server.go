package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

func NewServer(config *Config) *Server {
	return &Server{
		config:     config,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, config.BufferSize),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
		cmdDone:    make(chan struct{}),
		stats: &Stats{
			StartTime: time.Now(),
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	log.Printf("Starting command stream server on %s:%s", s.config.Host, s.config.Port)
	log.Printf("Command: %s", s.config.Command)
	log.Printf("Timeout: %v", s.config.Timeout)

	// Start the command execution
	go s.runCommand(ctx)

	// Start client management
	go s.manageClients()

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/stats", s.handleStats)
	mux.HandleFunc("/", s.handleRoot)

	// Create HTTP server
	server := &http.Server{
		Addr:    s.config.Host + ":" + s.config.Port,
		Handler: mux,
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on http://%s:%s", s.config.Host, s.config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received, closing server...")
	case err := <-serverErr:
		log.Printf("Server error: %v", err)
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Close all client connections
	s.closeAllClients()

	// Stop the command if it's still running
	close(s.done)

	// Wait for command to finish or timeout
	select {
	case <-s.cmdDone:
		log.Println("Command finished gracefully")
	case <-time.After(3 * time.Second):
		log.Println("Command shutdown timeout")
	}

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	return nil
}
