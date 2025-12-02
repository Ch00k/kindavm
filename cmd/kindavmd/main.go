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

	// Create web server
	server := web.NewServer(*addr, handler)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := server.Run(ctx); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Println("KindaVM daemon started")
	log.Printf("Web interface: http://%s", *addr)
	log.Println("Press Ctrl+C to stop")

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received")
	cancel()

	log.Println("KindaVM daemon stopped")
}
