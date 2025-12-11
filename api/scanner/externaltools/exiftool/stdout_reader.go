package exiftool

import (
	"bufio"
	"errors"
	"io"
)

var errEndOfFrame = errors.New("end of the frame")

// stdtoutReader is only designed to work with exiftool json output in stay_open mode.
// Don't use it in other cases.
type stdtoutReader struct {
	delimiter  string
	reader     *bufio.Reader
	lastRemain []byte
	lastError  error
}

func newStdoutReader(stdout io.Reader, delimiter string, bufSize int) *stdtoutReader {
	return &stdtoutReader{
		delimiter: delimiter,
		reader:    bufio.NewReaderSize(stdout, max(len(delimiter)*2, bufSize)),
	}
}

func (r *stdtoutReader) ResetFrame() {
	if r.lastError == errEndOfFrame {
		r.lastError = nil
	}
}

func (r *stdtoutReader) Read(p []byte) (int, error) {
	if err := r.lastError; err != nil {
		if r.lastError == errEndOfFrame {
			err = io.EOF
		}
		return 0, err
	}

	n := 0
	for len(p) > 0 {
		if r.lastRemain != nil {
			n += copyAndMove(&p, &r.lastRemain)
			if len(r.lastRemain) == 0 {
				r.lastRemain = nil
			}

			continue
		}

		line, prefix, err := r.reader.ReadLine()
		if err != nil {
			r.lastError = err
			return n, r.lastError
		}

		if !prefix && string(line) == r.delimiter {
			r.lastError = errEndOfFrame
			return n, io.EOF
		}

		n += copyAndMove(&p, &line)
		if len(line) > 0 {
			r.lastRemain = line
		}
	}

	return n, nil
}

func copyAndMove(dst, src *[]byte) int {
	size := copy(*dst, *src)
	*dst = (*dst)[size:]
	*src = (*src)[size:]

	return size
}
