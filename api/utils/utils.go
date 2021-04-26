package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"path"
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

var test_cache_path string = ""

func ConfigureTestCache(tmp_dir string) {
	test_cache_path = tmp_dir
}

// MediaCachePath returns the path for where the media cache is located on the file system
func MediaCachePath() string {
	if test_cache_path != "" {
		return test_cache_path
	}

	photoCache := EnvMediaCachePath.GetValue()
	if photoCache == "" {
		photoCache = "./media_cache"
	}

	return photoCache
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
