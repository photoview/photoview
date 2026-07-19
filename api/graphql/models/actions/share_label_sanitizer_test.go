package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeShareLabel(t *testing.T) {
	tests := []struct {
		name     string
		label    *string
		expected *string
	}{
		{
			name: "nil label",
		},
		{
			name:     "control characters",
			label:    shareLabelPointer("Fam\x00ily\r\n\t\u007f\u0085 album"),
			expected: shareLabelPointer("Family album"),
		},
		{
			name:     "surrounding whitespace",
			label:    shareLabelPointer("\u00a0  Family album \t"),
			expected: shareLabelPointer("Family album"),
		},
		{
			name:  "empty label",
			label: shareLabelPointer(""),
		},
		{
			name:  "control characters only",
			label: shareLabelPointer("\x00\r\n\t\u007f\u0085"),
		},
		{
			name:     "ordinary unicode",
			label:    shareLabelPointer("  Семья 日本語 👩‍👩‍👧‍👦  "),
			expected: shareLabelPointer("Семья 日本語 👩‍👩‍👧‍👦"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, sanitizeShareLabel(test.label))
		})
	}
}

func shareLabelPointer(value string) *string {
	return &value
}
