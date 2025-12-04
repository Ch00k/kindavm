package hid

// Keyboard modifier bits
const (
	ModNone       = 0x00
	ModLeftCtrl   = 0x01
	ModLeftShift  = 0x02
	ModLeftAlt    = 0x04
	ModLeftMeta   = 0x08
	ModRightCtrl  = 0x10
	ModRightShift = 0x20
	ModRightAlt   = 0x40
	ModRightMeta  = 0x80
)

// HID Usage IDs for special keys
const (
	KeyEscape      = 0x29
	KeyF1          = 0x3A
	KeyF2          = 0x3B
	KeyF3          = 0x3C
	KeyF4          = 0x3D
	KeyF5          = 0x3E
	KeyF6          = 0x3F
	KeyF7          = 0x40
	KeyF8          = 0x41
	KeyF9          = 0x42
	KeyF10         = 0x43
	KeyF11         = 0x44
	KeyF12         = 0x45
	KeyPrintScreen = 0x46
	KeyScrollLock  = 0x47
	KeyPause       = 0x48
	KeyInsert      = 0x49
	KeyHome        = 0x4A
	KeyPageUp      = 0x4B
	KeyDelete      = 0x4C
	KeyEnd         = 0x4D
	KeyPageDown    = 0x4E
	KeyRightArrow  = 0x4F
	KeyLeftArrow   = 0x50
	KeyDownArrow   = 0x51
	KeyUpArrow     = 0x52
)

// BrowserKeyCodeMap maps browser KeyboardEvent.code to HID Usage IDs.
// Based on the UI Events KeyboardEvent code Values specification:
// https://www.w3.org/TR/uievents-code/
var BrowserKeyCodeMap = map[string]byte{
	// Alphanumeric keys
	"KeyA": 0x04,
	"KeyB": 0x05,
	"KeyC": 0x06,
	"KeyD": 0x07,
	"KeyE": 0x08,
	"KeyF": 0x09,
	"KeyG": 0x0A,
	"KeyH": 0x0B,
	"KeyI": 0x0C,
	"KeyJ": 0x0D,
	"KeyK": 0x0E,
	"KeyL": 0x0F,
	"KeyM": 0x10,
	"KeyN": 0x11,
	"KeyO": 0x12,
	"KeyP": 0x13,
	"KeyQ": 0x14,
	"KeyR": 0x15,
	"KeyS": 0x16,
	"KeyT": 0x17,
	"KeyU": 0x18,
	"KeyV": 0x19,
	"KeyW": 0x1A,
	"KeyX": 0x1B,
	"KeyY": 0x1C,
	"KeyZ": 0x1D,

	// Number row
	"Digit1": 0x1E,
	"Digit2": 0x1F,
	"Digit3": 0x20,
	"Digit4": 0x21,
	"Digit5": 0x22,
	"Digit6": 0x23,
	"Digit7": 0x24,
	"Digit8": 0x25,
	"Digit9": 0x26,
	"Digit0": 0x27,

	// Special keys
	"Enter":     0x28,
	"Escape":    0x29,
	"Backspace": 0x2A,
	"Tab":       0x2B,
	"Space":     0x2C,

	// Punctuation
	"Minus":        0x2D,
	"Equal":        0x2E,
	"BracketLeft":  0x2F,
	"BracketRight": 0x30,
	"Backslash":    0x31,
	"Semicolon":    0x33,
	"Quote":        0x34,
	"Backquote":    0x35,
	"Comma":        0x36,
	"Period":       0x37,
	"Slash":        0x38,

	// Function keys
	"F1":  0x3A,
	"F2":  0x3B,
	"F3":  0x3C,
	"F4":  0x3D,
	"F5":  0x3E,
	"F6":  0x3F,
	"F7":  0x40,
	"F8":  0x41,
	"F9":  0x42,
	"F10": 0x43,
	"F11": 0x44,
	"F12": 0x45,

	// Navigation keys
	"PrintScreen": 0x46,
	"ScrollLock":  0x47,
	"Pause":       0x48,
	"Insert":      0x49,
	"Home":        0x4A,
	"PageUp":      0x4B,
	"Delete":      0x4C,
	"End":         0x4D,
	"PageDown":    0x4E,
	"ArrowRight":  0x4F,
	"ArrowLeft":   0x50,
	"ArrowDown":   0x51,
	"ArrowUp":     0x52,

	// Numpad
	"NumLock":        0x53,
	"NumpadDivide":   0x54,
	"NumpadMultiply": 0x55,
	"NumpadSubtract": 0x56,
	"NumpadAdd":      0x57,
	"NumpadEnter":    0x58,
	"Numpad1":        0x59,
	"Numpad2":        0x5A,
	"Numpad3":        0x5B,
	"Numpad4":        0x5C,
	"Numpad5":        0x5D,
	"Numpad6":        0x5E,
	"Numpad7":        0x5F,
	"Numpad8":        0x60,
	"Numpad9":        0x61,
	"Numpad0":        0x62,
	"NumpadDecimal":  0x63,

	// Additional keys
	"IntlBackslash": 0x64, // Non-US \ and |
	"ContextMenu":   0x65, // Application/Menu key
	"CapsLock":      0x39,
}

