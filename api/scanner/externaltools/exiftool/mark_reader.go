package exiftool

import (
	"bytes"
	"errors"
	"io"
)

// MarkReader reads from upstream and stops at marker boundaries.
type MarkReader struct {
	upstream io.Reader
	buf      []byte
	mark     []byte

	// valid data: buf[start:pending]
	// pending, need to check if contains mark: buf[pending:end]
	start   int
	pending int
	end     int

	// if true, mark begins from buf[pending]
	hasMark bool
	// if true, upstream is EOF
	upstreamEOF bool
	// if true, this reader returns EOF
	paused bool
}

// NewMarkReader creates a reader that stops at each marker and requires Reset to continue.
func NewMarkReader(upstream io.Reader, bufferSize int, mark string) (*MarkReader, error) {
	if bufferSize < 2*len(mark) {
		return nil, errors.New("buffer too small")
	}

	return &MarkReader{
		upstream: upstream,
		buf:      make([]byte, bufferSize),
		mark:     []byte(mark),
	}, nil
}

// Reset resumes reading after a previously encountered marker boundary.
func (r *MarkReader) Reset() {
	r.paused = false
}

// Read returns bytes up to (but excluding) the next marker, then reports EOF until Reset.
func (r *MarkReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	if r.paused {
		return 0, io.EOF
	}

	if r.start == r.pending {
		// compact buffer
		copy(r.buf[0:], r.buf[r.start:r.end])
		r.end -= r.start
		r.pending -= r.start
		r.start = 0

		// fill buffer
		n, err := io.ReadAtLeast(r.upstream, r.buf[r.end:], len(r.mark))
		r.end += n
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			r.upstreamEOF = true
			err = nil
		}
		if err != nil {
			return 0, err
		}

		r.checkMarkInPending()
	}

	// read from valid data
	n := copy(p, r.buf[r.start:r.pending])
	r.start += n

	if r.start == r.pending {
		switch {
		case r.hasMark:
			r.pending += len(r.mark)
			r.start += len(r.mark)
			r.hasMark = false
			r.paused = true

			r.checkMarkInPending()
		case r.upstreamEOF:
			r.paused = r.start == r.end
		}
	}

	if r.paused {
		return n, io.EOF
	}

	return n, nil
}

func (r *MarkReader) checkMarkInPending() {
	if markAt := bytes.Index(r.buf[r.pending:r.end], r.mark); markAt >= 0 {
		r.hasMark = true
		r.pending = r.pending + markAt
		return
	}

	if r.upstreamEOF {
		r.pending = r.end
		return
	}

	r.pending = max(r.pending, r.end-len(r.mark)+1)
}
