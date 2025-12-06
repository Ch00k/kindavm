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

	// Create video streamer with default config (0x0 means use camera defaults)
	config := video.H264Config{
		Width:     0,
		Height:    0,
		Framerate: 30,
		Bitrate:   2000,
	}
	h264Streamer := video.NewH264Streamer(config)

	// Create web server
	server := web.NewServer(*addr, handler, h264Streamer)

	if err := run(*addr, server); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(addr string, server *web.Server) error {
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
