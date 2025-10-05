package main

import (
	"fmt"
	"net/http"
	"time"
)

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	uptime := time.Since(s.stats.StartTime)

	stats := struct {
		ClientsConnected int    `json:"clients_connected"`
		TotalClients     int    `json:"total_clients"`
		BytesBroadcast   int64  `json:"bytes_broadcast"`
		MessagesSent     int64  `json:"messages_sent"`
		Uptime           string `json:"uptime"`
		Command          string `json:"command"`
	}{
		ClientsConnected: s.stats.ClientsConnected,
		TotalClients:     s.stats.TotalClients,
		BytesBroadcast:   s.stats.BytesBroadcast,
		MessagesSent:     s.stats.MessagesSent,
		Uptime:           uptime.String(),
		Command:          s.config.Command,
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"clients_connected": %d,
		"total_clients": %d,
		"bytes_broadcast": %d,
		"messages_sent": %d,
		"uptime": "%s",
		"command": "%s"
	}`, stats.ClientsConnected, stats.TotalClients, stats.BytesBroadcast,
		stats.MessagesSent, stats.Uptime, stats.Command)
}
