package exiftool

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

type Instance struct {
	binary  string
	version string
	cmd     *exec.Cmd
	input   io.Writer
	output  *stdtoutReader
}

func New() (*Instance, error) {
	bin, err := exec.LookPath("exiftool")
	if err != nil {
		return nil, fmt.Errorf("lookup exiftool error: %w", err)
	}

	cmd := exec.Command(bin, "-stay_open", "True", "-@", "-")
	input, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe error: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe error: %w", err)
	}
	reader := newStdoutReader(stdout, "{ready}", 1024*10)
	if err := cmd.Start(); err != nil {
		cmd.Wait()
		return nil, fmt.Errorf("launch exiftool error: %w", err)
	}

	ret := &Instance{
		binary: bin,
		cmd:    cmd,
		input:  input,
		output: reader,
	}

	version, err := ret.getRawString("-ver")
	if err != nil {
		ret.Close()
		return nil, fmt.Errorf("get exiftool version error: %w", err)
	}
	ret.version = version

	return ret, nil
}

func (i *Instance) Close() error {
	if err := i.send("-stay_open", "False"); err != nil {
		return fmt.Errorf("close exiftool error: %w", err)
	}

	if err := i.cmd.Wait(); err != nil {
		return fmt.Errorf("close exiftool error: %w", err)
	}

	return nil
}

func (i *Instance) send(args ...string) error {
	for _, arg := range args {
		if _, err := fmt.Fprint(i.input, arg); err != nil {
			return err
		}
	}

	return nil
}

func (i *Instance) fetchString() (string, error) {
	ret, err := io.ReadAll(i.output)
	if err != nil {
		return "", err
	}

	return string(ret), nil
}

func (i *Instance) fetchJson(v any) error {
	if err := json.NewDecoder(i.output).Decode(v); err != nil {
		return err
	}

	return nil
}

func (i *Instance) getRawString(args ...string) (string, error) {
	i.output.ResetFrame()

	if err := i.send(append(args, "-execute")...); err != nil {
		return "", fmt.Errorf("send command %v error: %w", args, err)
	}

	ret, err := i.fetchString()
	if err != nil {
		return "", fmt.Errorf("fetch string from command %v error: %w", args, err)
	}

	return ret, nil
}

func (i *Instance) getJson(v any, args ...string) error {
	i.output.ResetFrame()

	if err := i.send(append(args, "-j", "-execute")...); err != nil {
		return fmt.Errorf("send command %v error: %w", args, err)
	}

	if err := i.fetchJson(v); err != nil {
		return fmt.Errorf("fetch json from command %v error: %w", args, err)
	}

	return nil
}
