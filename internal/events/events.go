// Package events handles translation of browser events to HID reports.
package events

import (
	"encoding/json"
	"fmt"

	"github.com/Ch00k/kindavm/internal/hid"
)

// EventType represents the type of browser event.
type EventType string

// Event type constants for browser events
const (
	EventKeyDown        EventType = "keydown"
	EventKeyUp          EventType = "keyup"
	EventMouseMove      EventType = "mousemove"
	EventMouseDown      EventType = "mousedown"
	EventMouseUp        EventType = "mouseup"
	EventWheel          EventType = "wheel"
	EventBrightnessUp   EventType = "brightness_up"
	EventBrightnessDown EventType = "brightness_down"
	EventVolumeUp       EventType = "volume_up"
	EventVolumeDown     EventType = "volume_down"
	EventCtrlW          EventType = "ctrl_w"
	EventCtrlT          EventType = "ctrl_t"
	EventCtrlN          EventType = "ctrl_n"
	EventCtrlTab        EventType = "ctrl_tab"
	EventCtrlShiftTab   EventType = "ctrl_shift_tab"
	EventCtrlShiftT     EventType = "ctrl_shift_t"
	EventCtrlQ          EventType = "ctrl_q"
	EventCtrlF4         EventType = "ctrl_f4"
	EventAltF4          EventType = "alt_f4"
	EventF11            EventType = "f11"
)

// BrowserEvent represents an event from the browser
type BrowserEvent struct {
	Type      EventType `json:"type"`
	Code      string    `json:"code,omitempty"`      // For keyboard events
	Modifiers []string  `json:"modifiers,omitempty"` // For keyboard events
	X         int       `json:"x,omitempty"`         // For mouse move events
	Y         int       `json:"y,omitempty"`         // For mouse move events
	Button    string    `json:"button,omitempty"`    // For mouse button events
	Delta     int       `json:"delta,omitempty"`     // For wheel events
}

// Handler processes browser events and sends HID reports
type Handler struct {
	keyboard       *hid.Keyboard
	mouse          *hid.Mouse
	consumer       *hid.Consumer
	pressedKeys    map[string]bool // Track currently pressed keys
	pressedButtons map[string]bool // Track currently pressed mouse buttons
}

// NewHandler creates a new event handler
func NewHandler(device *hid.Device) *Handler {
	return &Handler{
		keyboard:       hid.NewKeyboard(device),
		mouse:          hid.NewMouse(device),
		consumer:       hid.NewConsumer(device),
		pressedKeys:    make(map[string]bool),
		pressedButtons: make(map[string]bool),
	}
}

