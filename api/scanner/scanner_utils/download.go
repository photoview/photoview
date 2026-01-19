package scanner_utils

import (
	"io"
	"log"
	"os"

	"github.com/spf13/afero"
)

func fileExistsLocally(fs afero.Fs, testPath string) bool {
	_, err := os.Stat(testPath)

	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		// unexpected error logging
		log.Printf("Error: checking for file existence (%s): %s", testPath, err)
		return false
	}
	return true
}

func downloadToTempLocalPath(fs afero.Fs, mediaPath string) (string, error) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "photoview_media_*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Open the source file from the filesystem
	srcFile, err := fs.Open(mediaPath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	// Copy the contents to the temporary file
	_, err = io.Copy(tempFile, srcFile)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func DownloadToLocalIfNeeded(fs afero.Fs, mediaPath string) (string, error) {
	// Check if the file is already local
	if fileExistsLocally(fs, mediaPath) {
		return mediaPath, nil
	}

	// If not local, download to a temporary local path
	localPath, err := downloadToTempLocalPath(fs, mediaPath)
	if err != nil {
		return "", err
	}

	return localPath, nil
}
