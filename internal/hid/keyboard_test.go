package hid

import (
	"bytes"
	"testing"
)

func TestBrowserKeyCodeMap(t *testing.T) {
	tests := []struct {
		code     string
		expected byte
	}{
		{"KeyA", 0x04},
		{"KeyZ", 0x1D},
		{"Digit1", 0x1E},
		{"Digit0", 0x27},
		{"Enter", 0x28},
		{"Space", 0x2C},
		{"ArrowUp", 0x52},
		{"F1", 0x3A},
		{"F12", 0x45},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got, exists := BrowserKeyCodeMap[tt.code]
			if !exists {
				t.Errorf("BrowserKeyCodeMap[%s] does not exist", tt.code)
				return
			}
			if got != tt.expected {
				t.Errorf("BrowserKeyCodeMap[%s] = 0x%02X, want 0x%02X", tt.code, got, tt.expected)
			}
		})
	}
}

func TestKeyboardReportFormat(t *testing.T) {
	tests := []struct {
		name     string
		modifier byte
		keycodes []byte
		expected []byte
	}{
		{
			name:     "single key no modifier",
			modifier: ModNone,
			keycodes: []byte{0x04}, // 'a'
			expected: []byte{0x01, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "single key with shift",
			modifier: ModLeftShift,
			keycodes: []byte{0x04}, // 'A'
			expected: []byte{0x01, 0x02, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "multiple keys",
			modifier: ModNone,
			keycodes: []byte{0x04, 0x05, 0x06}, // 'a', 'b', 'c'
			expected: []byte{0x01, 0x00, 0x00, 0x04, 0x05, 0x06, 0x00, 0x00, 0x00},
		},
		{
			name:     "six keys (maximum)",
			modifier: ModNone,
			keycodes: []byte{0x04, 0x05, 0x06, 0x07, 0x08, 0x09},
			expected: []byte{0x01, 0x00, 0x00, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09},
		},
		{
			name:     "empty report (all keys up)",
			modifier: 0x00,
			keycodes: []byte{},
			expected: []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "ctrl+alt+del",
			modifier: ModLeftCtrl | ModLeftAlt,
			keycodes: []byte{KeyDelete},
			expected: []byte{0x01, 0x05, 0x00, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := make([]byte, 9)
			report[0] = 0x01 // Report ID
			report[1] = tt.modifier
			report[2] = 0x00 // Reserved

			for i := 0; i < 6 && i < len(tt.keycodes); i++ {
				report[3+i] = tt.keycodes[i]
			}

			if !bytes.Equal(report, tt.expected) {
				t.Errorf("Report = %v, want %v", report, tt.expected)
			}
		})
	}
}

func TestModifierConstants(t *testing.T) {
	tests := []struct {
		name     string
		modifier byte
		expected byte
	}{
		{"ModNone", ModNone, 0x00},
		{"ModLeftCtrl", ModLeftCtrl, 0x01},
		{"ModLeftShift", ModLeftShift, 0x02},
		{"ModLeftAlt", ModLeftAlt, 0x04},
		{"ModLeftMeta", ModLeftMeta, 0x08},
		{"ModRightCtrl", ModRightCtrl, 0x10},
		{"ModRightShift", ModRightShift, 0x20},
		{"ModRightAlt", ModRightAlt, 0x40},
		{"ModRightMeta", ModRightMeta, 0x80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.modifier != tt.expected {
				t.Errorf("%s = 0x%02X, want 0x%02X", tt.name, tt.modifier, tt.expected)
			}
		})
	}
}

func TestModifierCombinations(t *testing.T) {
	tests := []struct {
		name     string
		mods     []byte
		expected byte
	}{
		{
			name:     "Ctrl+Shift",
			mods:     []byte{ModLeftCtrl, ModLeftShift},
			expected: 0x03,
		},
		{
			name:     "Ctrl+Alt",
			mods:     []byte{ModLeftCtrl, ModLeftAlt},
			expected: 0x05,
		},
		{
			name:     "Ctrl+Alt+Shift",
			mods:     []byte{ModLeftCtrl, ModLeftAlt, ModLeftShift},
			expected: 0x07,
		},
		{
			name: "All modifiers",
			mods: []byte{
				ModLeftCtrl,
				ModLeftShift,
				ModLeftAlt,
				ModLeftMeta,
				ModRightCtrl,
				ModRightShift,
				ModRightAlt,
				ModRightMeta,
			},
			expected: 0xFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var combined byte
			for _, mod := range tt.mods {
				combined |= mod
			}
			if combined != tt.expected {
				t.Errorf("Combined = 0x%02X, want 0x%02X", combined, tt.expected)
			}
		})
	}
}
