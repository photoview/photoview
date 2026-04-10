package darktable

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/photoview/photoview/api/scanner/externaltools"
)

type Darktable struct {
	mu      sync.Mutex
	path    string
	version string
	err     error
}

func New() *Darktable {
	path, err := exec.LookPath("darktable-cli")
	if err != nil {
		return &Darktable{
			err: fmt.Errorf("darktable cli: %w: %w", externaltools.ErrInvalidCLI, err),
		}
	}

	voutput, err := exec.Command(path, "--version").Output()
	if err != nil {
		return &Darktable{
			err: fmt.Errorf("darktable cli version: %w: %w", externaltools.ErrInvalidCLI, err),
		}
	}

	version := string(bytes.Trim(bytes.SplitN(voutput, []byte("\n"), 2)[0], " \t\n\r"))

	return &Darktable{
		path:    path,
		version: version,
	}
}

func (cli *Darktable) PathVersion() (string, string, error) {
	return cli.path, cli.version, cli.err
}

func (cli *Darktable) EncodeJpeg(inputPath string, outputPath string, jpegQuality uint) error {
	if cli.err != nil {
		return cli.err
	}

	tmpDir, err := os.MkdirTemp("", "photoview-darktable-*")
	if err != nil {
		return fmt.Errorf("can't create tmp: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	args := []string{
		inputPath, outputPath,
		"--core",
		"--conf", fmt.Sprintf("plugins/imageio/format/jpeg/quality=%d", jpegQuality),
		"--configdir", tmpDir,
	}

	cmd := exec.Command(cli.path, args...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("run darktable(configdir:%q) %v error: %w: %s", tmpDir, args, err, string(output))
	}

	return nil
}
