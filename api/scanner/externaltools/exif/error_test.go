package exif

import (
	"errors"
	"fmt"
	"testing"
)

func TestParseFailures(t *testing.T) {
	var pErr ParseFailures
	pErr.Append("key1", errors.New("error1"))
	pErr.Append("key2", errors.New("error2"))

	if got, want := fmt.Sprintf("%v", pErr), `["key1": error1; "key2": error2]`; got != want {
		t.Errorf("fmt.Sprintf(pErr) = %q, want: %q", got, want)
	}
}
