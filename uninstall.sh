#!/usr/bin/env bash

set -e

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root (use sudo)" >&2
    exit 1
fi

echo "Uninstalling KindaVM..."

# Stop and disable systemd service
if systemctl is-active --quiet kindavm.service; then
    echo "Stopping kindavm service..."
    systemctl stop kindavm.service
fi

if systemctl is-enabled --quiet kindavm.service 2>/dev/null; then
    echo "Disabling kindavm service..."
    systemctl disable kindavm.service
fi

# Remove systemd service file
if [ -f /etc/systemd/system/kindavm.service ]; then
    echo "Removing systemd service file..."
    rm -f /etc/systemd/system/kindavm.service
    systemctl daemon-reload
fi

# Unbind and remove USB gadget
if [ -f /sys/kernel/config/usb_gadget/kindavm/UDC ]; then
    echo "Unbinding USB gadget..."
    echo "" > /sys/kernel/config/usb_gadget/kindavm/UDC 2>/dev/null || true
fi

if [ -d /sys/kernel/config/usb_gadget/kindavm ]; then
    echo "Removing USB gadget configuration..."
    rm -rf /sys/kernel/config/usb_gadget/kindavm
fi

# Remove kinda wrapper script
if [ -f /usr/local/bin/kinda ]; then
    echo "Removing kinda wrapper script..."
    rm -f /usr/local/bin/kinda
fi

# Remove init script
if [ -f /usr/local/bin/kindavm-init-hid.sh ]; then
    echo "Removing HID initialization script..."
    rm -f /usr/local/bin/kindavm-init-hid.sh
fi

# Remove uninstall script (this script)
if [ -f /usr/local/bin/kindavm-uninstall.sh ]; then
    echo "Removing uninstall script..."
    rm -f /usr/local/bin/kindavm-uninstall.sh
fi

# Remove module loading configuration
if [ -f /etc/modules-load.d/kindavm.conf ]; then
    echo "Removing module loading configuration..."
    rm -f /etc/modules-load.d/kindavm.conf
fi

# Remove dtoverlay from config.txt
if grep -q '^dtoverlay=dwc2$' /boot/firmware/config.txt; then
    echo "Removing USB Gadget support from /boot/firmware/config.txt"
    sed -i '/^dtoverlay=dwc2$/d' /boot/firmware/config.txt
fi

# Remove virtualenv
if [ -d /opt/kindavm/venv ]; then
    echo "Removing virtualenv..."
    rm -rf /opt/kindavm/venv
fi

# Remove /opt/kindavm directory if empty
if [ -d /opt/kindavm ] && [ -z "$(ls -A /opt/kindavm)" ]; then
    echo "Removing empty /opt/kindavm directory..."
    rmdir /opt/kindavm
fi

# Verify HID device is gone
if [ -e /dev/hidg0 ]; then
    echo "Warning: /dev/hidg0 still exists" >&2
else
    echo "HID device removed successfully"
fi

echo ""
echo "Uninstallation complete!"
echo "Note: This does not remove the source directory or any local files."
