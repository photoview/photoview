package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"path"

	"github.com/spf13/afero"
)

func GenerateToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	charLen := big.NewInt(int64(len(charset)))

	b := make([]byte, length)
	for i := range b {

		n, err := rand.Int(rand.Reader, charLen)
		if err != nil {
			log.Panicf("Could not generate random number: %v", err)
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

type PhotoviewError struct {
	message  string
	original error
}

func (e PhotoviewError) Error() string {
	return fmt.Sprintf("%s: %s", e.message, e.original)
}

func HandleError(message string, err error) PhotoviewError {
	log.Printf("ERROR: %s: %s", message, err)
	return PhotoviewError{
		message:  message,
		original: err,
	}
}

var test_face_recognition_models_path string = ""

func ConfigureTestFaceRecognitionModelsPath(path string) {
	test_face_recognition_models_path = path
}

func FaceRecognitionModelsPath() string {
	if test_face_recognition_models_path != "" {
		return test_face_recognition_models_path
	}

	if EnvFaceRecognitionModelsPath.GetValue() == "" {
		return path.Join("data", "models")
	}

	return EnvFaceRecognitionModelsPath.GetValue()
}

// IsDirSymlink checks that the given path is a symlink and resolves to a
// directory.
func IsDirSymlink(fs afero.Fs, linkPath string) (bool, error) {

	lst, ok := fs.(afero.Symlinker)
	if !ok {
		// Not a Lstater, not a symlink
		return false, nil
	}

	// FIXME
	fileInfo, ok, err := lst.LstatIfPossible(linkPath)
	if err != nil {
		return false, fmt.Errorf("cannot get fileinfo of the symlink %q: %w", linkPath, err)
	}

	if !ok {
		// Not a symlink
		return false, nil
	}

	// Resolve symlinks
	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		resolvedPath, err := lst.ReadlinkIfPossible(linkPath)
		if err != nil {
			return false, fmt.Errorf("cannot resolve symlink target for %q, skipping it: %w", linkPath, err)
		}

		resolvedFile, err := fs.Stat(resolvedPath)
		if err != nil {
			return false, fmt.Errorf("cannot get fileinfo of the symlink %q target %q, skipping it: %w",
				linkPath, resolvedPath, err)
		}

		return resolvedFile.IsDir(), nil
	}

	return false, nil
}
