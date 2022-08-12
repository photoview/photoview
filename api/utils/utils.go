package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
)

func GenerateToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	charLen := big.NewInt(int64(len(charset)))

	b := make([]byte, length)
	for i := range b {

		n, err := rand.Int(rand.Reader, charLen)
		if err != nil {
			log.Fatalf("Could not generate random number: %s\n", err)
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
func IsDirSymlink(path string) (bool, error) {
	isDirSymlink := false

	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false, errors.Wrapf(err, "could not stat %s", path)
	}

	//Resolve symlinks
	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		resolvedPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return false, errors.Wrapf(err, "Cannot resolve linktarget of %s, ignoring it", path)
		}

		resolvedFile, err := os.Stat(resolvedPath)
		if err != nil {
			return false, errors.Wrapf(err, "Cannot get fileinfo of linktarget %s of symlink %s, ignoring it", resolvedPath, path)
		}
		isDirSymlink = resolvedFile.IsDir()

		return isDirSymlink, nil
	}

	return false, nil
}
