package media_type

import (
	"path"
	"strings"

	"github.com/spf13/afero"
)

// FindWebCounterpart returns the filename if the file `imagePath` has a counterpart file competible with the browser.
func FindWebCounterpart(fs afero.Fs, imagePath string) (string, bool) {
	return findCounterpart(fs, imagePath, func(filename string) bool {
		mt := GetMediaType(filename)
		return mt.IsImage() && mt.IsWebCompatible()
	})
}

// FindRawCounterpart returns the filename if the file `imagePath` has a counterpart file which needs to be processed before showing in the browser.
func FindRawCounterpart(fs afero.Fs, imagePath string) (string, bool) {
	return findCounterpart(fs, imagePath, func(filename string) bool {
		mt := GetMediaType(filename)
		return mt.IsImage() && !mt.IsWebCompatible()
	})
}

func findCounterpart(fs afero.Fs, filename string, acceptFn func(filepath string) bool) (string, bool) {
	ext := path.Ext(filename)
	filenamePattern := strings.TrimSuffix(filename, ext) + ".*"

	filenames, err := afero.Glob(fs, filenamePattern)
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
