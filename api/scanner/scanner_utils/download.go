package scanner_utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

func fileExistsLocally(testPath string) bool {
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
	baseFilePath := filepath.Base(mediaPath)
	fileName := strings.TrimSuffix(baseFilePath, filepath.Ext(mediaPath))
	fileExt := strings.TrimPrefix(filepath.Ext(mediaPath), ".")
	tempFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s_*.%s", fileName, fileExt))
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
