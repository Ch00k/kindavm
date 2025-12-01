"""Command-line interface for KindaVM."""

import sys
from typing import List, cast

from . import keyboard, mouse
from .hid import check_device
from .mouse import ButtonType


def print_help() -> None:
    """Print usage information."""
    print("KindaVM - USB HID Keyboard and Mouse Emulation", file=sys.stderr)
    print("", file=sys.stderr)
    print("Usage:", file=sys.stderr)
    print("  kinda type <text>              Type text string", file=sys.stderr)
    print(
        "  kinda key <keycode> [modifier] Send specific key with optional modifier",
        file=sys.stderr,
    )
    print("  kinda mouse move <x> <y>       Move mouse cursor", file=sys.stderr)
    print(
        "  kinda mouse click [button]     Click mouse button (left/right/middle)",
        file=sys.stderr,
    )
    print("  kinda mouse scroll <amount>    Scroll mouse wheel", file=sys.stderr)
    print("  kinda mouse drag <x> <y> [button] Drag with button held", file=sys.stderr)
    print("", file=sys.stderr)
    print("Examples:", file=sys.stderr)
    print("  kinda type 'Hello World'", file=sys.stderr)
    print("  echo 'test' | kinda type", file=sys.stderr)
    print("  kinda mouse move 10 20", file=sys.stderr)
    print("  kinda mouse click right", file=sys.stderr)
    print("  kinda mouse scroll -5", file=sys.stderr)


def cmd_type(args: List[str]) -> int:
    """Handle 'type' command."""
    if not args:
        # Check if reading from pipe/stdin
        if not sys.stdin.isatty():
            text = sys.stdin.read()
        else:
            print("Error: No text provided", file=sys.stderr)
            print("Usage: kinda type <text>", file=sys.stderr)
            print("   or: echo 'text' | kinda type", file=sys.stderr)
            return 1
    else:
        text = " ".join(args)

    keyboard.type_string(text)
    return 0


def cmd_key(args: List[str]) -> int:
    """Handle 'key' command."""
    if not args:
        print("Error: No keycode provided", file=sys.stderr)
        print("Usage: kinda key <keycode> [modifier]", file=sys.stderr)
        return 1

    try:
        keycode = int(args[0], 0)  # Support hex (0x..) and decimal
    except ValueError:
        print(f"Error: Invalid keycode: {args[0]}", file=sys.stderr)
        return 1

    modifier = keyboard.MOD_NONE
    if len(args) > 1:
        try:
            modifier = int(args[1], 0)
        except ValueError:
            print(f"Error: Invalid modifier: {args[1]}", file=sys.stderr)
            return 1

    keyboard.send_key(keycode, modifier)
    return 0


def cmd_mouse(args: List[str]) -> int:
    """Handle 'mouse' subcommands."""
    if not args:
        print("Error: No mouse command provided", file=sys.stderr)
        print("Usage: kinda mouse <move|click|scroll|drag> [args...]", file=sys.stderr)
        return 1

    subcmd = args[0]
    subargs = args[1:]

    if subcmd == "move":
        if len(subargs) < 2:
            print("Error: move requires x and y arguments", file=sys.stderr)
            print("Usage: kinda mouse move <x> <y>", file=sys.stderr)
            return 1
        try:
            x = int(subargs[0])
            y = int(subargs[1])
        except ValueError:
            print("Error: x and y must be integers", file=sys.stderr)
            return 1
        mouse.move(x, y)
        return 0

    elif subcmd == "click":
        button_str = subargs[0] if subargs else "left"
        if button_str not in ["left", "right", "middle"]:
            print(f"Error: Invalid button: {button_str}", file=sys.stderr)
            print("Must be 'left', 'right', or 'middle'", file=sys.stderr)
            return 1
        button = cast(ButtonType, button_str)
        count = 1
        if len(subargs) > 1:
            try:
                count = int(subargs[1])
            except ValueError:
                print("Error: count must be an integer", file=sys.stderr)
                return 1
        mouse.click(button, count)
        return 0

    elif subcmd == "scroll":
        if not subargs:
            print("Error: scroll requires amount argument", file=sys.stderr)
            print("Usage: kinda mouse scroll <amount>", file=sys.stderr)
            return 1
        try:
            amount = int(subargs[0])
        except ValueError:
            print("Error: amount must be an integer", file=sys.stderr)
            return 1
        mouse.scroll(amount)
        return 0

    elif subcmd == "drag":
        if len(subargs) < 2:
            print("Error: drag requires x and y arguments", file=sys.stderr)
            print("Usage: kinda mouse drag <x> <y> [button]", file=sys.stderr)
            return 1
        try:
            x = int(subargs[0])
            y = int(subargs[1])
        except ValueError:
            print("Error: x and y must be integers", file=sys.stderr)
            return 1
        button_str = subargs[2] if len(subargs) > 2 else "left"
        if button_str not in ["left", "right", "middle"]:
            print(f"Error: Invalid button: {button_str}", file=sys.stderr)
            print("Must be 'left', 'right', or 'middle'", file=sys.stderr)
            return 1
        button = cast(ButtonType, button_str)
        mouse.drag(x, y, button)
        return 0

    else:
        print(f"Error: Unknown mouse command: {subcmd}", file=sys.stderr)
        print("Usage: kinda mouse <move|click|scroll|drag> [args...]", file=sys.stderr)
        return 1


def main() -> int:
    """Main entry point for CLI."""
    if len(sys.argv) < 2:
        print_help()
        return 1

    # Check if HID device is available
    if not check_device():
        print("Error: HID device not found at /dev/hidg0", file=sys.stderr)
        print("Run 'sudo ./init_hid.sh' to initialize the USB gadget", file=sys.stderr)
        return 1

    command = sys.argv[1]
    args = sys.argv[2:]

    if command == "type":
        return cmd_type(args)
    elif command == "key":
        return cmd_key(args)
    elif command == "mouse":
        return cmd_mouse(args)
    elif command in ["help", "-h", "--help"]:
        print_help()
        return 0
    else:
        print(f"Error: Unknown command: {command}", file=sys.stderr)
        print_help()
        return 1


if __name__ == "__main__":
    sys.exit(main())
