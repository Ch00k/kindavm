"""Mouse HID functionality."""

from typing import Literal

from .hid import send_report

# Mouse button bits
BUTTON_NONE = 0x00
BUTTON_LEFT = 0x01
BUTTON_RIGHT = 0x02
BUTTON_MIDDLE = 0x04

ButtonType = Literal["left", "right", "middle"]


def _clamp_movement(value: int) -> int:
    """Clamp movement value to valid range (-127 to 127).

    Args:
        value: The movement value to clamp

    Returns:
        Clamped value in range -127 to 127
    """
    return max(-127, min(127, value))


def _button_to_bits(button: ButtonType) -> int:
    """Convert button name to button bits.

    Args:
        button: Button name ("left", "right", or "middle")

    Returns:
        Button bit value

    Raises:
        ValueError: If button name is invalid
    """
    button_map = {
        "left": BUTTON_LEFT,
        "right": BUTTON_RIGHT,
        "middle": BUTTON_MIDDLE,
    }
    if button not in button_map:
        raise ValueError(f"Invalid button: {button}. Must be 'left', 'right', or 'middle'")
    return button_map[button]


def send_mouse_report(buttons: int = 0, x: int = 0, y: int = 0, wheel: int = 0) -> None:
    """Send a mouse HID report.

    Args:
        buttons: Button bits (BUTTON_LEFT, BUTTON_RIGHT, BUTTON_MIDDLE)
        x: X movement (-127 to 127)
        y: Y movement (-127 to 127)
        wheel: Wheel movement (-127 to 127)
    """
    x = _clamp_movement(x)
    y = _clamp_movement(y)
    wheel = _clamp_movement(wheel)

    # Convert signed values to unsigned bytes
    x_byte = x & 0xFF
    y_byte = y & 0xFF
    wheel_byte = wheel & 0xFF

    # Report ID 4 + buttons + x + y + wheel = 5 bytes total
    report = bytes([0x04, buttons, x_byte, y_byte, wheel_byte])
    send_report(report, delay_ms=10)


def move(x: int, y: int) -> None:
    """Move the mouse cursor.

    Args:
        x: Horizontal movement (-127 to 127, positive is right)
        y: Vertical movement (-127 to 127, positive is down)
    """
    send_mouse_report(buttons=0, x=x, y=y, wheel=0)


def click(button: ButtonType = "left", count: int = 1) -> None:
    """Click a mouse button.

    Args:
        button: Button to click ("left", "right", or "middle")
        count: Number of times to click (default: 1)
    """
    button_bits = _button_to_bits(button)

    for _ in range(count):
        # Press
        send_mouse_report(buttons=button_bits, x=0, y=0, wheel=0)
        # Release
        send_mouse_report(buttons=0, x=0, y=0, wheel=0)


def scroll(amount: int) -> None:
    """Scroll the mouse wheel.

    Args:
        amount: Scroll amount (-127 to 127, positive is down)
    """
    send_mouse_report(buttons=0, x=0, y=0, wheel=amount)


def drag(x: int, y: int, button: ButtonType = "left") -> None:
    """Drag the mouse with button held down.

    Args:
        x: Horizontal movement (-127 to 127)
        y: Vertical movement (-127 to 127)
        button: Button to hold ("left", "right", or "middle")
    """
    button_bits = _button_to_bits(button)

    # Press button
    send_mouse_report(buttons=button_bits, x=0, y=0, wheel=0)

    # Move with button held
    # For large movements, break into chunks
    steps = max(abs(x), abs(y))
    if steps == 0:
        steps = 1

    x_step = x / steps
    y_step = y / steps

    for i in range(steps):
        step_x = int((i + 1) * x_step) - int(i * x_step)
        step_y = int((i + 1) * y_step) - int(i * y_step)
        send_mouse_report(buttons=button_bits, x=step_x, y=step_y, wheel=0)

    # Release button
    send_mouse_report(buttons=0, x=0, y=0, wheel=0)
