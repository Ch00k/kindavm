#!/usr/bin/env bash

set -e

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root (use sudo)" >&2
    exit 1
fi

# Install dependencies
echo "Installing dependencies..."
apt-get update
apt-get install --no-install-recommends --no-install-suggests --yes python3-pip python3-venv

echo "Installing KindaVM..."

# Create virtualenv
VENV_PATH="/opt/kindavm/venv"
if [ -d "$VENV_PATH" ]; then
    echo "Virtualenv already exists at $VENV_PATH"
else
    echo "Creating virtualenv at $VENV_PATH..."
    mkdir -p /opt/kindavm
    python3 -m venv "$VENV_PATH"
fi

# Configure boot-time USB Gadget support
if grep -q '^dtoverlay=dwc2$' /boot/firmware/config.txt; then
    echo "USB Gadget support already enabled in /boot/firmware/config.txt"
else
    echo "Enabling USB Gadget support in /boot/firmware/config.txt"
    echo 'dtoverlay=dwc2' | tee -a /boot/firmware/config.txt
fi

# Configure automatic module loading
if [ -f /etc/modules-load.d/kindavm.conf ]; then
    echo "Module loading configuration already exists at /etc/modules-load.d/kindavm.conf"
else
    echo "Creating module loading configuration at /etc/modules-load.d/kindavm.conf"
    echo -en 'dwc2\nlibcomposite\n' | tee /etc/modules-load.d/kindavm.conf
fi

# Install Python package
echo "Installing Python package..."
"$VENV_PATH/bin/pip" install --upgrade pip
"$VENV_PATH/bin/pip" install --upgrade .

# Create wrapper script for kinda command
echo "Creating kinda wrapper script..."
cat > /usr/local/bin/kinda << 'EOF'
#!/bin/bash
exec /opt/kindavm/venv/bin/python -m kindavm.cli "$@"
EOF
chmod +x /usr/local/bin/kinda

# Install init and uninstall scripts
echo "Installing HID initialization script..."
cp init_hid.sh /usr/local/bin/kindavm-init-hid.sh
chmod +x /usr/local/bin/kindavm-init-hid.sh

echo "Installing uninstall script..."
cp uninstall.sh /usr/local/bin/kindavm-uninstall.sh
chmod +x /usr/local/bin/kindavm-uninstall.sh

# Install systemd service
echo "Installing systemd service..."
cp kindavm.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable kindavm.service

echo ""
echo "Installation complete!"
echo ""

# Check if system needs reboot
NEEDS_REBOOT=false
if ! grep -q '^dtoverlay=dwc2$' /boot/firmware/config.txt 2>/dev/null; then
    NEEDS_REBOOT=true
fi

if ! lsmod | grep -q '^dwc2\|^libcomposite'; then
    NEEDS_REBOOT=true
fi

if [ "$NEEDS_REBOOT" = true ]; then
    echo "IMPORTANT: A reboot is required to enable USB gadget support."
    echo "After rebooting, the HID gadget will be initialized automatically."
    echo ""
    echo "To reboot now: sudo reboot"
    echo ""
else
    echo "USB gadget support is already enabled."
    echo "The HID gadget will be initialized on next boot, or you can initialize it now:"
    echo "  sudo systemctl start kindavm.service"
    echo ""
fi

echo "Usage:"
echo "  kinda type 'Hello World'     - Type text"
echo "  kinda mouse move 10 20       - Move mouse"
echo "  kinda mouse click left       - Click mouse button"
echo "  kinda mouse scroll -5        - Scroll wheel"
echo ""
echo "To uninstall: sudo kindavm-uninstall.sh"
