package scanner

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type ExecutableWorker struct {
	Name    string
	Path    string
	argsFmt string
}

func newExecutableWorker(name string, argsFmt string) ExecutableWorker {
	path, err := exec.LookPath(name)
	if err != nil {
		log.Printf("WARN: %s was not found\n", name)
	} else {
		log.Printf("Found executable worker: %s\n", name)
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
	args := make([]string, 0)
	for _, arg := range strings.Split(execWorker.argsFmt, " ") {
		if strings.Contains(arg, "%") {
			arg = fmt.Sprintf(arg, inputPath, outputPath, jpegQuality)
		}
		args = append(args, arg)
	}

	cmd := exec.Command(execWorker.Path, args...)

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "error encoding image using: %s %v", execWorker.Name, args)
	}

	return nil
}

var DarktableCli = newExecutableWorker("darktable-cli", "%[1]s %[2]s --core --conf plugins/imageio/format/jpeg/quality=%[3]d")
