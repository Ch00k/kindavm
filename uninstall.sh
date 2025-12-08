#!/usr/bin/env bash

set -e

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root (use sudo)" >&2
    exit 1
fi

echo "Uninstalling KindaVM..."

# Stop and disable systemd services
if systemctl is-active --quiet kindavmd.service; then
    echo "Stopping kindavmd service..."
    systemctl stop kindavmd.service
fi

if systemctl is-active --quiet kindavm-init.service; then
    echo "Stopping kindavm-init service..."
    systemctl stop kindavm-init.service
fi

if systemctl is-enabled --quiet kindavmd.service 2>/dev/null; then
    echo "Disabling kindavmd service..."
    systemctl disable kindavmd.service
fi

if systemctl is-enabled --quiet kindavm-init.service 2>/dev/null; then
    echo "Disabling kindavm-init service..."
    systemctl disable kindavm-init.service
fi

# Remove systemd service files
if [ -f /etc/systemd/system/kindavmd.service ]; then
    echo "Removing kindavmd service file..."
    rm -f /etc/systemd/system/kindavmd.service
fi

if [ -f /etc/systemd/system/kindavm-init.service ]; then
    echo "Removing kindavm-init service file..."
    rm -f /etc/systemd/system/kindavm-init.service
fi

systemctl daemon-reload

# Unbind and remove USB gadget
if [ -f /sys/kernel/config/usb_gadget/kindavm/UDC ]; then
    echo "Unbinding USB gadget..."
    echo "" > /sys/kernel/config/usb_gadget/kindavm/UDC 2>/dev/null || true
fi

if [ -d /sys/kernel/config/usb_gadget/kindavm ]; then
    echo "Removing USB gadget configuration..."
    rm -rf /sys/kernel/config/usb_gadget/kindavm
fi

# Remove kindavmd binary
if [ -f /usr/local/bin/kindavmd ]; then
    echo "Removing kindavmd binary..."
    rm -f /usr/local/bin/kindavmd
fi

# Remove edidmod tool
if [ -f /usr/local/bin/edidmod ]; then
    echo "Removing edidmod tool..."
    rm -f /usr/local/bin/edidmod
fi

# Remove ustreamer tool
if [ -f /usr/local/bin/ustreamer ]; then
    echo "Removing ustreamer tool..."
    rm -f /usr/local/bin/ustreamer
fi

# Remove init scripts
if [ -f /usr/local/bin/kindavm-init-hid.sh ]; then
    echo "Removing HID initialization script..."
    rm -f /usr/local/bin/kindavm-init-hid.sh
fi

if [ -f /usr/local/bin/kindavm-init-hdmi.sh ]; then
    echo "Removing HDMI initialization script..."
    rm -f /usr/local/bin/kindavm-init-hdmi.sh
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

# Remove EDID file and kindavm directory
if [ -f /usr/local/lib/kindavm/edid.hex ]; then
    echo "Removing EDID file..."
    rm -f /usr/local/lib/kindavm/edid.hex
fi

if [ -d /usr/local/lib/kindavm ] && [ -z "$(ls -A /usr/local/lib/kindavm)" ]; then
    echo "Removing empty /usr/local/lib/kindavm directory..."
    rmdir /usr/local/lib/kindavm
fi

# Restore boot configuration backup
if [ -f /boot/firmware/config.txt.backup ]; then
    echo "Restoring boot configuration from backup..."
    cp /boot/firmware/config.txt.backup /boot/firmware/config.txt
    rm -f /boot/firmware/config.txt.backup
fi

# Verify HID device is gone
if [ -e /dev/hidg0 ]; then
    echo "Warning: /dev/hidg0 still exists" >&2
else
    echo "HID device removed successfully"
fi

echo ""
echo "Uninstallation complete!"
echo "Note: A reboot may be required to fully remove kernel modules."
