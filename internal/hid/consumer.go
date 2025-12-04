package hid

// Consumer control bit positions for Report ID 2
// Byte 1: Media controls
const (
	ConsumerVolumeUp      = 0x01 // bit 0
	ConsumerVolumeDown    = 0x02 // bit 1
	ConsumerMute          = 0x04 // bit 2
	ConsumerPlayPause     = 0x08 // bit 3
	ConsumerNextTrack     = 0x10 // bit 4
	ConsumerPreviousTrack = 0x20 // bit 5
	ConsumerStop          = 0x40 // bit 6
	ConsumerEmail         = 0x80 // bit 7
)

// Byte 2: Brightness and browser controls
const (
	ConsumerBrightnessUp   = 0x01 // bit 0
	ConsumerBrightnessDown = 0x02 // bit 1
	ConsumerACSearch       = 0x04 // bit 2
	ConsumerACHome         = 0x08 // bit 3
	ConsumerACBack         = 0x10 // bit 4
	ConsumerACForward      = 0x20 // bit 5
	ConsumerACStop         = 0x40 // bit 6
	ConsumerACRefresh      = 0x80 // bit 7
)

// Byte 3: Additional controls
const (
	ConsumerEject = 0x01 // bit 0
)

// Consumer represents a HID consumer control interface
type Consumer struct {
	device *Device
}

// NewConsumer creates a new consumer control interface
func NewConsumer(device *Device) *Consumer {
	return &Consumer{device: device}
}

// SendConsumerReport sends a consumer control report
// Report format (4 bytes):
//
//	Byte 0: Report ID (0x02)
//	Byte 1: Media control bits
//	Byte 2: Brightness/browser control bits
//	Byte 3: Additional control bits
func (c *Consumer) SendConsumerReport(byte1, byte2, byte3 byte) error {
	report := []byte{
		0x02,  // Report ID for consumer control
		byte1, // Media controls
		byte2, // Brightness/browser controls
		byte3, // Additional controls
	}
	return c.device.SendReport(report, DefaultDelayMS)
}

// SendConsumerKey sends a consumer key press and release
func (c *Consumer) SendConsumerKey(byte1, byte2, byte3 byte) error {
	// Press
	if err := c.SendConsumerReport(byte1, byte2, byte3); err != nil {
		return err
	}
	// Release
	return c.SendConsumerReport(0x00, 0x00, 0x00)
}

// VolumeUp sends a volume up key press
func (c *Consumer) VolumeUp() error {
	return c.SendConsumerKey(ConsumerVolumeUp, 0x00, 0x00)
}

// VolumeDown sends a volume down key press
func (c *Consumer) VolumeDown() error {
	return c.SendConsumerKey(ConsumerVolumeDown, 0x00, 0x00)
}

// BrightnessUp sends a brightness up key press
func (c *Consumer) BrightnessUp() error {
	return c.SendConsumerKey(0x00, ConsumerBrightnessUp, 0x00)
}

// BrightnessDown sends a brightness down key press
func (c *Consumer) BrightnessDown() error {
	return c.SendConsumerKey(0x00, ConsumerBrightnessDown, 0x00)
}
