"""Keyboard HID functionality."""

import sys

from .hid import send_report

# HID modifier bits
MOD_NONE = 0x00
MOD_LEFT_CTRL = 0x01
MOD_LEFT_SHIFT = 0x02
MOD_LEFT_ALT = 0x04
MOD_LEFT_META = 0x08
MOD_RIGHT_CTRL = 0x10
MOD_RIGHT_SHIFT = 0x20
MOD_RIGHT_ALT = 0x40
MOD_RIGHT_META = 0x80

# HID usage codes for keys
# Format: (keycode, needs_shift)
KEY_MAP = {
    "a": (0x04, False),
    "b": (0x05, False),
    "c": (0x06, False),
    "d": (0x07, False),
    "e": (0x08, False),
    "f": (0x09, False),
    "g": (0x0A, False),
    "h": (0x0B, False),
    "i": (0x0C, False),
    "j": (0x0D, False),
    "k": (0x0E, False),
    "l": (0x0F, False),
    "m": (0x10, False),
    "n": (0x11, False),
    "o": (0x12, False),
    "p": (0x13, False),
    "q": (0x14, False),
    "r": (0x15, False),
    "s": (0x16, False),
    "t": (0x17, False),
    "u": (0x18, False),
    "v": (0x19, False),
    "w": (0x1A, False),
    "x": (0x1B, False),
    "y": (0x1C, False),
    "z": (0x1D, False),
    "A": (0x04, True),
    "B": (0x05, True),
    "C": (0x06, True),
    "D": (0x07, True),
    "E": (0x08, True),
    "F": (0x09, True),
    "G": (0x0A, True),
    "H": (0x0B, True),
    "I": (0x0C, True),
    "J": (0x0D, True),
    "K": (0x0E, True),
    "L": (0x0F, True),
    "M": (0x10, True),
    "N": (0x11, True),
    "O": (0x12, True),
    "P": (0x13, True),
    "Q": (0x14, True),
    "R": (0x15, True),
    "S": (0x16, True),
    "T": (0x17, True),
    "U": (0x18, True),
    "V": (0x19, True),
    "W": (0x1A, True),
    "X": (0x1B, True),
    "Y": (0x1C, True),
    "Z": (0x1D, True),
    "1": (0x1E, False),
    "2": (0x1F, False),
    "3": (0x20, False),
    "4": (0x21, False),
    "5": (0x22, False),
    "6": (0x23, False),
    "7": (0x24, False),
    "8": (0x25, False),
    "9": (0x26, False),
    "0": (0x27, False),
    "!": (0x1E, True),
    "@": (0x1F, True),
    "#": (0x20, True),
    "$": (0x21, True),
    "%": (0x22, True),
    "^": (0x23, True),
    "&": (0x24, True),
    "*": (0x25, True),
    "(": (0x26, True),
    ")": (0x27, True),
    "\n": (0x28, False),  # Enter
    "\b": (0x2A, False),  # Backspace
    "\t": (0x2B, False),  # Tab
    " ": (0x2C, False),  # Space
    "-": (0x2D, False),
    "_": (0x2D, True),
    "=": (0x2E, False),
    "+": (0x2E, True),
    "[": (0x2F, False),
    "{": (0x2F, True),
    "]": (0x30, False),
    "}": (0x30, True),
    "\\": (0x31, False),
    "|": (0x31, True),
    ";": (0x33, False),
    ":": (0x33, True),
    "'": (0x34, False),
    '"': (0x34, True),
    "`": (0x35, False),
    "~": (0x35, True),
    ",": (0x36, False),
    "<": (0x36, True),
    ".": (0x37, False),
    ">": (0x37, True),
    "/": (0x38, False),
    "?": (0x38, True),
}

# Special key codes (not characters, used with send_key directly)
KEY_ESCAPE = 0x29
KEY_F1 = 0x3A
KEY_F2 = 0x3B
KEY_F3 = 0x3C
KEY_F4 = 0x3D
KEY_F5 = 0x3E
KEY_F6 = 0x3F
KEY_F7 = 0x40
KEY_F8 = 0x41
KEY_F9 = 0x42
KEY_F10 = 0x43
KEY_F11 = 0x44
KEY_F12 = 0x45
KEY_PRINT_SCREEN = 0x46
KEY_SCROLL_LOCK = 0x47
KEY_PAUSE = 0x48
KEY_INSERT = 0x49
KEY_HOME = 0x4A
KEY_PAGE_UP = 0x4B
KEY_DELETE = 0x4C
KEY_END = 0x4D
KEY_PAGE_DOWN = 0x4E
KEY_RIGHT_ARROW = 0x4F
KEY_LEFT_ARROW = 0x50
KEY_DOWN_ARROW = 0x51
KEY_UP_ARROW = 0x52


def send_key(keycode: int, modifier: int = MOD_NONE) -> None:
    """Send a key press and release.

    Args:
        keycode: HID keycode to send
        modifier: Modifier keys to apply (default: MOD_NONE)
    """
    # Press (Report ID 1 + 8 bytes = 9 bytes total)
    report = bytes([0x01, modifier, 0x00, keycode, 0x00, 0x00, 0x00, 0x00, 0x00])
    send_report(report, delay_ms=10)

    # Release (Report ID 1 + 8 bytes = 9 bytes total)
    report = bytes([0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00])
    send_report(report, delay_ms=10)


def type_string(text: str) -> None:
    """Type a string by sending HID reports for each character.

    Args:
        text: The text to type (escape sequences like \\n are interpreted)
    """
    # Process escape sequences
    text = text.replace("\\n", "\n").replace("\\t", "\t").replace("\\b", "\b")

    for char in text:
        if char not in KEY_MAP:
            print(f"Warning: Character '{char}' not in keymap, skipping", file=sys.stderr)
            continue

        keycode, needs_shift = KEY_MAP[char]
        modifier = MOD_LEFT_SHIFT if needs_shift else MOD_NONE
        send_key(keycode, modifier)
