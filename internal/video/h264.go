// Package video provides H264 video streaming functionality.
package video

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/coder/websocket"
)

// H264Streamer manages H264 video streaming via WebSocket
type H264Streamer struct {
	config H264Config

	// Client management
	client   *h264Client
	clientMu sync.Mutex

	// Pipeline processes
	cameraCmd *exec.Cmd
	cancel    context.CancelFunc

	// Frame distribution
	running   bool
	runningMu sync.Mutex

	// Camera capabilities
	cameraModes []CameraMode
}

// h264Client represents a connected WebSocket client
type h264Client struct {
	conn *websocket.Conn
	done chan struct{}
}

// H264Config holds configuration for H264 streaming
type H264Config struct {
	Width     int
	Height    int
	Framerate int
	Bitrate   int // in kbps
}

// CameraMode represents a camera resolution mode
type CameraMode struct {
	Width  int
	Height int
}

// NewH264Streamer creates a new H264 streamer
func NewH264Streamer(config H264Config) *H264Streamer {
	modes := DetectCameraModes()
	return &H264Streamer{
		config:      config,
		cameraModes: modes,
	}
}

// GetCameraModes returns the detected camera modes
func (s *H264Streamer) GetCameraModes() []CameraMode {
	return s.cameraModes
}

// Start begins capturing and streaming H264 video
func (s *H264Streamer) Start(ctx context.Context) error {
	s.runningMu.Lock()
	if s.running {
		s.runningMu.Unlock()
		return fmt.Errorf("streamer already running")
	}
	s.running = true
	s.runningMu.Unlock()

	// Start pipeline
	if err := s.startPipeline(ctx); err != nil {
		s.runningMu.Lock()
		s.running = false
		s.runningMu.Unlock()
		return fmt.Errorf("failed to start pipeline: %w", err)
	}

	log.Println("H264 streamer started")
	return nil
}

// Stop stops the video stream
func (s *H264Streamer) Stop() {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if !s.running {
		return
	}

	// Stop pipeline
	if s.cancel != nil {
		s.cancel()
	}

	// Disconnect client
	s.clientMu.Lock()
	if s.client != nil {
		close(s.client.done)
		s.client = nil
	}
	s.clientMu.Unlock()

	s.running = false
	log.Println("H264 streamer stopped")
}

// IsRunning returns whether the streamer is currently running
func (s *H264Streamer) IsRunning() bool {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	return s.running
}

// startPipeline starts the rpicam-vid pipeline
func (s *H264Streamer) startPipeline(ctx context.Context) error {
	pipeCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// Start rpicam-vid to output raw H.264
	cameraArgs := []string{
		"--timeout", "0",
		"--nopreview",
	}

	// Only add width/height if they are non-zero (default mode uses camera defaults)
	if s.config.Width > 0 && s.config.Height > 0 {
		cameraArgs = append(cameraArgs, "--width", fmt.Sprintf("%d", s.config.Width))
		cameraArgs = append(cameraArgs, "--height", fmt.Sprintf("%d", s.config.Height))
	}

	cameraArgs = append(cameraArgs,
		"--framerate", fmt.Sprintf("%d", s.config.Framerate),
		"--codec", "h264",
		"--output", "-", // Output to stdout
	)

	// Log the command being executed
	log.Printf("Executing: rpicam-vid %s", joinArgs(cameraArgs))

	s.cameraCmd = exec.CommandContext(pipeCtx, "rpicam-vid", cameraArgs...)

	cameraStdout, err := s.cameraCmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create camera stdout pipe: %w", err)
	}

	cameraStderr, err := s.cameraCmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create camera stderr pipe: %w", err)
	}

	// Start camera process
	if err := s.cameraCmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start rpicam-vid: %w", err)
	}

	// Log stderr from camera
	go func() {
		scanner := bufio.NewScanner(cameraStderr)
		for scanner.Scan() {
			log.Printf("[rpicam-vid] %s", scanner.Text())
		}
	}()

	// Read raw H.264 from camera and stream to client
	go s.streamRawH264(cameraStdout)

	// Monitor camera process
	go func() {
		cameraErr := s.cameraCmd.Wait()
		if cameraErr != nil && pipeCtx.Err() == nil {
			log.Printf("rpicam-vid exited with error: %v", cameraErr)
		}
		cancel()
	}()

	return nil
}

// streamRawH264 reads raw H.264 from camera and sends to client
func (s *H264Streamer) streamRawH264(r io.Reader) {
	buf := make([]byte, 32*1024)

	for {
		n, err := r.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading H.264 stream: %v", err)
			}
			return
		}

		if n == 0 {
			continue
		}

		// Send raw H.264 data to client
		s.sendRawH264(buf[:n])
	}
}

func (s *H264Streamer) sendRawH264(data []byte) {
	s.clientMu.Lock()
	client := s.client
	s.clientMu.Unlock()

	if client == nil {
		return
	}

	if err := client.conn.Write(context.Background(), websocket.MessageBinary, data); err != nil {
		log.Printf("Failed to send H.264 data: %v", err)
	}
}

