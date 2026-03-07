package exiftool

import (
	"io"
	"runtime"
	"strings"
	"testing"
	"testing/iotest"
)

func readAllString(t *testing.T, r io.Reader) string {
	t.Helper()

	var data []byte

	for {
		buf := make([]byte, 1024)

		var n int
		var err error

		var memstats runtime.MemStats
		runtime.ReadMemStats(&memstats)
		mallocs := 0 - memstats.Mallocs

		n, err = r.Read(buf)

		runtime.ReadMemStats(&memstats)
		mallocs += memstats.Mallocs

		if mallocs > 0 {
			t.Errorf("r.Read() allocs %d bytes", mallocs)
		}

		data = append(data, buf[:n]...)

		if err != nil {
			if err == io.EOF {
				break
			}

			t.Fatalf("read failed: %v", err)
		}

		if n == 0 {
			t.Error("read error, return 0 byte without errors")
		}
	}

	return string(data)
}

func TestNewMarkReader(t *testing.T) {
	tests := []struct {
		name       string
		bufferSize int
		mark       string
		wantErr    string
	}{
		{
			name:       "BufferTooSmall",
			bufferSize: 5,
			mark:       "abcd",
			wantErr:    "buffer too small",
		},
		{
			name:       "TwiceMarkLen",
			bufferSize: 8,
			mark:       "abcd",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewMarkReader(strings.NewReader("x"), tc.bufferSize, tc.mark)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("NewMarkReader() returns error: %v", err)
				}
				return
			}

			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("err = %v, want: %q", err, tc.wantErr)
			}
		})
	}
}

func TestMarkReader(t *testing.T) {
	tests := []struct {
		name         string
		mark         string
		input        string
		bufferSize   int
		wantSegments []string
	}{
		{
			name:         "CrossRead",
			mark:         "<<MARK>>",
			input:        "hello<<MARK>>world",
			bufferSize:   2 * 8, /* len(mark) = 8 */
			wantSegments: []string{"hello", "world"},
		},
		{
			name:         "EndWithMark",
			mark:         "<<MARK>>",
			input:        "foo<<MARK>>bar<<MARK>>",
			bufferSize:   64,
			wantSegments: []string{"foo", "bar"},
		},
		{
			name:         "TrailingPrefixMark",
			mark:         "<<MARK>>",
			input:        "foo<<MARK>>bar<<MAR",
			bufferSize:   64,
			wantSegments: []string{"foo", "bar<<MAR"},
		},
		{
			name:         "Empty",
			mark:         "<<MARK>>",
			input:        "<<MARK>><<MARK>>abc<<MARK>>",
			bufferSize:   64,
			wantSegments: []string{"", "", "abc"},
		},
		{
			name:         "LongMark",
			mark:         "01234567890123456789",
			input:        "a01234567890123456789b",
			bufferSize:   64,
			wantSegments: []string{"a", "b"},
		},
		{
			name:         "RepeatMark",
			mark:         "aaaaa", /* a * 5 */
			input:        "0aaaaa1aaa2aaaa3aaaaa4aaaaaa5",
			bufferSize:   64,
			wantSegments: []string{"0", "1aaa2aaaa3", "4", "a5"},
		},
		{
			name:         "OneByteMark",
			mark:         "|",
			input:        "0|1|234|5|6",
			bufferSize:   64,
			wantSegments: []string{"0", "1", "234", "5", "6"},
		},
	}

	readWraper := []struct {
		name string
		wrap func(r io.Reader) io.Reader
	}{
		{"Original", func(r io.Reader) io.Reader { return r }},
		{"Half", iotest.HalfReader},
		{"OneByte", iotest.OneByteReader},
	}

	for _, wrapper := range readWraper {
		t.Run(wrapper.name, func(t *testing.T) {

			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {

					up := wrapper.wrap(strings.NewReader(tc.input))

					r, err := NewMarkReader(up, tc.bufferSize, tc.mark)
					if err != nil {
						t.Fatal("NewMarkReader() returns error:", err)
					}

					for i, want := range tc.wantSegments {
						if got := readAllString(t, r); got != want {
							t.Errorf("segment[%d] = %q, want: %q", i, got, want)
						}

						var buf [1024]byte
						if _, err := r.Read(buf[:]); err != io.EOF {
							t.Error("r.Read() after readAllString() returns non-EOF error:", err)
						}

						r.Reset()
					}

					var buf [1024]byte
					if _, err := r.Read(buf[:]); err != io.EOF {
						t.Error("r.Read() after all segments returns non-EOF error:", err)
					}
				})
			}
		})
	}
}
