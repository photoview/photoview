package actions

import (
	"strings"
	"unicode"
)

func sanitizeShareLabel(label *string) *string {
	if label == nil {
		return nil
	}

	sanitizedLabel := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, *label)
	sanitizedLabel = strings.TrimSpace(sanitizedLabel)
	if sanitizedLabel == "" {
		return nil
	}

	return &sanitizedLabel
}
