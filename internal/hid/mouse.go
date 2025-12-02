package hid

// Mouse button bits
const (
	ButtonNone   = 0x00
	ButtonLeft   = 0x01
	ButtonRight  = 0x02
	ButtonMiddle = 0x04
)

// Mouse represents a HID mouse interface
type Mouse struct {
	device *Device
}

// NewMouse creates a new mouse interface
func NewMouse(device *Device) *Mouse {
	return &Mouse{device: device}
}

// clampMovement clamps a movement value to the valid range (-127 to 127)
func clampMovement(value int) int8 {
	if value > 127 {
		return 127
	}
	if value < -127 {
		return -127
	}
	return int8(value)
}

// SendMouseReport sends a mouse HID report
// Report format (5 bytes):
//
//	Byte 0: Report ID (0x04)
//	Byte 1: Buttons (bit 0: left, bit 1: right, bit 2: middle, bits 3-7: padding)
//	Byte 2: X movement (signed 8-bit, -127 to +127)
//	Byte 3: Y movement (signed 8-bit, -127 to +127)
//	Byte 4: Wheel (signed 8-bit, -127 to +127)
func (m *Mouse) SendMouseReport(buttons byte, x, y, wheel int) error {
	xClamped := clampMovement(x)
	yClamped := clampMovement(y)
	wheelClamped := clampMovement(wheel)

	report := []byte{
		0x04,               // Report ID for mouse
		buttons,            // Button bits
		byte(xClamped),     // X movement (signed to unsigned byte)
		byte(yClamped),     // Y movement (signed to unsigned byte)
		byte(wheelClamped), // Wheel movement (signed to unsigned byte)
	}

	return m.device.SendReport(report, DefaultDelayMS)
}

// Move moves the mouse cursor
func (m *Mouse) Move(x, y int) error {
	return m.SendMouseReport(ButtonNone, x, y, 0)
}

// Click performs a mouse button click
func (m *Mouse) Click(button byte) error {
	// Press
	if err := m.SendMouseReport(button, 0, 0, 0); err != nil {
		return err
	}
	// Release
	return m.SendMouseReport(ButtonNone, 0, 0, 0)
}

// PressButton presses a mouse button (holds it down)
func (m *Mouse) PressButton(button byte) error {
	return m.SendMouseReport(button, 0, 0, 0)
}

// ReleaseButton releases all mouse buttons
func (m *Mouse) ReleaseButton() error {
	return m.SendMouseReport(ButtonNone, 0, 0, 0)
}

// Scroll scrolls the mouse wheel
func (m *Mouse) Scroll(amount int) error {
	return m.SendMouseReport(ButtonNone, 0, 0, amount)
}

// MoveWithButton moves the mouse with a button held down (for dragging)
func (m *Mouse) MoveWithButton(button byte, x, y int) error {
	return m.SendMouseReport(button, x, y, 0)
}
