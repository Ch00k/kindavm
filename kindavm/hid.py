"""Core HID report sending functionality."""

import time
from pathlib import Path

HID_DEVICE = "/dev/hidg0"


def send_report(report: bytes, delay_ms: float = 10) -> None:
    """Send a HID report to the device.

    Args:
        report: The HID report bytes to send
        delay_ms: Delay in milliseconds after sending the report
    """
    with open(HID_DEVICE, "rb+") as hid:
        hid.write(report)

    if delay_ms > 0:
        time.sleep(delay_ms / 1000.0)


def check_device() -> bool:
    """Check if the HID device is available.

    Returns:
        True if device exists and is writable, False otherwise
    """
    device = Path(HID_DEVICE)
    return device.exists() and device.is_char_device()
