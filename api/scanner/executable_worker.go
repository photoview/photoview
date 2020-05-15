package scanner

import "os/exec"

import "log"

import "fmt"

import "github.com/pkg/errors"

type ExecutableWorker struct {
	Name    string
	Path    string
	argsFmt string
}

func newExecutableWorker(name string, argsFmt string) ExecutableWorker {
	path, err := exec.LookPath(name)
	if err != nil {
		log.Printf(fmt.Sprintf("WARN: %s was not found", name))
	}

	return ExecutableWorker{
		Name:    name,
		Path:    path,
		argsFmt: argsFmt,
	}
}

func (execWorker *ExecutableWorker) isInstalled() bool {
	return execWorker.Path != ""
}

func (execWorker *ExecutableWorker) EncodeJpeg(inputPath string, outputPath string, jpegQuality int) error {
	args := fmt.Sprintf(execWorker.argsFmt, inputPath, outputPath, jpegQuality)
	cmd := exec.Command(execWorker.Path, args)

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "error encoding image using '%s'", execWorker.Name)
	}

	return nil
}

var DarktableCli = newExecutableWorker("darktable-cli", "%s %s --core --conf plugins/imageio/format/jpeg/quality=%d")
