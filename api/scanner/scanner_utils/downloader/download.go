package downloader

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/afero"
)

var TempSubdir = "photoview_temp"

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

func getTempDir(albumID int) string {
	return path.Join(os.TempDir(), TempSubdir, strconv.Itoa(albumID))
}

func downloadToTempLocalPath(albumID int, fs afero.Fs, mediaPath string) (string, error) {
	// Create a temporary file
	baseFilePath := filepath.Base(mediaPath)
	fileName := strings.TrimSuffix(baseFilePath, filepath.Ext(mediaPath))
	fileExt := strings.TrimPrefix(filepath.Ext(mediaPath), ".")
	tmpDir := getTempDir(albumID)

	// Ensure the temp directory exists
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		return "", err
	}

	tempFile, err := os.CreateTemp(
		getTempDir(albumID),
		fmt.Sprintf("%s_*.%s", fileName, fileExt),
	)
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

func DownloadToLocalIfNeeded(albumID int, fs afero.Fs, mediaPath string) (string, error) {
	// Check if the file is already local
	if fileExistsLocally(mediaPath) {
		return mediaPath, nil // No cleanup needed
	}

	// If not local, download to a temporary local path
	localPath, err := downloadToTempLocalPath(albumID, fs, mediaPath)
	if err != nil {
		return "", err
	}

	return localPath, nil
}

func CleanupTempFiles(albumID int) error {
	return os.RemoveAll(
		getTempDir(albumID),
	)
}
