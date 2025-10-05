package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
)

func (s *Server) runCommand(ctx context.Context) {
	defer close(s.cmdDone)

	var cmdCtx context.Context
	var cancel context.CancelFunc

	if s.config.Timeout > 0 {
		cmdCtx, cancel = context.WithTimeout(ctx, s.config.Timeout)
	} else {
		cmdCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	// Execute the command
	cmd := exec.CommandContext(cmdCtx, "sh", "-c", s.config.Command)

	// Get stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating stdout pipe: %v", err)
		s.broadcast <- []byte(fmt.Sprintf("Error: Failed to create stdout pipe: %v\n", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Error creating stderr pipe: %v", err)
		s.broadcast <- []byte(fmt.Sprintf("Error: Failed to create stderr pipe: %v\n", err))
		return
	}

	// Start the command
	log.Println("Starting command execution...")
	s.broadcast <- []byte(fmt.Sprintf("Starting command: %s\n", s.config.Command))

	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command: %v", err)
		s.broadcast <- []byte(fmt.Sprintf("Error: Failed to start command: %v\n", err))
		return
	}

	// Stream outputs
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		s.streamOutput(stdout, "STDOUT")
	}()

	go func() {
		defer wg.Done()
		s.streamOutput(stderr, "STDERR")
	}()

	// Wait for outputs to complete
	wg.Wait()

	// Wait for command to complete
	err = cmd.Wait()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			log.Printf("Command finished with exit code: %d", exitErr.ExitCode())
			s.broadcast <- []byte(fmt.Sprintf("\nCommand finished with exit code: %d\n", exitErr.ExitCode()))
		} else {
			log.Printf("Command finished with error: %v", err)
			s.broadcast <- []byte(fmt.Sprintf("\nCommand finished with error: %v\n", err))
		}
	} else {
		log.Println("Command finished successfully")
		s.broadcast <- []byte("\nCommand finished successfully\n")
	}

	// Notify clients that command has finished
	s.broadcast <- []byte("[STREAM ENDED - Command execution complete]\n")
}

func (s *Server) streamOutput(reader io.Reader, prefix string) {
	scanner := bufio.NewScanner(reader)
	buffer := make([]byte, 0, 64*1024) // 64KB buffer
	scanner.Buffer(buffer, 1024*1024)  // 1MB max line length

	for scanner.Scan() {
		select {
		case <-s.done:
			return
		default:
			line := scanner.Bytes()
			// Create a copy to avoid reuse issues
			lineCopy := make([]byte, len(line))
			copy(lineCopy, line)

			// Add newline and send to broadcast
			output := append(lineCopy, '\n')
			s.broadcast <- output

			// Update stats
			s.stats.mu.Lock()
			s.stats.BytesBroadcast += int64(len(output))
			s.stats.MessagesSent++
			s.stats.mu.Unlock()
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading %s: %v", prefix, err)
		s.broadcast <- []byte(fmt.Sprintf("Error reading %s: %v\n", prefix, err))
	}
}
