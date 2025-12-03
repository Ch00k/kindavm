// Package web provides the HTTP server and WebSocket endpoint for remote control.
package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/Ch00k/kindavm/internal/events"
	"github.com/Ch00k/kindavm/internal/video"
	"github.com/coder/websocket"
)

//go:embed static/*
var staticFiles embed.FS

// Server represents the HTTP server with WebSocket support
type Server struct {
	handler  *events.Handler
	streamer *video.MJPEGStreamer
	addr     string
}

// NewServer creates a new web server
func NewServer(addr string, handler *events.Handler, streamer *video.MJPEGStreamer) *Server {
	return &Server{
		addr:     addr,
		handler:  handler,
		streamer: streamer,
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
		Addr:         s.addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
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

	// Video stream endpoint
	if s.streamer != nil {
		mux.HandleFunc("/stream", s.streamer.ServeHTTP)
	}

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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}
