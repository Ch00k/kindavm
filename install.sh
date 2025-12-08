#!/usr/bin/env bash
set -e

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root (use sudo)" >&2
    exit 1
fi

# Detect installation mode
UPGRADE_MODE=false
CURRENT_VERSION="unknown"
INSTALL_DIR="/tmp/kindavm-install"

if [ -x "/usr/local/bin/kindavmd" ]; then
    UPGRADE_MODE=true
    CURRENT_VERSION=$(/usr/local/bin/kindavmd --version 2>/dev/null || echo "unknown")
fi

# Get latest release information
echo "Fetching latest release information..."
RELEASE_INFO=$(curl -sSL https://api.github.com/repos/Ch00k/kindavm/releases/latest)
LATEST_VERSION=$(echo "$RELEASE_INFO" | grep -o '"tag_name": "[^"]*"' | cut -d'"' -f4)

if [ -n "$LATEST_VERSION" ]; then
    echo "Latest release: $LATEST_VERSION"
else
    echo "Error: Could not fetch latest version"
    exit 1
fi

# Show installation mode
echo
if [ "$UPGRADE_MODE" = true ]; then
    echo "Existing kindavm installation detected (version: $CURRENT_VERSION)"

    # Check if versions are the same
    if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ]; then
        echo
        echo "You already have the latest version ($CURRENT_VERSION) installed."
        echo "No upgrade needed. Exiting."
        exit 0
    else
        echo "Upgrading kindavm to version: $LATEST_VERSION"
    fi
else
    echo "Installing kindavm version: $LATEST_VERSION"
fi

echo

# Get download URL
ARCHIVE_URL=$(echo "$RELEASE_INFO" | grep -o '"browser_download_url": "[^"]*kindavm-linux-arm64.tar.gz"' | cut -d'"' -f4)

if [ -z "$ARCHIVE_URL" ]; then
    echo "Error: Could not find release archive in latest release"
    exit 1
fi

# Stop services if upgrading
if [ "$UPGRADE_MODE" = true ]; then
    echo "Stopping services..."
    systemctl stop kindavmd.service || true
    systemctl stop kindavm-init.service || true
fi

# Download and extract release archive
echo "Downloading release archive..."
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"
curl -sSL "$ARCHIVE_URL" -o kindavm.tar.gz

echo "Extracting archive..."
tar -xzf kindavm.tar.gz
rm kindavm.tar.gz

# Install dependencies
echo "Installing dependencies..."
apt-get update
apt-get install --no-install-recommends --no-install-suggests --yes v4l-utils

echo "Installing KindaVM..."

# Track if we need to reboot
NEEDS_REBOOT=false

# Install boot configuration
echo "Installing boot configuration..."
if ! cmp -s /boot/firmware/config.txt configs/config.txt; then
    echo "Boot configuration has changed, backup and update..."
    cp /boot/firmware/config.txt /boot/firmware/config.txt.backup
    cp configs/config.txt /boot/firmware/config.txt
    NEEDS_REBOOT=true
else
    echo "Boot configuration is already up to date"
fi

# Configure automatic module loading
if [ -f /etc/modules-load.d/kindavm.conf ]; then
    echo "Module loading configuration already exists at /etc/modules-load.d/kindavm.conf"
else
    echo "Creating module loading configuration at /etc/modules-load.d/kindavm.conf"
    echo -en 'dwc2\nlibcomposite\n' | tee /etc/modules-load.d/kindavm.conf
    NEEDS_REBOOT=true
fi

# Install kindavmd binary
echo "Installing kindavmd binary..."
cp kindavmd /usr/local/bin/kindavmd
chmod +x /usr/local/bin/kindavmd

# Install edidmod tool
echo "Installing edidmod tool..."
cp edidmod /usr/local/bin/edidmod
chmod +x /usr/local/bin/edidmod

# Install ustreamer tool
echo "Installing ustreamer tool..."
cp ustreamer /usr/local/bin/ustreamer
chmod +x /usr/local/bin/ustreamer

# Install init and uninstall scripts
echo "Installing HID initialization script..."
cp init_hid.sh /usr/local/bin/kindavm-init-hid.sh
chmod +x /usr/local/bin/kindavm-init-hid.sh

echo "Installing HDMI initialization script..."
cp init_hdmi.sh /usr/local/bin/kindavm-init-hdmi.sh
chmod +x /usr/local/bin/kindavm-init-hdmi.sh

echo "Installing HDMI EDID file..."
mkdir -p /usr/local/lib/kindavm
cp configs/edid.hex /usr/local/lib/kindavm/edid.hex

echo "Installing uninstall script..."
cp uninstall.sh /usr/local/bin/kindavm-uninstall.sh
chmod +x /usr/local/bin/kindavm-uninstall.sh

# Install systemd services
echo "Installing systemd services..."
cp kindavm-init.service /etc/systemd/system/
cp kindavmd.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable kindavm-init.service
systemctl enable kindavmd.service

# Cleanup
cd /
rm -rf "$INSTALL_DIR"

echo ""
echo "Installation complete!"
echo ""

# Show reboot message if needed
if [ "$NEEDS_REBOOT" = true ]; then
    echo "IMPORTANT: A reboot is required to enable USB gadget support."
    echo "After rebooting, services will start automatically."
    echo ""
    echo "To reboot now: sudo reboot"
    echo ""
else
    echo "USB gadget support is already enabled."
    echo "Services will start on next boot, or you can start them now:"
    echo "  sudo systemctl start kindavm-init.service"
    echo "  sudo systemctl start kindavmd.service"
    echo ""
fi

echo "To uninstall: sudo kindavm-uninstall.sh"

# Show completion message
echo
if [ "$UPGRADE_MODE" = true ]; then
    echo "kindavm upgraded: $CURRENT_VERSION → $LATEST_VERSION"
else
    echo "kindavm $LATEST_VERSION installed successfully"
fi
