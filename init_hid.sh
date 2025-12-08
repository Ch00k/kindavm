#!/usr/bin/env bash

set -e

# Logging function
log() {
    echo "[kindavm-init] $*" >&2
}

log "Starting KindaVM HID gadget initialization"

sleep 3

# ConfigFS paths for USB gadget configuration
USB_GADGET_DIR=/sys/kernel/config/usb_gadget
HID_GADGET_DIR=${USB_GADGET_DIR}/kindavm
HID_STRINGS_DIR=${HID_GADGET_DIR}/strings/0x409
HID_CONFIGS_DIR=${HID_GADGET_DIR}/configs/c.1
HID_CONFIG_STRINGS_DIR=${HID_CONFIGS_DIR}/strings/0x409
HID_FUNCTIONS_DIR=${HID_GADGET_DIR}/functions/hid.usb0
HID_FUNCTIONS_LINK=${HID_CONFIGS_DIR}/hid.usb0
HID_UDC_FILE=${HID_GADGET_DIR}/UDC

# Create directory structure for gadget configuration
log "Creating gadget directory structure"
mkdir -p ${HID_GADGET_DIR}
mkdir -p ${HID_STRINGS_DIR}
mkdir -p ${HID_CONFIG_STRINGS_DIR}
mkdir -p ${HID_FUNCTIONS_DIR}

# Set USB device descriptor values
log "Configuring USB device descriptor"
echo 0x1d6b >${HID_GADGET_DIR}/idVendor  # Linux Foundation
echo 0x0104 >${HID_GADGET_DIR}/idProduct # Multifunction Composite Gadget
echo 0x0110 >${HID_GADGET_DIR}/bcdDevice # v1.1.0
echo 0x0200 >${HID_GADGET_DIR}/bcdUSB    # USB2.0

# Set device strings (shown to host computer)
log "Setting device identification strings"
echo "KVM00000042" >${HID_STRINGS_DIR}/serialnumber
echo "KindaVM" >${HID_STRINGS_DIR}/manufacturer
echo "KindaVM HID" >${HID_STRINGS_DIR}/product

# Configure the USB configuration
log "Configuring USB configuration parameters"
echo "Config 1: HID Keyboard + Mouse" >${HID_CONFIG_STRINGS_DIR}/configuration
echo 0x80 >${HID_CONFIGS_DIR}/bmAttributes # Bus-powered (not self-powered)
echo 250 >${HID_CONFIGS_DIR}/MaxPower      # 250 * 2mA = 500mA max current

# Configure HID function parameters
log "Configuring HID function parameters"
echo 1 >${HID_FUNCTIONS_DIR}/protocol      # Keyboard protocol
echo 1 >${HID_FUNCTIONS_DIR}/subclass      # Boot interface subclass
echo 9 >${HID_FUNCTIONS_DIR}/report_length # Max report size (keyboard = 9 bytes)

# Write HID Report Descriptor
# This descriptor defines the structure of HID reports we'll send to the host.
# See hid_report_desc.md for byte-by-byte documentation.
# Contains 4 report IDs:
#   Report ID 1: Keyboard (modifiers + 6 key rollover)
#   Report ID 2: Consumer Control (media keys, volume, brightness, etc.)
#   Report ID 3: System Control (power, sleep, wake)
#   Report ID 4: Mouse (buttons, X, Y, wheel)
log "Writing HID report descriptor"
cat /usr/local/lib/kindavm/hid_report_desc.bin >${HID_FUNCTIONS_DIR}/report_desc

# Link the HID function to the configuration
# This associates the HID function with configuration 1
log "Linking HID function to configuration"
ln -s ${HID_FUNCTIONS_DIR} ${HID_FUNCTIONS_LINK}

# Find and bind to a USB Device Controller (UDC)
# The UDC is the hardware that implements USB device functionality
log "Finding available USB Device Controller"
# shellcheck disable=SC2012
UDC=$(ls /sys/class/udc 2>/dev/null | head -n1)
if [ -z "$UDC" ]; then
    log "ERROR: No UDC (USB Device Controller) found"
    log "Check that your device supports USB gadget mode"
    log "Available UDCs: $(ls /sys/class/udc 2>&1)"
    exit 1
fi

log "Binding gadget to UDC: $UDC"
echo "$UDC" >${HID_UDC_FILE}

# Verify the HID device was created
if [ -e /dev/hidg0 ]; then
    log "SUCCESS: HID device created at /dev/hidg0"
else
    log "WARNING: /dev/hidg0 not found after binding to UDC"
fi

log "KindaVM HID gadget initialization complete"
