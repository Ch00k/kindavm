// Package video provides video streaming functionality for KindaVM.
package video

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os/exec"
	"sync"
	"time"
)

// MJPEGStreamer manages MJPEG video streaming from a camera
type MJPEGStreamer struct {
	config Config

	// Client management
	client   *mjpegClient
	clientMu sync.Mutex

	// Camera process
	cmd       *exec.Cmd
	cmdCancel context.CancelFunc

	// Frame distribution
	frameChan chan []byte
	running   bool
	runningMu sync.Mutex
}

// mjpegClient represents a connected client
type mjpegClient struct {
	writer http.ResponseWriter
	done   chan struct{}
}

// Config holds configuration for video streaming
type Config struct {
	Width     int
	Height    int
	Framerate int
	Quality   int
}

// NewMJPEGStreamer creates a new MJPEG streamer
func NewMJPEGStreamer(config Config) *MJPEGStreamer {
	return &MJPEGStreamer{
		config:    config,
		frameChan: make(chan []byte, 30), // Buffer up to 1 second of frames at 30fps
	}
}

// Start begins capturing and streaming video
func (s *MJPEGStreamer) Start(ctx context.Context) error {
	s.runningMu.Lock()
	if s.running {
		s.runningMu.Unlock()
		return fmt.Errorf("streamer already running")
	}
	s.running = true
	s.runningMu.Unlock()

	// Start camera capture
	if err := s.startCamera(ctx); err != nil {
		s.runningMu.Lock()
		s.running = false
		s.runningMu.Unlock()
		return fmt.Errorf("failed to start camera: %w", err)
	}

	log.Println("MJPEG streamer started")
	return nil
}

// Stop stops the video stream
func (s *MJPEGStreamer) Stop() {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if !s.running {
		return
	}

	// Stop camera
	if s.cmdCancel != nil {
		s.cmdCancel()
	}

	// Disconnect client
	s.clientMu.Lock()
	if s.client != nil {
		close(s.client.done)
		s.client = nil
	}
	s.clientMu.Unlock()

	s.running = false
	log.Println("MJPEG streamer stopped")
}

// IsRunning returns whether the streamer is currently running
func (s *MJPEGStreamer) IsRunning() bool {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	return s.running
}

// startCamera starts the rpicam-vid subprocess
func (s *MJPEGStreamer) startCamera(ctx context.Context) error {
	cmdCtx, cancel := context.WithCancel(ctx)
	s.cmdCancel = cancel

	// Build rpicam-vid command
	args := []string{
		"--timeout", "0", // Run indefinitely
		"--nopreview", // No preview
		"--width", fmt.Sprintf("%d", s.config.Width),
		"--height", fmt.Sprintf("%d", s.config.Height),
		"--framerate", fmt.Sprintf("%d", s.config.Framerate),
		"--codec", "mjpeg",
		"--quality", fmt.Sprintf("%d", s.config.Quality),
		"--output", "-", // Output to stdout
	}

	s.cmd = exec.CommandContext(cmdCtx, "rpicam-vid", args...)

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := s.cmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := s.cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start rpicam-vid: %w", err)
	}

	// Log stderr in background
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("[rpicam-vid] %s", scanner.Text())
		}
	}()

	// Read frames from stdout
	go s.readFrames(stdout)

	// Monitor process
	go func() {
		err := s.cmd.Wait()
		if err != nil && cmdCtx.Err() == nil {
			log.Printf("rpicam-vid exited with error: %v", err)
		}
		cancel()
	}()

	return nil
}

// readFrames reads MJPEG frames from the camera output
func (s *MJPEGStreamer) readFrames(r io.Reader) {
	frameStart := []byte{0xFF, 0xD8} // JPEG start marker (SOI)
	frameEnd := []byte{0xFF, 0xD9}   // JPEG end marker (EOI)

	buf := make([]byte, 4096)
	currentFrame := make([]byte, 0, 64*1024) // Pre-allocate 64KB to avoid reallocations
	maxFrameSize := 1024 * 1024              // 1MB max per frame

	for {
		n, err := r.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading frames: %v", err)
			}
			return
		}

		if n == 0 {
			continue
		}

		data := buf[:n]

		for len(data) > 0 {
			if len(currentFrame) == 0 {
				// Looking for frame start
				startIdx := bytes.Index(data, frameStart)
				if startIdx == -1 {
					// No start marker found, skip this data
					break
				}
				// Found start marker
				currentFrame = append(currentFrame, data[startIdx:]...)
				data = data[startIdx+len(frameStart):]
			} else {
				// Looking for frame end
				endIdx := bytes.Index(data, frameEnd)
				if endIdx == -1 {
					// No end marker yet, append all data
					currentFrame = append(currentFrame, data...)
					if len(currentFrame) > maxFrameSize {
						// Frame too large, discard and start over
						log.Printf("Frame too large (%d bytes), discarding", len(currentFrame))
						currentFrame = currentFrame[:0]
					}
					break
				}

				// Found end marker, complete the frame
				currentFrame = append(currentFrame, data[:endIdx+len(frameEnd)]...)

				// Send frame to channel
				frame := make([]byte, len(currentFrame))
				copy(frame, currentFrame)
				select {
				case s.frameChan <- frame:
				default:
					// Drop frame if channel is full (backpressure)
				}

				// Reset for next frame
				currentFrame = currentFrame[:0]
				data = data[endIdx+len(frameEnd):]
			}
		}
	}
}

// ServeHTTP handles HTTP requests for MJPEG stream
func (s *MJPEGStreamer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if already streaming to a client
	s.clientMu.Lock()
	if s.client != nil {
		s.clientMu.Unlock()
		http.Error(w, "Stream already in use", http.StatusConflict)
		return
	}

	// Create new client
	client := &mjpegClient{
		writer: w,
		done:   make(chan struct{}),
	}
	s.client = client
	s.clientMu.Unlock()

	// Cleanup on disconnect
	defer func() {
		s.clientMu.Lock()
		if s.client == client {
			s.client = nil
		}
		s.clientMu.Unlock()
		log.Printf("Client disconnected: %s", r.RemoteAddr)
	}()

	log.Printf("Client connected: %s", r.RemoteAddr)

	// Set headers for multipart response
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	mw := multipart.NewWriter(w)
	if err := mw.SetBoundary("frame"); err != nil {
		log.Printf("Failed to set boundary: %v", err)
		return
	}

	// Stream frames to client
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	frameCount := 0
	lastFrameCount := 0

	for {
		select {
		case <-client.done:
			return
		case <-r.Context().Done():
			return
		case frame := <-s.frameChan:
			// Create part headers
			partHeader := textproto.MIMEHeader{}
			partHeader.Set("Content-Type", "image/jpeg")
			partHeader.Set("Content-Length", fmt.Sprintf("%d", len(frame)))

			// Write part
			part, err := mw.CreatePart(partHeader)
			if err != nil {
				log.Printf("Failed to create multipart part: %v", err)
				return
			}

			if _, err := part.Write(frame); err != nil {
				log.Printf("Failed to write frame: %v", err)
				return
			}

			// Flush to ensure frame is sent immediately
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

			frameCount++

		case <-ticker.C:
			fps := frameCount - lastFrameCount
			if fps > 0 {
				log.Printf("Streaming at %d fps", fps)
			}
			lastFrameCount = frameCount
		}
	}
}
