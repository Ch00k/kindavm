"""System Control HID functionality (power, sleep, wake)."""

from .hid import send_report

# System control button bits (Report ID 3, Byte 1)
POWER = 0x01
SLEEP = 0x02
WAKE = 0x04


def send_system_key(buttons: int) -> None:
    """Send a system control key press and release.

    Args:
        buttons: Bit mask of system control buttons
    """
    # Press (Report ID 3 + 1 byte = 2 bytes total)
    report = bytes([0x03, buttons])
    send_report(report, delay_ms=10)

    # Release (Report ID 3 + 1 byte = 2 bytes total)
    report = bytes([0x03, 0x00])
    send_report(report, delay_ms=10)


def power() -> None:
    """Send power button."""
    send_system_key(POWER)


def sleep() -> None:
    """Send sleep button."""
    send_system_key(SLEEP)


def wake() -> None:
    """Send wake button."""
    send_system_key(WAKE)
