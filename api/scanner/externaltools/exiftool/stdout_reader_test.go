package exiftool

import (
	"fmt"
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
		{"1ShortFrame", "0123\n{ready}\n", []string{"0123\n", ""}},
		{"1LongFrame", "01234567890123456789\n{ready}\n", []string{"01234567890123456789\n", ""}},
		{"1MultilineFrame", "012345\n67890123456789\n{ready}\n", []string{"012345\n67890123456789\n", ""}},

		{"2EmptyFrame", "{ready}\n{ready}\n", []string{"", "", ""}},
		{"2ShortFrame", "0123\n{ready}\nabcd\n{ready}\n", []string{"0123\n", "abcd\n", ""}},
		{"2LongFrame", "01234567890123456789\n{ready}\nabcdefghijklmnopq\n{ready}\n", []string{"01234567890123456789\n", "abcdefghijklmnopq\n", ""}},
		{"2MultilineFrame", "012345\n67890123456789\n{ready}\nabcdef\nghijklmnopq\n{ready}\n", []string{"012345\n67890123456789\n", "abcdef\nghijklmnopq\n", ""}},

		{"1Error", "Error: some error\n{ready}\n", []string{"E: some error"}},
		{"1Error1Empty", "Error: some error\n{ready}\n{ready}\n", []string{"E: some error", ""}},
		{"1Empty1Error", "{ready}\nError: some error\n{ready}\n", []string{"", "E: some error"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newStdoutReader(strings.NewReader(tc.input), "{ready}", 10240)
			for i, want := range tc.wantFrames {
				got, err := io.ReadAll(r)

				if strings.HasPrefix(want, "E: ") {
					if fmt.Sprintf("E: %v", err) != want {
						t.Errorf("ReadAll(r) = (%q, %v), want: %v", got, err, want)
					}

					r.ResetFrame()
					continue
				}

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
