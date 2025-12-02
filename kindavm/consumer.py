"""Consumer Control HID functionality (media keys, volume, brightness)."""

from .hid import send_report

# Consumer control button bits (Report ID 2, Byte 1)
VOLUME_UP = 0x01
VOLUME_DOWN = 0x02
MUTE = 0x04
PLAY_PAUSE = 0x08
NEXT_TRACK = 0x10
PREV_TRACK = 0x20
STOP = 0x40
MAIL = 0x80

# Consumer control button bits (Report ID 2, Byte 2)
BRIGHTNESS_UP = 0x01
BRIGHTNESS_DOWN = 0x02
AC_SEARCH = 0x04
AC_HOME = 0x08
AC_BACK = 0x10
AC_FORWARD = 0x20
AC_STOP = 0x40
AC_REFRESH = 0x80

# Consumer control button bits (Report ID 2, Byte 3)
EJECT = 0x01


def send_consumer_key(byte1: int = 0, byte2: int = 0, byte3: int = 0) -> None:
    """Send a consumer control key press and release.

    Args:
        byte1: First byte of consumer control bits (volume, media controls)
        byte2: Second byte of consumer control bits (brightness, browser)
        byte3: Third byte of consumer control bits (eject)
    """
    # Press (Report ID 2 + 3 bytes = 4 bytes total)
    report = bytes([0x02, byte1, byte2, byte3])
    send_report(report, delay_ms=10)

    # Release (Report ID 2 + 3 bytes = 4 bytes total)
    report = bytes([0x02, 0x00, 0x00, 0x00])
    send_report(report, delay_ms=10)


def volume_up() -> None:
    """Increase volume."""
    send_consumer_key(byte1=VOLUME_UP)


def volume_down() -> None:
    """Decrease volume."""
    send_consumer_key(byte1=VOLUME_DOWN)


def mute() -> None:
    """Toggle mute."""
    send_consumer_key(byte1=MUTE)


def play_pause() -> None:
    """Toggle play/pause."""
    send_consumer_key(byte1=PLAY_PAUSE)


def next_track() -> None:
    """Skip to next track."""
    send_consumer_key(byte1=NEXT_TRACK)


def prev_track() -> None:
    """Skip to previous track."""
    send_consumer_key(byte1=PREV_TRACK)


def stop() -> None:
    """Stop playback."""
    send_consumer_key(byte1=STOP)


def brightness_up() -> None:
    """Increase brightness."""
    send_consumer_key(byte2=BRIGHTNESS_UP)


def brightness_down() -> None:
    """Decrease brightness."""
    send_consumer_key(byte2=BRIGHTNESS_DOWN)


def eject() -> None:
    """Eject media."""
    send_consumer_key(byte3=EJECT)
