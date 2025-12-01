# KindaVM

USB HID keyboard and mouse emulation for KVM usage on Linux systems with USB gadget support.

## Overview

KindaVM turns a Linux device (like a Raspberry Pi Zero) into a USB HID device that can emulate both a keyboard and mouse. This is useful for KVM (Keyboard, Video, Mouse) applications where you want to control another computer over USB.

## Features

- **Keyboard Emulation**: Type text strings and send individual key presses
- **Mouse Emulation**: Move cursor, click buttons, scroll wheel, and drag
- **Single Composite Device**: Keyboard and mouse combined in one USB device
- **Simple CLI**: Easy-to-use command-line interface
- **Auto-initialization**: Systemd service for automatic setup on boot

## Requirements

- Linux system with USB gadget support (e.g., Raspberry Pi Zero, Pi 4)
- Kernel with `configfs` and `libcomposite` enabled
- Python 3.8 or higher
- Root access for USB gadget configuration

## Installation

```bash
sudo ./install.sh
```

The installation script will:
1. Create a virtualenv at `/opt/kindavm/venv`
2. Install the Python package in the virtualenv
3. Create the `kinda` wrapper script in `/usr/local/bin/`
4. Copy the HID initialization script to `/usr/local/bin/`
5. Install and enable the systemd service
6. Initialize the HID gadget

## Usage

### Keyboard Commands

Type a text string:
```bash
kinda type "Hello World"
```

Type from stdin:
```bash
echo "test text" | kinda type
```

Send a specific key with modifiers:
```bash
kinda key 0x04 0x02  # Send 'A' with left shift
```

### Mouse Commands

Move the mouse cursor:
```bash
kinda mouse move 10 20      # Move 10 pixels right, 20 pixels down
kinda mouse move -5 -10     # Move 5 pixels left, 10 pixels up
```

Click mouse buttons:
```bash
kinda mouse click left      # Left click
kinda mouse click right     # Right click
kinda mouse click middle    # Middle click
```

Scroll the wheel:
```bash
kinda mouse scroll -5       # Scroll up
kinda mouse scroll 5        # Scroll down
```

Drag with button held:
```bash
kinda mouse drag 50 50 left    # Drag 50 pixels right and down with left button
```

## Architecture

KindaVM uses a single composite USB HID device with multiple report IDs:

- **Report ID 1**: Keyboard (modifiers + 6-key rollover)
- **Report ID 2**: Consumer Control (media keys, volume, brightness)
- **Report ID 3**: System Control (power, sleep, wake)
- **Report ID 4**: Mouse (3 buttons + X/Y movement + scroll wheel)

All reports are sent through `/dev/hidg0`.

## Technical Details

### HID Report Descriptor

The HID descriptor defines a composite device with keyboard and mouse functionality. See `hid_report_desc.md` for the complete byte-by-byte breakdown.

### Mouse Report Format (Report ID 4)
```
Byte 0: Report ID (0x04)
Byte 1: Buttons (bit 0: left, bit 1: right, bit 2: middle)
Byte 2: X movement (signed 8-bit, -127 to +127)
Byte 3: Y movement (signed 8-bit, -127 to +127)
Byte 4: Wheel (signed 8-bit, -127 to +127)
```

### Keyboard Report Format (Report ID 1)
```
Byte 0: Report ID (0x01)
Byte 1: Modifier keys
Byte 2: Reserved
Byte 3-8: Key codes (up to 6 simultaneous keys)
```

## File Structure

```
kindavm/
├── pyproject.toml          # Python package configuration
├── init_hid.sh             # USB gadget initialization script
├── hid_report_desc.md      # HID descriptor documentation
├── kindavm/                # Python package
│   ├── __init__.py
│   ├── hid.py              # Core HID report sending
│   ├── keyboard.py         # Keyboard functionality
│   ├── mouse.py            # Mouse functionality
│   └── cli.py              # Command-line interface
├── kindavm.service         # Systemd service
├── install.sh              # Installation script
└── README.md               # This file
```

## Troubleshooting

### HID device not found

If you get an error about `/dev/hidg0` not existing:

```bash
sudo /usr/local/bin/kindavm-init-hid.sh
```

### No UDC (USB Device Controller) found

Your system may not support USB gadget mode. Check:
- Raspberry Pi Zero: Should work out of the box with `dwc2` overlay
- Raspberry Pi 4: May need `dtoverlay=dwc2` in `/boot/config.txt`
- Check `ls /sys/class/udc` to see if any controllers are available

### Permission denied

The `kinda` command needs access to `/dev/hidg0`. You may need to:
- Run commands with `sudo`
- Or add udev rules to allow your user access to the device

## Development

The package is installed in a virtualenv at `/opt/kindavm/venv`. The `kinda` command in `/usr/local/bin/` is a wrapper that calls the virtualenv's Python interpreter.

To test changes during development, you can either:
1. Run commands directly with the virtualenv Python:
   ```bash
   /opt/kindavm/venv/bin/python -m kindavm.cli type "test"
   ```

2. Or reinstall the package after making changes:
   ```bash
   sudo /opt/kindavm/venv/bin/pip install .
   ```

## Credits

Based on the [zero-keyboard](https://github.com/example/zero-keyboard) project, extended with mouse support and packaged as a proper Python application.

## License

MIT License - See LICENSE file for details
