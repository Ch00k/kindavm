#!/usr/bin/env bash

set -e

# Logging function
log() {
    echo "[kindavm-init] $*" >&2
}

log "Starting KindaVM HID gadget initialization"

# ConfigFS paths for USB gadget configuration
USB_GADGET_DIR=/sys/kernel/config/usb_gadget
HID_GADGET_DIR=${USB_GADGET_DIR}/kindavm
HID_STRINGS_DIR=${HID_GADGET_DIR}/strings/0x409
HID_CONFIGS_DIR=${HID_GADGET_DIR}/configs/c.1
HID_CONFIG_STRINGS_DIR=${HID_CONFIGS_DIR}/strings/0x409
HID_FUNCTIONS_DIR=${HID_GADGET_DIR}/functions/hid.usb0
HID_FUNCTIONS_LINK=${HID_CONFIGS_DIR}/hid.usb0
HID_UDC_FILE=${HID_GADGET_DIR}/UDC

# Unbind gadget from UDC if already bound
# This allows us to reconfigure the gadget
if [ -f ${HID_UDC_FILE} ]; then
    log "Unbinding existing gadget from UDC"
    echo "" >${HID_UDC_FILE} 2>/dev/null || true
fi

# Remove existing gadget configuration to ensure clean state
if [ -d ${HID_GADGET_DIR} ]; then
    log "Removing existing gadget configuration"
    rm -rf ${HID_GADGET_DIR}
fi

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
/usr/bin/printf '%b' \
    '\x05\x01' \
    '\x09\x06' \
    '\xa1\x01' \
    '\x85\x01' \
    '\x05\x07' \
    '\x19\xe0' \
    '\x29\xe7' \
    '\x15\x00' \
    '\x25\x01' \
    '\x75\x01' \
    '\x95\x08' \
    '\x81\x02' \
    '\x95\x01' \
    '\x75\x08' \
    '\x81\x03' \
    '\x95\x05' \
    '\x75\x01' \
    '\x05\x08' \
    '\x19\x01' \
    '\x29\x05' \
    '\x91\x02' \
    '\x95\x01' \
    '\x75\x03' \
    '\x91\x03' \
    '\x95\x06' \
    '\x75\x08' \
    '\x15\x00' \
    '\x25\x65' \
    '\x05\x07' \
    '\x19\x00' \
    '\x29\x65' \
    '\x81\x00' \
    '\xc0' \
    '\x05\x0c' \
    '\x09\x01' \
    '\xa1\x01' \
    '\x85\x02' \
    '\x15\x00' \
    '\x25\x01' \
    '\x75\x01' \
    '\x95\x08' \
    '\x09\xe9' \
    '\x09\xea' \
    '\x09\xe2' \
    '\x09\xcd' \
    '\x09\xb5' \
    '\x09\xb6' \
    '\x09\xb7' \
    '\x0a\x8a\x01' \
    '\x81\x02' \
    '\x95\x08' \
    '\x09\x6f' \
    '\x09\x70' \
    '\x0a\x21\x02' \
    '\x0a\x23\x02' \
    '\x0a\x24\x02' \
    '\x0a\x25\x02' \
    '\x0a\x26\x02' \
    '\x0a\x27\x02' \
    '\x81\x02' \
    '\x95\x01' \
    '\x09\xb8' \
    '\x81\x02' \
    '\x95\x07' \
    '\x81\x03' \
    '\xc0' \
    '\x05\x01' \
    '\x09\x80' \
    '\xa1\x01' \
    '\x85\x03' \
    '\x15\x00' \
    '\x25\x01' \
    '\x75\x01' \
    '\x95\x03' \
    '\x09\x81' \
    '\x09\x82' \
    '\x09\x83' \
    '\x81\x02' \
    '\x95\x05' \
    '\x81\x03' \
    '\xc0' \
    '\x05\x01' \
    '\x09\x02' \
    '\xa1\x01' \
    '\x85\x04' \
    '\x09\x01' \
    '\xa1\x00' \
    '\x05\x09' \
    '\x19\x01' \
    '\x29\x03' \
    '\x15\x00' \
    '\x25\x01' \
    '\x95\x03' \
    '\x75\x01' \
    '\x81\x02' \
    '\x95\x01' \
    '\x75\x05' \
    '\x81\x03' \
    '\x05\x01' \
    '\x09\x30' \
    '\x09\x31' \
    '\x09\x38' \
    '\x15\x81' \
    '\x25\x7f' \
    '\x75\x08' \
    '\x95\x03' \
    '\x81\x06' \
    '\xc0' \
    '\xc0' \
    >${HID_FUNCTIONS_DIR}/report_desc

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