// Keyboard represents a HID keyboard interface
type Keyboard struct {
	device *Device
}

// NewKeyboard creates a new keyboard interface
func NewKeyboard(device *Device) *Keyboard {
	return &Keyboard{device: device}
}

// SendKeyReport sends a keyboard report with the given modifier and key codes
// Report format (9 bytes):
//
//	Byte 0: Report ID (0x01)
//	Byte 1: Modifier keys
//	Byte 2: Reserved (0x00)
//	Bytes 3-8: Key codes (up to 6 simultaneous keys)
func (k *Keyboard) SendKeyReport(modifier byte, keycodes []byte) error {
	report := make([]byte, 9)
	report[0] = 0x01 // Report ID for keyboard
	report[1] = modifier
	report[2] = 0x00 // Reserved

	// Copy up to 6 keycodes
	for i := 0; i < 6 && i < len(keycodes); i++ {
		report[3+i] = keycodes[i]
	}

	return k.device.SendReport(report, DefaultDelayMS)
}

// PressKey sends a key press (key down)
func (k *Keyboard) PressKey(modifier byte, keycodes []byte) error {
	return k.SendKeyReport(modifier, keycodes)
}

// ReleaseKey sends a key release (all keys up)
func (k *Keyboard) ReleaseKey() error {
	return k.SendKeyReport(0x00, []byte{})
}

// SendKey sends a complete key press and release
func (k *Keyboard) SendKey(modifier byte, keycode byte) error {
	if err := k.PressKey(modifier, []byte{keycode}); err != nil {
		return err
	}
	return k.ReleaseKey()
}

// SendCtrlW sends Ctrl-W key combination
func (k *Keyboard) SendCtrlW() error {
	return k.SendKey(ModLeftCtrl, BrowserKeyCodeMap["KeyW"])
}

// SendCtrlT sends Ctrl-T key combination
func (k *Keyboard) SendCtrlT() error {
	return k.SendKey(ModLeftCtrl, BrowserKeyCodeMap["KeyT"])
}

// SendCtrlN sends Ctrl-N key combination
func (k *Keyboard) SendCtrlN() error {
	return k.SendKey(ModLeftCtrl, BrowserKeyCodeMap["KeyN"])
}

// SendCtrlTab sends Ctrl-Tab key combination
func (k *Keyboard) SendCtrlTab() error {
	return k.SendKey(ModLeftCtrl, BrowserKeyCodeMap["Tab"])
}

// SendCtrlShiftTab sends Ctrl-Shift-Tab key combination
func (k *Keyboard) SendCtrlShiftTab() error {
	return k.SendKey(ModLeftCtrl|ModLeftShift, BrowserKeyCodeMap["Tab"])
}

// SendCtrlShiftT sends Ctrl-Shift-T key combination
func (k *Keyboard) SendCtrlShiftT() error {
	return k.SendKey(ModLeftCtrl|ModLeftShift, BrowserKeyCodeMap["KeyT"])
}

// SendCtrlQ sends Ctrl-Q key combination
func (k *Keyboard) SendCtrlQ() error {
	return k.SendKey(ModLeftCtrl, BrowserKeyCodeMap["KeyQ"])
}

// SendCtrlF4 sends Ctrl-F4 key combination
func (k *Keyboard) SendCtrlF4() error {
	return k.SendKey(ModLeftCtrl, KeyF4)
}

// SendAltF4 sends Alt-F4 key combination
func (k *Keyboard) SendAltF4() error {
	return k.SendKey(ModLeftAlt, KeyF4)
}

// SendF11 sends F11 key
func (k *Keyboard) SendF11() error {
	return k.SendKey(ModNone, KeyF11)
}
