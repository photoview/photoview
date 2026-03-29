package exiftool

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// Exiftool launches an external `exiftool` process to query photos' exif info.
// It doesn't support concurrency usage.
type Exiftool struct {
	path    string
	version string
	marker  string

	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdinBuf *bufio.Writer
	stdout   *MarkReader
	stderr   *MarkReader

	closeOnce sync.Once
}

const marker = "{ready}\n"
const bufferSize = 10240

// New returns a new instance of Exiftool.
func New() (*Exiftool, error) {
	path, err := exec.LookPath("exiftool")
	if err != nil {
		return nil, err
	}

	voutput, err := exec.Command(path, "-ver").Output()
	if err != nil {
		return nil, fmt.Errorf("run `exiftool -ver` error: %w", err)
	}

	version := string(bytes.Trim(voutput, " \t\n\r"))
	if version == "" {
		return nil, fmt.Errorf("run `exiftool -ver` error: no output")
	}

	cmd := exec.Command(path, "-stay_open", "True", "-@", "-")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("get `exiftool -stay_open True -@ -` stdin error: %w", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		defer stdin.Close()
		return nil, fmt.Errorf("get `exiftool -stay_open True -@ -` stdout error: %w", err)
	}

	stdout, err := NewMarkReader(stdoutPipe, bufferSize, marker)
	if err != nil {
		defer stdin.Close()
		return nil, err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		defer stdin.Close()
		return nil, fmt.Errorf("get `exiftool -stay_open True -@ -` stderr error: %w", err)
	}

	stderr, err := NewMarkReader(stderrPipe, bufferSize, marker)
	if err != nil {
		defer stdin.Close()
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		defer stdin.Close()
		return nil, fmt.Errorf("run `exiftool -stay_open True -@ -` error: %w", err)
	}

	return &Exiftool{
		path:     path,
		version:  version,
		marker:   marker,
		cmd:      cmd,
		stdin:    stdin,
		stdinBuf: bufio.NewWriterSize(stdin, bufferSize),
		stdout:   stdout,
		stderr:   stderr,
	}, nil
}

// BinaryPath returns the path of `exiftool` binary.
func (e *Exiftool) BinaryPath() string {
	return e.path
}

// Version returns the version of `exiftool` binary.
func (e *Exiftool) Version() string {
	return e.version
}

// Close stops the external process of `exiftool`.
func (e *Exiftool) Close() (err error) {
	e.closeOnce.Do(func() {
		defer e.cmd.Wait()
		defer e.stdin.Close()

		if err = e.rawSendCommand("-stay_open", "False"); err != nil {
			_ = e.cmd.Process.Kill()
		}
	})

	return
}

func (e *Exiftool) rawSendCommand(args ...string) error {
	e.stdout.Reset()
	e.stderr.Reset()

	for _, arg := range args {
		if _, err := e.stdinBuf.WriteString(arg); err != nil {
			return err
		}
		if err := e.stdinBuf.WriteByte('\n'); err != nil {
			return err
		}
	}

	for _, arg := range []string{"-echo4\n", "{ready}\n", "-execute\n"} {
		if _, err := e.stdinBuf.WriteString(arg); err != nil {
			return err
		}
	}

	return e.stdinBuf.Flush()
}

func (e *Exiftool) rawReadStderr() error {
	stderrOutput, err := io.ReadAll(e.stderr)
	if err != nil {
		return err
	}

	outStr := string(bytes.Trim(stderrOutput, " \n\t\r"))
	if outStr == "" {
		return nil
	}

	return errors.New(outStr)
}

func (e *Exiftool) rawGetTags(v any, args ...string) (err error) {
	if err = e.rawSendCommand(append(args, "-json")...); err != nil {
		return
	}
	defer func() {
		_, _ = io.Copy(io.Discard, e.stdout)
		err = e.rawReadStderr()
	}()

	if err = json.NewDecoder(e.stdout).Decode(v); err != nil {
		return
	}

	return nil
}

func (e *Exiftool) rawSaveEmbedFile(outputPath string, args ...string) (hasEmbededFile bool, err error) {
	if err = e.rawSendCommand(append(args, "-b", "-W", outputPath)...); err != nil {
		return
	}
	defer func() {
		_, _ = io.Copy(io.Discard, e.stdout)
		err = e.rawReadStderr()
	}()

	var output []byte
	if output, err = io.ReadAll(e.stdout); err != nil {
		return
	}

	outStr := string(bytes.Trim(output, " \n\r\t"))
	switch outStr {
	case "0 output files created":
		return
	case "1 output files created":
	default:
		err = fmt.Errorf("invalid output: %s", outStr)
		return
	}

	if err = e.rawReadStderr(); err != nil {
		return
	}

	hasEmbededFile = true
	return
}

func (e *Exiftool) rawUpdateFile(args ...string) (err error) {
	if err = e.rawSendCommand(args...); err != nil {
		return
	}
	defer func() {
		_, _ = io.Copy(io.Discard, e.stdout)
		err = e.rawReadStderr()
	}()

	var output []byte
	if output, err = io.ReadAll(e.stdout); err != nil {
		return
	}

	outStr := string(bytes.Trim(output, " \n\r\t"))
	if outStr != "1 image files updated" {
		err = fmt.Errorf("invalid output: %s", outStr)
		return
	}

	return
}

// QueryJSONTagsByNumber queries the exif info of `file` with a given struct `value`. Tags are fields of the `value`. The values are a number if possible.
// See values.go for example structs.
func (e *Exiftool) QueryJSONTagsByNumber(file string, value any) error {
	rows := []any{value}
	if err := e.rawGetTags(&rows, "-n", file); err != nil {
		return fmt.Errorf("query %q tags error: %w", file, err)
	}

	if len(rows) != 1 {
		return fmt.Errorf("query %q tags error: return %d responses, should be only 1", file, len(rows))
	}

	return nil
}

// SaveJPEGPreview saves a preview jpeg from `src` to `previewOutput`.
func (e *Exiftool) SaveJPEGPreview(src string, previewOutput string) (bool, error) {
	saved, err := e.rawSaveEmbedFile(previewOutput, "-JpgFromRaw", src)
	if err != nil {
		return false, fmt.Errorf("save jpeg preview for %q error: %w", src, err)
	}

	if !saved {
		return false, nil
	}

	if err = e.rawUpdateFile("-TagsFromFile", src, "-overwrite_original", previewOutput); err != nil {
		return false, fmt.Errorf("save tags to jpeg preview for %q error: %w", src, err)
	}

	return true, nil
}
