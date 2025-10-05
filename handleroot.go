package main

import (
	"fmt"
	"net/http"
)

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Command Stream: %s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        #output { 
            font-family: 'Courier New', monospace; 
            white-space: pre; 
            background: #000; 
            color: #0f0; 
            padding: 15px; 
            border-radius: 5px;
            max-height: 80vh;
            overflow-y: auto;
        }
        .stats { 
            background: #f5f5f5; 
            padding: 10px; 
            border-radius: 5px; 
            margin-bottom: 10px;
        }
    </style>
</head>
<body>
    <h1>Command Output Stream</h1>
    <div class="stats">
        <strong>Command:</strong> %s<br>
        <strong>Connected Clients:</strong> <span id="clientCount">0</span> |
        <strong>Status:</strong> <span id="status">Connecting...</span> |
        <a href="/stats" target="_blank">Detailed Stats</a>
    </div>
    <div id="output"></div>
    
    <script>
        const output = document.getElementById('output');
        const status = document.getElementById('status');
        const clientCount = document.getElementById('clientCount');
        let reconnectAttempts = 0;
        let maxReconnectAttempts = 5;
        
        function connect() {
            const ws = new WebSocket('ws://' + location.host + '/ws');
            
            ws.onopen = function() {
                status.textContent = 'Connected';
                status.style.color = 'green';
                reconnectAttempts = 0;
            };
            
            ws.onmessage = function(event) {
                output.textContent += event.data;
                output.scrollTop = output.scrollHeight;
            };
            
            ws.onclose = function(event) {
                status.textContent = 'Disconnected';
                status.style.color = 'red';
                
                if (reconnectAttempts < maxReconnectAttempts) {
                    status.textContent = 'Reconnecting...';
                    setTimeout(connect, 1000 * Math.pow(2, reconnectAttempts));
                    reconnectAttempts++;
                } else {
                    status.textContent = 'Connection failed';
                    output.textContent += '\n\nFailed to reconnect after ' + maxReconnectAttempts + ' attempts.';
                }
            };
            
            ws.onerror = function(error) {
                console.error('WebSocket error:', error);
                status.textContent = 'Connection error';
                status.style.color = 'red';
            };
        }
        
        connect();
        
        // Periodically update client count (simplified - in real app you'd get this from server)
        setInterval(() => {
            fetch('/stats')
                .then(r => r.json())
                .then(data => {
                    clientCount.textContent = data.clients_connected;
                })
                .catch(err => console.error('Failed to fetch stats:', err));
        }, 5000);
    </script>
</body>
</html>`, s.config.Command, s.config.Command)
}
