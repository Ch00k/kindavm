"""Command-line interface for KindaVM."""

import sys
from typing import List, cast

from . import consumer, keyboard, mouse, system
from .hid import check_device
from .mouse import ButtonType


def print_help() -> None:
    """Print usage information."""
    print("KindaVM - USB HID Keyboard and Mouse Emulation", file=sys.stderr)
    print("", file=sys.stderr)
    print("Usage:", file=sys.stderr)
    print("  kinda type <text>                      Type text string", file=sys.stderr)
    print("  kinda special-key <key>                Send special key", file=sys.stderr)
    print("  kinda raw-key <keycode> [modifier]     Send raw HID keycode", file=sys.stderr)
    print("  kinda mouse move <x> <y>               Move mouse cursor", file=sys.stderr)
    print("  kinda mouse click [button]             Click mouse button", file=sys.stderr)
    print("  kinda mouse scroll <amount>            Scroll mouse wheel", file=sys.stderr)
    print("  kinda mouse drag <x> <y> [button]      Drag with button held", file=sys.stderr)
    print("", file=sys.stderr)
    print("Special keys:", file=sys.stderr)
    print("  Navigation: f1-f12, esc, home, end, pageup, pagedown, insert, delete,", file=sys.stderr)
    print("              up, down, left, right, printscreen, scrolllock, pause", file=sys.stderr)
    print("  Media:      play, next, prev, stop", file=sys.stderr)
    print("  Volume:     volume-up, volume-down, mute", file=sys.stderr)
    print("  Brightness: brightness-up, brightness-down", file=sys.stderr)
    print("  Power:      power, sleep, wake", file=sys.stderr)
    print("", file=sys.stderr)
    print("Examples:", file=sys.stderr)
    print("  kinda type 'Hello World'", file=sys.stderr)
    print("  echo 'test' | kinda type", file=sys.stderr)
    print("  kinda special-key f1", file=sys.stderr)
    print("  kinda special-key play", file=sys.stderr)
    print("  kinda special-key volume-up", file=sys.stderr)
    print("  kinda raw-key 0x04 0x02", file=sys.stderr)
    print("  kinda mouse move 10 20", file=sys.stderr)
    print("  kinda mouse click right", file=sys.stderr)


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


def cmd_raw_key(args: List[str]) -> int:
    """Handle 'raw-key' command."""
    if not args:
        print("Error: No keycode provided", file=sys.stderr)
        print("Usage: kinda raw-key <keycode> [modifier]", file=sys.stderr)
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


def cmd_special_key(args: List[str]) -> int:
    """Handle 'special-key' command for all special keys."""
    if not args:
        print("Error: No key specified", file=sys.stderr)
        print("Usage: kinda special-key <key>", file=sys.stderr)
        print("See 'kinda help' for available keys", file=sys.stderr)
        return 1

    key = args[0].lower()

    # Map of all special keys to their functions
    special_key_map = {
        # Navigation keys (send keyboard keycodes)
        "esc": lambda: keyboard.send_key(keyboard.KEY_ESCAPE),
        "f1": lambda: keyboard.send_key(keyboard.KEY_F1),
        "f2": lambda: keyboard.send_key(keyboard.KEY_F2),
        "f3": lambda: keyboard.send_key(keyboard.KEY_F3),
        "f4": lambda: keyboard.send_key(keyboard.KEY_F4),
        "f5": lambda: keyboard.send_key(keyboard.KEY_F5),
        "f6": lambda: keyboard.send_key(keyboard.KEY_F6),
        "f7": lambda: keyboard.send_key(keyboard.KEY_F7),
        "f8": lambda: keyboard.send_key(keyboard.KEY_F8),
        "f9": lambda: keyboard.send_key(keyboard.KEY_F9),
        "f10": lambda: keyboard.send_key(keyboard.KEY_F10),
        "f11": lambda: keyboard.send_key(keyboard.KEY_F11),
        "f12": lambda: keyboard.send_key(keyboard.KEY_F12),
        "printscreen": lambda: keyboard.send_key(keyboard.KEY_PRINT_SCREEN),
        "scrolllock": lambda: keyboard.send_key(keyboard.KEY_SCROLL_LOCK),
        "pause": lambda: keyboard.send_key(keyboard.KEY_PAUSE),
        "insert": lambda: keyboard.send_key(keyboard.KEY_INSERT),
        "home": lambda: keyboard.send_key(keyboard.KEY_HOME),
        "pageup": lambda: keyboard.send_key(keyboard.KEY_PAGE_UP),
        "delete": lambda: keyboard.send_key(keyboard.KEY_DELETE),
        "end": lambda: keyboard.send_key(keyboard.KEY_END),
        "pagedown": lambda: keyboard.send_key(keyboard.KEY_PAGE_DOWN),
        "right": lambda: keyboard.send_key(keyboard.KEY_RIGHT_ARROW),
        "left": lambda: keyboard.send_key(keyboard.KEY_LEFT_ARROW),
        "down": lambda: keyboard.send_key(keyboard.KEY_DOWN_ARROW),
        "up": lambda: keyboard.send_key(keyboard.KEY_UP_ARROW),
        # Media keys
        "play": consumer.play_pause,
        "next": consumer.next_track,
        "prev": consumer.prev_track,
        "stop": consumer.stop,
        # Volume keys
        "volume-up": consumer.volume_up,
        "volume-down": consumer.volume_down,
        "mute": consumer.mute,
        # Brightness keys
        "brightness-up": consumer.brightness_up,
        "brightness-down": consumer.brightness_down,
        # Power keys
        "power": system.power,
        "sleep": system.sleep,
        "wake": system.wake,
    }

    if key in special_key_map:
        special_key_map[key]()
        return 0
    else:
        print(f"Error: Unknown special key: {key}", file=sys.stderr)
        print("See 'kinda help' for available keys", file=sys.stderr)
        return 1


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
    elif command in ["special-key", "special"]:
        return cmd_special_key(args)
    elif command in ["raw-key", "key"]:
        return cmd_raw_key(args)
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
