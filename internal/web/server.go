// Package web provides the HTTP server and WebSocket endpoint for remote control.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/Ch00k/kindavm/internal/events"
	"github.com/coder/websocket"
)

//go:embed static/*
var staticFiles embed.FS

// VideoSettings holds configurable ustreamer parameters
type VideoSettings struct {
	Quality    int  `json:"quality"`
	DesiredFPS int  `json:"desiredFps"`
	Buffers    int  `json:"buffers"`
	TCPNodelay bool `json:"tcpNodelay"`
}

// DefaultVideoSettings returns the default video settings
func DefaultVideoSettings() VideoSettings {
	return VideoSettings{
		Quality:    80,
		DesiredFPS: 30,
		Buffers:    5,
		TCPNodelay: false,
	}
}

// Server represents the HTTP server with WebSocket support
type Server struct {
	handler       *events.Handler
	addr          string
	ustreamerAddr string
	videoDevice   string

	// ustreamer process management
	ustreamerCmd *exec.Cmd
	ustreamerMu  sync.Mutex

	// Video settings
	videoSettings VideoSettings
	settingsMu    sync.RWMutex
}

// NewServer creates a new web server
func NewServer(
	addr string,
	handler *events.Handler,
	ustreamerAddr string,
	videoDevice string,
) *Server {
	return &Server{
		addr:          addr,
		handler:       handler,
		ustreamerAddr: ustreamerAddr,
		videoDevice:   videoDevice,
		videoSettings: DefaultVideoSettings(),
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Serve static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("failed to create static file system: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// WebSocket endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)

	log.Printf("Starting server on %s", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // No origin checking for now
	})
	if err != nil {
		log.Printf("Failed to accept WebSocket: %v", err)
		return
	}
	defer func() {
		_ = conn.Close(websocket.StatusInternalError, "unexpected close")
	}()

	log.Printf("WebSocket connection established from %s", r.RemoteAddr)

	ctx := r.Context()
	if err := s.handleConnection(ctx, conn); err != nil {
		log.Printf("WebSocket error: %v", err)
	}

	_ = conn.Close(websocket.StatusNormalClosure, "")
	log.Printf("WebSocket connection closed from %s", r.RemoteAddr)
}

// handleHostname returns the server hostname as JSON
func (s *Server) handleHostname(w http.ResponseWriter, _ *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Error getting hostname: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"hostname": hostname}); err != nil {
		log.Printf("Error encoding hostname: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleConfig returns the client configuration as JSON
func (s *Server) handleConfig(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, ustreamerPort, err := net.SplitHostPort(s.ustreamerAddr)
	if err != nil {
		log.Printf("Error parsing ustreamer address: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(map[string]string{
		"ustreamerPort": ustreamerPort,
	}); err != nil {
		log.Printf("Error encoding config: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// startUstreamer starts the ustreamer process
func (s *Server) startUstreamer() error {
	s.ustreamerMu.Lock()
	defer s.ustreamerMu.Unlock()

	if s.ustreamerCmd != nil {
		return nil // Already running
	}

	ustreamerHost, ustreamerPort, err := net.SplitHostPort(s.ustreamerAddr)
	if err != nil {
		return fmt.Errorf("invalid ustreamer address: %w", err)
	}

	s.settingsMu.RLock()
	settings := s.videoSettings
	s.settingsMu.RUnlock()

	args := []string{
		"--persistent",
		"--device", s.videoDevice,
		"--dv-timings",
		"--host", ustreamerHost,
		"--port", ustreamerPort,
		"--resolution", "1280x720",
		"--format", "UYVY",
		"--quality", fmt.Sprintf("%d", settings.Quality),
		"--buffers", fmt.Sprintf("%d", settings.Buffers),
		"--drop-same-frames", "30",
		"--slowdown",
	}

	if settings.DesiredFPS > 0 {
		args = append(args, "--desired-fps", fmt.Sprintf("%d", settings.DesiredFPS))
	}

	if settings.TCPNodelay {
		args = append(args, "--tcp-nodelay")
	}

	log.Printf("Starting ustreamer on %s with settings: quality=%d, fps=%d, buffers=%d, tcp-nodelay=%v",
		s.ustreamerAddr, settings.Quality, settings.DesiredFPS, settings.Buffers, settings.TCPNodelay)
	s.ustreamerCmd = exec.Command("ustreamer", args...)

	s.ustreamerCmd.Stderr = os.Stderr
	s.ustreamerCmd.Stdout = os.Stdout

	if err := s.ustreamerCmd.Start(); err != nil {
		s.ustreamerCmd = nil
		return fmt.Errorf("failed to start ustreamer: %w", err)
	}

	log.Printf("ustreamer started (PID: %d)", s.ustreamerCmd.Process.Pid)
	return nil
}

// stopUstreamer stops the ustreamer process
func (s *Server) stopUstreamer() {
	s.ustreamerMu.Lock()
	defer s.ustreamerMu.Unlock()

	if s.ustreamerCmd == nil {
		return
	}

	if s.ustreamerCmd.Process != nil {
		log.Printf("Stopping ustreamer (PID: %d)...", s.ustreamerCmd.Process.Pid)
		if err := s.ustreamerCmd.Process.Kill(); err != nil {
			log.Printf("Error killing ustreamer: %v", err)
		}
		// Wait for the process to exit to avoid zombie processes
		if err := s.ustreamerCmd.Wait(); err != nil {
			log.Printf("Error waiting for ustreamer to exit: %v", err)
		}
	}

	s.ustreamerCmd = nil
	log.Println("ustreamer stopped")
}

// handleStartVideo starts the ustreamer process
func (s *Server) handleStartVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.startUstreamer(); err != nil {
		log.Printf("Failed to start ustreamer: %v", err)
		http.Error(w, "Failed to start video stream", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "started"}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// handleStopVideo stops the ustreamer process
func (s *Server) handleStopVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.stopUstreamer()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "stopped"}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// handleGetSettings returns the current video settings
func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.settingsMu.RLock()
	settings := s.videoSettings
	s.settingsMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(settings); err != nil {
		log.Printf("Error encoding settings: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleUpdateSettings updates the video settings
func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var settings VideoSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		log.Printf("Error decoding settings: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if settings.Quality < 1 || settings.Quality > 100 {
		http.Error(w, "Quality must be between 1 and 100", http.StatusBadRequest)
		return
	}

	if settings.Buffers < 2 || settings.Buffers > 10 {
		http.Error(w, "Buffers must be between 2 and 10", http.StatusBadRequest)
		return
	}

	if settings.DesiredFPS < 0 || settings.DesiredFPS > 60 {
		http.Error(w, "Desired FPS must be between 0 and 60", http.StatusBadRequest)
		return
	}

	s.settingsMu.Lock()
	s.videoSettings = settings
	s.settingsMu.Unlock()

	log.Printf("Updated video settings: quality=%d, fps=%d, buffers=%d, tcp-nodelay=%v",
		settings.Quality, settings.DesiredFPS, settings.Buffers, settings.TCPNodelay)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "updated"}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// handleConnection handles messages from a WebSocket connection
func (s *Server) handleConnection(ctx context.Context, conn *websocket.Conn) error {
	for {
		msgType, data, err := conn.Read(ctx)
		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		if msgType != websocket.MessageText {
			log.Printf("Warning: Received non-text message, ignoring")
			continue
		}

		// Process the event
		if err := s.handler.HandleEvent(data); err != nil {
			log.Printf("Error handling event: %v", err)
			// Don't close connection on event handling errors, just log them
		}
	}
}

// Run starts the server with graceful shutdown support
func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:        s.addr,
		ReadTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	mux := http.NewServeMux()

	// Serve static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("failed to create static file system: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// WebSocket endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)

	// API endpoints
	mux.HandleFunc("/hostname", s.handleHostname)
	mux.HandleFunc("/config", s.handleConfig)
	mux.HandleFunc("/video/start", s.handleStartVideo)
	mux.HandleFunc("/video/stop", s.handleStopVideo)
	mux.HandleFunc("/settings", s.handleGetSettings)
	mux.HandleFunc("/settings/update", s.handleUpdateSettings)

	srv.Handler = mux

	// Run server in goroutine so we can listen for context cancellation
	errChan := make(chan error, 1)
	go func() {
		log.Printf("Starting server on %s", s.addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Println("Shutting down server...")
		s.stopUstreamer()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errChan:
		s.stopUstreamer()
		return err
	}
}
