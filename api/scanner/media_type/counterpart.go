package media_type

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// FindWebCounterpart returns the filename if the file `imagePath` has a counterpart file competible with the browser.
func FindWebCounterpart(fs afero.Fs, imagePath string) (string, bool) {
	return findCounterpart(imagePath, func(filename string) bool {
		mt := GetMediaType(fs, filename)
		return mt.IsImage() && mt.IsWebCompatible()
	})
}

// FindRawCounterpart returns the filename if the file `imagePath` has a counterpart file which needs to be processed before showing in the browser.
func FindRawCounterpart(fs afero.Fs, imagePath string) (string, bool) {
	return findCounterpart(imagePath, func(filename string) bool {
		mt := GetMediaType(fs, filename)
		return mt.IsImage() && !mt.IsWebCompatible()
	})
}

func findCounterpart(filename string, acceptFn func(filepath string) bool) (string, bool) {
	ext := path.Ext(filename)
	filenamePattern := strings.TrimSuffix(filename, ext) + ".*"

	filenames, err := filepath.Glob(filenamePattern)
	if err != nil {
		return "", false
	}

	for _, f := range filenames {
		if f == filename {
			continue
		}

		if acceptFn(f) {
			return f, true
		}
	}

	return "", false
}
