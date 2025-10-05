package main

import "log"

func (s *Server) manageClients() {

	// Infinite loop - waits for channels
	// - clients register/unregister and maintain array of clients.
	// - broadcast. Sends broadcase messages to all clients in an array.
	// Never breaks the loop ?
	//
	for {
		select {
		case client := <-s.register:
			s.clientsMu.Lock()
			if len(s.clients) < s.config.MaxClients {
				s.clients[client] = true
				s.stats.mu.Lock()
				s.stats.ClientsConnected++
				s.stats.TotalClients++
				s.stats.mu.Unlock()
				log.Printf("Client connected. Total: %d", len(s.clients))
			} else {
				log.Printf("Max clients reached, rejecting connection")
				close(client.send)
			}
			s.clientsMu.Unlock()

		case client := <-s.unregister:
			s.clientsMu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
				s.stats.mu.Lock()
				s.stats.ClientsConnected--
				s.stats.mu.Unlock()
				log.Printf("Client disconnected. Total: %d", len(s.clients))
			}
			s.clientsMu.Unlock()

		case message := <-s.broadcast:
			s.clientsMu.RLock()
			for client := range s.clients {
				select {
				case client.send <- message:
					// Message sent successfully
				default:
					// Client buffer full, remove client
					go func(c *Client) {
						s.unregister <- c
					}(client)
				}
			}
			s.clientsMu.RUnlock()
		}
	}
}

func (s *Server) closeAllClients() {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for client := range s.clients {
		close(client.send)
		client.conn.Close()
	}
	s.clients = make(map[*Client]bool)
	s.stats.ClientsConnected = 0
}