// HandleEvent processes a browser event and sends appropriate HID reports
func (h *Handler) HandleEvent(data []byte) error {
	var event BrowserEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	switch event.Type {
	case EventKeyDown:
		return h.handleKeyDown(event)
	case EventKeyUp:
		return h.handleKeyUp(event)
	case EventMouseMove:
		return h.handleMouseMove(event)
	case EventMouseDown:
		return h.handleMouseDown(event)
	case EventMouseUp:
		return h.handleMouseUp(event)
	case EventWheel:
		return h.handleWheel(event)
	case EventBrightnessUp:
		return h.handleBrightnessUp()
	case EventBrightnessDown:
		return h.handleBrightnessDown()
	case EventVolumeUp:
		return h.handleVolumeUp()
	case EventVolumeDown:
		return h.handleVolumeDown()
	case EventCtrlW:
		return h.handleCtrlW()
	case EventCtrlT:
		return h.handleCtrlT()
	case EventCtrlN:
		return h.handleCtrlN()
	case EventCtrlTab:
		return h.handleCtrlTab()
	case EventCtrlShiftTab:
		return h.handleCtrlShiftTab()
	case EventCtrlShiftT:
		return h.handleCtrlShiftT()
	case EventCtrlQ:
		return h.handleCtrlQ()
	case EventCtrlF4:
		return h.handleCtrlF4()
	case EventAltF4:
		return h.handleAltF4()
	case EventF11:
		return h.handleF11()
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (h *Handler) handleKeyDown(event BrowserEvent) error {
	if event.Code == "" {
		return fmt.Errorf("keydown event missing code")
	}

	// Track that this key is pressed
	h.pressedKeys[event.Code] = true

	// Calculate modifier byte from modifiers array
	modifier := h.calculateModifier(event.Modifiers)

	// Collect all currently pressed keys (up to 6)
	keycodes := h.getKeycodes()

	// Send key press report
	return h.keyboard.PressKey(modifier, keycodes)
}

func (h *Handler) handleKeyUp(event BrowserEvent) error {
	if event.Code == "" {
		return fmt.Errorf("keyup event missing code")
	}

	// Track that this key is released
	delete(h.pressedKeys, event.Code)

	// Calculate modifier byte from modifiers array
	modifier := h.calculateModifier(event.Modifiers)

	// Collect remaining pressed keys
	keycodes := h.getKeycodes()

	// If no keys are pressed, send empty report
	if len(keycodes) == 0 && modifier == 0 {
		return h.keyboard.ReleaseKey()
	}

	// Otherwise send report with remaining keys
	return h.keyboard.PressKey(modifier, keycodes)
}

func (h *Handler) handleMouseMove(event BrowserEvent) error {
	// Get current button state
	buttonBits := h.getMouseButtonBits()

	// If any button is pressed, send move with button (drag)
	if buttonBits != hid.ButtonNone {
		return h.mouse.MoveWithButton(buttonBits, event.X, event.Y)
	}

	// Otherwise just move
	return h.mouse.Move(event.X, event.Y)
}

func (h *Handler) handleMouseDown(event BrowserEvent) error {
	if event.Button == "" {
		return fmt.Errorf("mousedown event missing button")
	}

	// Track that this button is pressed
	h.pressedButtons[event.Button] = true

	// Send report with all currently pressed buttons
	buttonBits := h.getMouseButtonBits()
	return h.mouse.PressButton(buttonBits)
}

func (h *Handler) handleMouseUp(event BrowserEvent) error {
	if event.Button == "" {
		return fmt.Errorf("mouseup event missing button")
	}

	// Track that this button is released
	delete(h.pressedButtons, event.Button)

	// If no buttons are pressed, release all
	if len(h.pressedButtons) == 0 {
		return h.mouse.ReleaseButton()
	}

	// Otherwise send report with remaining pressed buttons
	buttonBits := h.getMouseButtonBits()
	return h.mouse.PressButton(buttonBits)
}

func (h *Handler) handleWheel(event BrowserEvent) error {
	return h.mouse.Scroll(event.Delta)
}

// calculateModifier converts browser modifier strings to HID modifier byte
func (h *Handler) calculateModifier(modifiers []string) byte {
	var modifier byte

	for _, mod := range modifiers {
		switch mod {
		case "ctrl", "control":
			modifier |= hid.ModLeftCtrl
		case "shift":
			modifier |= hid.ModLeftShift
		case "alt":
			modifier |= hid.ModLeftAlt
		case "meta", "super", "cmd", "win":
			modifier |= hid.ModLeftMeta
		}
	}

	return modifier
}

// getKeycodes returns HID keycodes for all currently pressed keys (up to 6)
func (h *Handler) getKeycodes() []byte {
	keycodes := make([]byte, 0, 6)

	for code := range h.pressedKeys {
		if len(keycodes) >= 6 {
			break // HID keyboard supports max 6 simultaneous keys
		}

		if hidCode, exists := hid.BrowserKeyCodeMap[code]; exists {
			keycodes = append(keycodes, hidCode)
		}
	}

	return keycodes
}

// browserButtonToHID converts browser button name to HID button bits
func (h *Handler) browserButtonToHID(button string) (byte, error) {
	switch button {
	case "left", "0":
		return hid.ButtonLeft, nil
	case "middle", "1":
		return hid.ButtonMiddle, nil
	case "right", "2":
		return hid.ButtonRight, nil
	default:
		return 0, fmt.Errorf("unknown button: %s", button)
	}
}

// getMouseButtonBits returns the combined button bits for all pressed buttons
func (h *Handler) getMouseButtonBits() byte {
	var bits byte

	for button := range h.pressedButtons {
		hidButton, err := h.browserButtonToHID(button)
		if err == nil {
			bits |= hidButton
		}
	}

	return bits
}

func (h *Handler) handleBrightnessUp() error {
	return h.consumer.BrightnessUp()
}

func (h *Handler) handleBrightnessDown() error {
	return h.consumer.BrightnessDown()
}

func (h *Handler) handleVolumeUp() error {
	return h.consumer.VolumeUp()
}

func (h *Handler) handleVolumeDown() error {
	return h.consumer.VolumeDown()
}

func (h *Handler) handleCtrlW() error {
	return h.keyboard.SendCtrlW()
}

func (h *Handler) handleCtrlT() error {
	return h.keyboard.SendCtrlT()
}

func (h *Handler) handleCtrlN() error {
	return h.keyboard.SendCtrlN()
}

func (h *Handler) handleCtrlTab() error {
	return h.keyboard.SendCtrlTab()
}

func (h *Handler) handleCtrlShiftTab() error {
	return h.keyboard.SendCtrlShiftTab()
}

func (h *Handler) handleCtrlShiftT() error {
	return h.keyboard.SendCtrlShiftT()
}

func (h *Handler) handleCtrlQ() error {
	return h.keyboard.SendCtrlQ()
}

func (h *Handler) handleCtrlF4() error {
	return h.keyboard.SendCtrlF4()
}

func (h *Handler) handleAltF4() error {
	return h.keyboard.SendAltF4()
}

func (h *Handler) handleF11() error {
	return h.keyboard.SendF11()
}
