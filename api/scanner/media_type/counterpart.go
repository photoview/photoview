package media_type

import (
	"path"
	"path/filepath"
	"strings"
)

// FindWebCounterpart returns the filename if the file `imagePath` has a conterpart file competible with the browser.
func FindWebCounterpart(imagePath string) (string, bool) {
	return findCounterpart(imagePath, func(filename string) bool {
		mt := GetMediaType(filename)
		return mt.IsImage() && mt.IsWebCompatible()
	})
}

// FindRawConterpart returns the filename if the file `imagePath` has a conterpart file which needs to be processed before showing in the browser.
func FindRawCounterpart(imagePath string) (string, bool) {
	return findCounterpart(imagePath, func(filename string) bool {
		mt := GetMediaType(filename)
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
