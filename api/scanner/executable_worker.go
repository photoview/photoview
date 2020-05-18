package scanner

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type DarktableWorker struct {
	path string
}

func newDarktableWorker() DarktableWorker {
	path, err := exec.LookPath("darktable-cli")
	if err != nil {
		log.Println("Executable worker not found: darktable")
	} else {
		log.Println("Found executable worker: darktable")
	}

	return DarktableWorker{
		path: path,
	}
}

func (worker *DarktableWorker) IsInstalled() bool {
	return worker.path != ""
}

func (worker *DarktableWorker) EncodeJpeg(inputPath string, outputPath string, jpegQuality int) error {
	tmpDir, err := ioutil.TempDir("/tmp", "photoview-darktable")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	args := []string{
		inputPath,
		outputPath,
		"--core",
		"--conf",
		fmt.Sprintf("plugins/imageio/format/jpeg/quality=%d", jpegQuality),
		"--configdir",
		tmpDir,
	}

	cmd := exec.Command(worker.path, args...)

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "encoding image using: %s %v", worker.path, args)
	}

	return nil
}

var DarktableCli = newDarktableWorker()
