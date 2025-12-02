// Package hid provides HID device interaction for keyboard and mouse control.
package hid

import (
	"fmt"
	"os"
	"time"
)

// Default configuration values
const (
	DefaultHIDDevice = "/dev/hidg0"
	DefaultDelayMS   = 10
)

// Device represents a HID device interface
type Device struct {
	path string
}

// NewDevice creates a new HID device interface
func NewDevice(path string) *Device {
	if path == "" {
		path = DefaultHIDDevice
	}
	return &Device{path: path}
}

// SendReport sends a HID report to the device
func (d *Device) SendReport(report []byte, delayMS int) error {
	f, err := os.OpenFile(d.path, os.O_RDWR, 0o666)
	if err != nil {
		return fmt.Errorf("failed to open HID device: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.Write(report)
	if err != nil {
		return fmt.Errorf("failed to write HID report: %w", err)
	}

	if delayMS > 0 {
		time.Sleep(time.Duration(delayMS) * time.Millisecond)
	}

	return nil
}

// CheckDevice verifies that the HID device is available
func (d *Device) CheckDevice() error {
	info, err := os.Stat(d.path)
	if err != nil {
		return fmt.Errorf("HID device not found: %w", err)
	}

	mode := info.Mode()
	if mode&os.ModeCharDevice == 0 {
		return fmt.Errorf("path is not a character device: %s", d.path)
	}

	return nil
}
