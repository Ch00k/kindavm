// KindaVM daemon provides a web-based interface for remote keyboard and mouse control via HID.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ch00k/kindavm/internal/events"
	"github.com/Ch00k/kindavm/internal/hid"
	"github.com/Ch00k/kindavm/internal/video"
	"github.com/Ch00k/kindavm/internal/web"
)

var Version = "dev"

func main() {
	// Command line flags
	addr := flag.String("addr", "localhost:8080", "HTTP server address")
	hidDevice := flag.String("hid", "/dev/hidg0", "HID device path")

	// Video streaming flags
	enableVideo := flag.Bool("video", false, "Enable video streaming")
	videoWidth := flag.Int("video-width", 640, "Video width in pixels")
	videoHeight := flag.Int("video-height", 480, "Video height in pixels")
	videoFramerate := flag.Int("video-framerate", 30, "Video framerate (fps)")
	videoQuality := flag.Int("video-quality", 80, "MJPEG quality (1-100)")

	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("kindavmd version %s\n", Version)
		os.Exit(0)
	}

	// Initialize HID device
	device := hid.NewDevice(*hidDevice)
	if err := device.CheckDevice(); err != nil {
		log.Printf("Warning: HID device check failed: %v", err)
		log.Printf("Make sure the HID gadget is configured correctly")
	}

	// Create event handler
	handler := events.NewHandler(device)

	// Create video streamer if enabled
	var streamer *video.MJPEGStreamer
	if *enableVideo {
		config := video.Config{
			Width:     *videoWidth,
			Height:    *videoHeight,
			Framerate: *videoFramerate,
			Quality:   *videoQuality,
		}
		streamer = video.NewMJPEGStreamer(config)
	}

	// Create web server
	server := web.NewServer(*addr, handler, streamer)

	if err := run(*addr, server, streamer); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(addr string, server *web.Server, streamer *video.MJPEGStreamer) error {
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start video streamer if enabled
	if streamer != nil {
		if err := streamer.Start(ctx); err != nil {
			return fmt.Errorf("failed to start video streamer: %w", err)
		}
		defer streamer.Stop()
		log.Println("Video streaming enabled")
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := server.Run(ctx); err != nil {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	log.Println("KindaVM daemon started")
	log.Printf("Web interface: http://%s", addr)
	log.Println("Press Ctrl+C to stop")

	// Wait for shutdown signal or server error
	select {
	case <-sigChan:
		log.Println("Shutdown signal received")
		cancel()
	case err := <-errChan:
		return err
	}

	log.Println("KindaVM daemon stopped")
	return nil
}
