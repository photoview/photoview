package scanner_utils

import (
	"log"
	"os"

	"github.com/spf13/afero"
)

func FileExists(fs afero.Fs, testPath string) bool {
	_, err := fs.Stat(testPath)

	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		// unexpected error logging
		log.Printf("Error: checking for file existence (%s): %s", testPath, err)
		return false
	}
	return true
}
