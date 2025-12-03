# kindavmd - KindaVM Daemon

The `kindavmd` daemon provides a web-based interface for controlling a target machine via HID (keyboard and mouse) over a network.

## Features

- **Web Interface**: Access control via any modern web browser
- **Keyboard Control**: Full keyboard input with modifier key support
- **Mouse Control**: Mouse movement, clicking, and scrolling via Pointer Lock API
- **WebSocket Communication**: Real-time, low-latency event transmission
- **No Video (Phase 2)**: Currently keyboard and mouse only; video streaming will be added in Phase 3

## Building

```bash
# Build for your local platform
make build

# Cross-compile for Raspberry Pi Zero 2W with 64-bit OS (ARMv8 AArch64)
make build-arm64

# Cross-compile for 32-bit ARM (ARMv7) - older Pi models or 32-bit OS
make build-arm

# Or directly with go
go build -o kindavmd ./cmd/kindavmd

# Cross-compile manually for 64-bit
GOOS=linux GOARCH=arm64 go build -o kindavmd-arm64 ./cmd/kindavmd

# Cross-compile manually for 32-bit
GOOS=linux GOARCH=arm GOARM=7 go build -o kindavmd-arm ./cmd/kindavmd
```

**For Raspberry Pi Zero 2W:**
- Use `build-arm64` if running 64-bit Raspberry Pi OS (produces `dist/kindavmd-arm64`)
- Use `build-arm` if running 32-bit Raspberry Pi OS (produces `dist/kindavmd-arm`)

The Pi Zero 2W has an ARM Cortex-A53 (ARMv8-A) processor that supports both 32-bit and 64-bit modes.

## Running

```bash
# Run with default settings (localhost:8080, /dev/hidg0)
./kindavmd

# Specify custom address and HID device
./kindavmd -addr localhost:9000 -hid /dev/hidg1
```

### Command Line Options

- `-addr`: HTTP server address (default: `localhost:8080`)
- `-hid`: HID device path (default: `/dev/hidg0`)

## Usage

1. Start the daemon on your Raspberry Pi Zero 2W (or any device with HID gadget configured)
2. Access the web interface via SSH port forward or Tailscale:
   ```bash
   # SSH port forward example
   ssh -L 8080:localhost:8080 pi@raspberrypi.local
   ```
3. Open your browser to `http://localhost:8080`
4. Click the control area to activate keyboard/mouse capture
5. Control the target machine
6. Press ESC to release control

## Architecture

The daemon consists of several components:

- **HID Module** (`internal/hid`): Generates HID reports for keyboard and mouse
- **Events Module** (`internal/events`): Translates browser events to HID reports
- **Web Module** (`internal/web`): HTTP server with WebSocket endpoint and static file serving
- **Main** (`cmd/kindavmd`): Daemon entry point with signal handling

## Browser Compatibility

The web interface requires:
- **Pointer Lock API**: For capturing mouse movement (all modern browsers)
- **WebSocket**: For real-time communication (all modern browsers)
- **JavaScript ES6+**: For client-side logic

Tested on:
- Chrome/Chromium 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Security

**Current Implementation**: No authentication or encryption

**Recommended Access Methods**:
- SSH port forwarding (encrypted tunnel)
- Tailscale (VPN with authentication)
- Local network only (not exposed to internet)

Authentication will be added in future versions if needed.

## Troubleshooting

### HID device not found

Make sure the HID gadget is configured correctly. Run the initialization script:
```bash
sudo ./init_hid.sh
```

### WebSocket connection fails

Check that:
1. The daemon is running
2. The address is correct
3. Firewall allows connections (if not using SSH tunnel)

### Keys not working

Verify the HID gadget is connected to the target machine via USB cable.

## Next Steps (Phase 3)

- Camera integration for video streaming
- MJPEG support
- H264 hardware encoding
- Video quality tuning

## License

See project LICENSE file.
