package hid

import (
	"testing"
)

func TestClampMovement(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int8
	}{
		{"zero", 0, 0},
		{"positive in range", 50, 50},
		{"negative in range", -50, -50},
		{"max value", 127, 127},
		{"min value", -127, -127},
		{"over max", 200, 127},
		{"under min", -200, -127},
		{"slightly over", 128, 127},
		{"slightly under", -128, -127},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampMovement(tt.input)
			if got != tt.expected {
				t.Errorf("clampMovement(%d) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMouseReportFormat(t *testing.T) {
	tests := []struct {
		name     string
		buttons  byte
		x        int
		y        int
		wheel    int
		expected []byte
	}{
		{
			name:     "no movement no buttons",
			buttons:  ButtonNone,
			x:        0,
			y:        0,
			wheel:    0,
			expected: []byte{0x04, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "left button click",
			buttons:  ButtonLeft,
			x:        0,
			y:        0,
			wheel:    0,
			expected: []byte{0x04, 0x01, 0x00, 0x00, 0x00},
		},
		{
			name:     "right button click",
			buttons:  ButtonRight,
			x:        0,
			y:        0,
			wheel:    0,
			expected: []byte{0x04, 0x02, 0x00, 0x00, 0x00},
		},
		{
			name:     "middle button click",
			buttons:  ButtonMiddle,
			x:        0,
			y:        0,
			wheel:    0,
			expected: []byte{0x04, 0x04, 0x00, 0x00, 0x00},
		},
		{
			name:     "positive movement",
			buttons:  ButtonNone,
			x:        10,
			y:        20,
			wheel:    0,
			expected: []byte{0x04, 0x00, 0x0A, 0x14, 0x00},
		},
		{
			name:     "negative movement",
			buttons:  ButtonNone,
			x:        -10,
			y:        -20,
			wheel:    0,
			expected: []byte{0x04, 0x00, 0xF6, 0xEC, 0x00}, // -10 = 0xF6, -20 = 0xEC
		},
		{
			name:     "wheel scroll up",
			buttons:  ButtonNone,
			x:        0,
			y:        0,
			wheel:    5,
			expected: []byte{0x04, 0x00, 0x00, 0x00, 0x05},
		},
		{
			name:     "wheel scroll down",
			buttons:  ButtonNone,
			x:        0,
			y:        0,
			wheel:    -5,
			expected: []byte{0x04, 0x00, 0x00, 0x00, 0xFB}, // -5 = 0xFB
		},
		{
			name:     "drag with left button",
			buttons:  ButtonLeft,
			x:        10,
			y:        10,
			wheel:    0,
			expected: []byte{0x04, 0x01, 0x0A, 0x0A, 0x00},
		},
		{
			name:     "max values",
			buttons:  ButtonNone,
			x:        127,
			y:        127,
			wheel:    127,
			expected: []byte{0x04, 0x00, 0x7F, 0x7F, 0x7F},
		},
		{
			name:     "min values",
			buttons:  ButtonNone,
			x:        -127,
			y:        -127,
			wheel:    -127,
			expected: []byte{0x04, 0x00, 0x81, 0x81, 0x81}, // -127 = 0x81
		},
		{
			name:     "clamped over max",
			buttons:  ButtonNone,
			x:        200,
			y:        200,
			wheel:    200,
			expected: []byte{0x04, 0x00, 0x7F, 0x7F, 0x7F},
		},
		{
			name:     "clamped under min",
			buttons:  ButtonNone,
			x:        -200,
			y:        -200,
			wheel:    -200,
			expected: []byte{0x04, 0x00, 0x81, 0x81, 0x81},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xClamped := clampMovement(tt.x)
			yClamped := clampMovement(tt.y)
			wheelClamped := clampMovement(tt.wheel)

			report := []byte{
				0x04,
				tt.buttons,
				byte(xClamped),
				byte(yClamped),
				byte(wheelClamped),
			}

			for i := 0; i < len(report); i++ {
				if report[i] != tt.expected[i] {
					t.Errorf("Report byte %d = 0x%02X, want 0x%02X\nFull report: %v\nExpected:    %v",
						i, report[i], tt.expected[i], report, tt.expected)
					break
				}
			}
		})
	}
}

func TestButtonConstants(t *testing.T) {
	tests := []struct {
		name     string
		button   byte
		expected byte
	}{
		{"ButtonNone", ButtonNone, 0x00},
		{"ButtonLeft", ButtonLeft, 0x01},
		{"ButtonRight", ButtonRight, 0x02},
		{"ButtonMiddle", ButtonMiddle, 0x04},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.button != tt.expected {
				t.Errorf("%s = 0x%02X, want 0x%02X", tt.name, tt.button, tt.expected)
			}
		})
	}
}

func TestButtonCombinations(t *testing.T) {
	tests := []struct {
		name     string
		buttons  []byte
		expected byte
	}{
		{
			name:     "left+right",
			buttons:  []byte{ButtonLeft, ButtonRight},
			expected: 0x03,
		},
		{
			name:     "left+middle",
			buttons:  []byte{ButtonLeft, ButtonMiddle},
			expected: 0x05,
		},
		{
			name:     "right+middle",
			buttons:  []byte{ButtonRight, ButtonMiddle},
			expected: 0x06,
		},
		{
			name:     "all buttons",
			buttons:  []byte{ButtonLeft, ButtonRight, ButtonMiddle},
			expected: 0x07,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var combined byte
			for _, btn := range tt.buttons {
				combined |= btn
			}
			if combined != tt.expected {
				t.Errorf("Combined = 0x%02X, want 0x%02X", combined, tt.expected)
			}
		})
	}
}
