package exiftool

import (
	"io"
	"strings"
	"testing"
)

func TestStdoutReader(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantFrames []string
	}{
		{"NoContent", "", []string{""}},
		{"1EmptyFrame", "{ready}\n", []string{"", ""}},
		{"1ShortFrame", "0123\n{ready}\n", []string{"0123", ""}},
		{"1LongFrame", "01234567890123456789\n{ready}\n", []string{"01234567890123456789", ""}},
		{"1MultilineFrame", "012345\n67890123456789\n{ready}\n", []string{"01234567890123456789", ""}},

		{"2EmptyFrame", "{ready}\n{ready}\n", []string{"", "", ""}},
		{"2ShortFrame", "0123\n{ready}\nabcd\n{ready}\n", []string{"0123", "abcd", ""}},
		{"2LongFrame", "01234567890123456789\n{ready}\nabcdefghijklmnopq\n{ready}\n", []string{"01234567890123456789", "abcdefghijklmnopq", ""}},
		{"2MultilineFrame", "012345\n67890123456789\n{ready}\nabcdef\nghijklmnopq\n{ready}\n", []string{"01234567890123456789", "abcdefghijklmnopq", ""}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newStdoutReader(strings.NewReader(tc.input), "{ready}", 10240)
			for i, want := range tc.wantFrames {
				got, err := io.ReadAll(r)
				if err != nil {
					t.Fatalf("read frame %d error: %v", i, err)
				}

				if got, want := string(got), want; got != want {
					t.Errorf("got = %q, want = % q", got, want)
				}

				r.ResetFrame()
			}
		})
	}
}
