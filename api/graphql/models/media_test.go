package models

import (
	"testing"
)

func TestSanitizeMediaName(t *testing.T) {
	tests := [][2]string{
		{"filename.png", "filename_png"},
		{"../..\\escape", "____escape"},
		{"..", "__"},
		{"..\\/", "__"},
	}

	for i, test := range tests {
		if SanitizeMediaName(test[0]) != test[1] {
			t.Errorf("SanitizeMediaName test %d failed: got '%s', expected '%s'", i, test[1], SanitizeMediaName(test[0]))
		}
	}
}
