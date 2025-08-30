package exif

import (
	"errors"
	"fmt"
	"strings"

	"github.com/barasher/go-exiftool"
)

type parseFailure struct {
	key string
	err error
}

func (e parseFailure) String() string {
	return fmt.Sprintf(`%q: %v`, e.key, e.err)
}

type ParseFailures []parseFailure

func (e *ParseFailures) Append(key string, err error) {
	if errors.Is(err, exiftool.ErrKeyNotFound) {
		return
	}

	*e = append(*e, parseFailure{
		key: key,
		err: err,
	})
}

func (e ParseFailures) String() string {
	errStrs := make([]string, 0, len(e))
	for _, pe := range e {
		errStrs = append(errStrs, pe.String())
	}

	return fmt.Sprintf("[%s]", strings.Join(errStrs, "; "))
}
