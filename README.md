# cmdout2net

complete Go program that runs a single command and streams its output to multiple network clients by a WebSocket Broadcast Server.

Usage Examples

Basic usage:
bash
go run main.go -cmd "ping google.com"
With custom port and timeout:
bash
go run main.go -cmd "docker logs -f mycontainer" -port 9090 -timeout 10m
Long-running process with monitoring:
bash
go run main.go -cmd "tail -f /var/log/syslog" -host localhost -port 8080 -max-clients 50
With production settings:
bash
go run main.go \
  -cmd "while true; do echo 'Health check: $(date)'; sleep 30; done" \
  -port 8080 \
  -max-clients 100 \
  -buffer 200 \
  -ping-period 60s \
  -pong-wait 70s
Key Features

Graceful Shutdown: Proper handling of OS signals and cleanup
Command Timeout: Configurable command execution timeout
Client Management: Limits maximum clients and handles disconnections
Statistics: Real-time stats endpoint at /stats
WebSocket Health: Ping/pong keepalive mechanism
Resource Management: Configurable buffer sizes and limits
Error Handling: Comprehensive error handling and logging
Cross-Platform: Works on Windows, Linux, and macOS
This enhanced version provides production-ready features for robust command streaming with multiple clients.
