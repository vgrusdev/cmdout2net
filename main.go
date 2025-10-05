package main

import (
	//"bufio"
	"context"
	"flag"

	//"fmt"
	//"io"
	"log"
	//"net/http"
	"os"
	//"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// Config holds application configuration
type Config struct {
	Command    string
	Port       string
	Host       string
	Timeout    time.Duration
	BufferSize int
	MaxClients int
	PingPeriod time.Duration
	PongWait   time.Duration
	WriteWait  time.Duration
}

// Client represents a connected WebSocket client
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// Server manages the application state
type Server struct {
	config     *Config
	clients    map[*Client]bool
	clientsMu  sync.RWMutex
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	done       chan struct{}
	cmdDone    chan struct{}
	stats      *Stats
}

// Stats tracks server statistics
type Stats struct {
	mu               sync.RWMutex
	ClientsConnected int
	TotalClients     int
	BytesBroadcast   int64
	MessagesSent     int64
	StartTime        time.Time
}

func main() {
	// Parse command line flags
	config := parseFlags()

	// Create and start server
	server := NewServer(config)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	setupSignalHandler(cancel)

	// Start the server
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server shutdown complete")
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.Command, "cmd", "ping 8.8.8.8", "Command to execute and stream")
	flag.StringVar(&config.Port, "port", "8080", "Port to serve on")
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host to bind to")
	flag.DurationVar(&config.Timeout, "timeout", 0, "Command timeout (0 = no timeout)")
	flag.IntVar(&config.BufferSize, "buffer", 100, "Client message buffer size")
	flag.IntVar(&config.MaxClients, "max-clients", 100, "Maximum number of concurrent clients")
	flag.DurationVar(&config.PingPeriod, "ping-period", 30*time.Second, "WebSocket ping period")
	flag.DurationVar(&config.PongWait, "pong-wait", 40*time.Second, "WebSocket pong wait time")
	flag.DurationVar(&config.WriteWait, "write-wait", 10*time.Second, "WebSocket write wait time")

	flag.Parse()

	return config
}

func setupSignalHandler(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v. Shutting down...", sig)
		cancel()

		// Force exit if graceful shutdown takes too long
		time.Sleep(10 * time.Second)
		log.Println("Forcing shutdown due to timeout")
		os.Exit(1)
	}()
}