// HandleWebSocket handles WebSocket connections for H264 streaming
func (s *H264Streamer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for dynamic configuration
	queryParams := r.URL.Query()
	width := s.config.Width
	height := s.config.Height
	framerate := s.config.Framerate

	if w := queryParams.Get("width"); w != "" {
		if parsedWidth, err := parseIntParam(w); err == nil && parsedWidth > 0 {
			width = parsedWidth
		}
	}
	if h := queryParams.Get("height"); h != "" {
		if parsedHeight, err := parseIntParam(h); err == nil && parsedHeight > 0 {
			height = parsedHeight
		}
	}
	if f := queryParams.Get("framerate"); f != "" {
		if parsedFramerate, err := parseIntParam(f); err == nil && parsedFramerate > 0 {
			framerate = parsedFramerate
		}
	}

	// Check if already streaming to a client
	s.clientMu.Lock()
	if s.client != nil {
		s.clientMu.Unlock()
		http.Error(w, "Stream already in use", http.StatusConflict)
		return
	}

	// Accept WebSocket connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		s.clientMu.Unlock()
		log.Printf("Failed to accept WebSocket: %v", err)
		return
	}

	// Create new client
	client := &h264Client{
		conn: conn,
		done: make(chan struct{}),
	}
	s.client = client
	s.clientMu.Unlock()

	// Start or reconfigure streamer
	needsStart := !s.IsRunning()
	needsRestart := s.IsRunning() &&
		(width != s.config.Width || height != s.config.Height || framerate != s.config.Framerate)

	if needsRestart {
		log.Printf("Reconfiguring stream: %dx%d @ %dfps", width, height, framerate)
		s.Stop()
		s.config.Width = width
		s.config.Height = height
		s.config.Framerate = framerate
		needsStart = true
	}

	if needsStart {
		if !needsRestart {
			log.Printf("Starting stream: %dx%d @ %dfps", width, height, framerate)
			s.config.Width = width
			s.config.Height = height
			s.config.Framerate = framerate
		}

		ctx := r.Context()
		if err := s.Start(ctx); err != nil {
			log.Printf("Failed to start stream: %v", err)
			s.clientMu.Lock()
			s.client = nil
			s.clientMu.Unlock()
			_ = conn.Close(websocket.StatusInternalError, "Failed to start stream")
			return
		}
	}

	// Cleanup on disconnect
	defer func() {
		s.clientMu.Lock()
		if s.client == client {
			s.client = nil
		}
		s.clientMu.Unlock()
		_ = conn.Close(websocket.StatusNormalClosure, "")

		// Stop the stream when client disconnects
		s.Stop()

		log.Printf("H264 client disconnected: %s", r.RemoteAddr)
	}()

	log.Printf("H264 client connected: %s (%dx%d @ %dfps)", r.RemoteAddr, width, height, framerate)

	// Keep connection alive and wait for disconnect
	for {
		_, _, err := conn.Read(r.Context())
		if err != nil {
			return
		}
	}
}

func parseIntParam(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		// Quote arguments that contain spaces
		if len(arg) > 0 && (arg[0] == '-' || !hasSpace(arg)) {
			result += arg
		} else {
			result += "'" + arg + "'"
		}
	}
	return result
}

func hasSpace(s string) bool {
	for _, c := range s {
		if c == ' ' {
			return true
		}
	}
	return false
}

// DetectCameraModes queries the camera for available resolution modes
func DetectCameraModes() []CameraMode {
	log.Println("Detecting camera modes...")

	cmd := exec.Command("rpicam-vid", "--list-cameras")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to detect camera modes: %v", err)
		return getDefaultModes()
	}

	modes := parseCameraModes(string(output))
	if len(modes) == 0 {
		log.Println("No camera modes detected, using defaults")
		return getDefaultModes()
	}

	log.Printf("Detected %d camera modes", len(modes))
	return modes
}

// parseCameraModes parses the output of rpicam-vid --list-cameras
func parseCameraModes(output string) []CameraMode {
	var modes []CameraMode
	seen := make(map[string]bool)

	// Pattern to match mode lines like: "1536x864 [120.13 fps - ..."
	// We look for resolution followed by fps info to distinguish from crop/sensor dimensions
	re := regexp.MustCompile(`(\d+)x(\d+)\s+\[[\d.]+\s+fps`)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			width, err1 := strconv.Atoi(matches[1])
			height, err2 := strconv.Atoi(matches[2])

			if err1 == nil && err2 == nil {
				key := fmt.Sprintf("%dx%d", width, height)
				if !seen[key] {
					seen[key] = true
					modes = append(modes, CameraMode{
						Width:  width,
						Height: height,
					})
				}
			}
		}
	}

	return modes
}

// getDefaultModes returns a sensible default set of resolution modes
func getDefaultModes() []CameraMode {
	return []CameraMode{
		{Width: 640, Height: 480},
		{Width: 800, Height: 600},
		{Width: 1024, Height: 768},
		{Width: 1280, Height: 720},
		{Width: 1920, Height: 1080},
	}
}
