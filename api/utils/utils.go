package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"
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

// SanitizeShareLabel removes control and non-semantic format characters and trims whitespace.
func SanitizeShareLabel(label *string) *string {
	if label == nil {
		return nil
	}

	sanitizedLabel := strings.Map(func(r rune) rune {
		switch {
		case unicode.IsControl(r):
			return -1
		case unicode.In(r, unicode.Cf) && r != '\u200c' && r != '\u200d':
			return -1
		default:
			return r
		}
	}, *label)
	sanitizedLabel = strings.TrimSpace(sanitizedLabel)
	if sanitizedLabel == "" {
		return nil
	}

	return &sanitizedLabel
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

// IsSubPath returns true if target is root itself or a descendant of root.
// Both paths should already be cleaned/absolute; this is meant as a defense-
// in-depth check after joining a user-supplied path onto a trusted root, not
// as the primary sanitization step.
func IsSubPath(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}

	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// SanitizeRelativePath cleans a user-supplied relative path (e.g. a form
// field name from an upload, or a folder name), rejecting absolute paths
// and any attempt to escape upwards via "..". The returned path uses the
// OS-native separator and is safe to filepath.Join onto a trusted root.
func SanitizeRelativePath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("path must not be empty")
	}

	slashed := filepath.ToSlash(p)
	if strings.HasPrefix(slashed, "/") {
		return "", fmt.Errorf("absolute paths are not allowed")
	}

	cleaned := filepath.Clean(filepath.FromSlash(slashed))
	if cleaned == "." {
		return "", fmt.Errorf("path must not be empty")
	}

	for _, part := range strings.Split(cleaned, string(filepath.Separator)) {
		if part == ".." {
			return "", fmt.Errorf("path must not contain '..'")
		}
	}

	return cleaned, nil
}

// IsDirSymlink checks that the given path is a symlink and resolves to a
// directory.
func IsDirSymlink(linkPath string) (bool, error) {

	fileInfo, err := os.Lstat(linkPath)
	if err != nil {
		return false, fmt.Errorf("cannot get fileinfo of the symlink %q: %w", linkPath, err)
	}

	// Resolve symlinks
	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		resolvedPath, err := filepath.EvalSymlinks(linkPath)
		if err != nil {
			return false, fmt.Errorf("cannot resolve symlink target for %q, skipping it: %w", linkPath, err)
		}

		resolvedFile, err := os.Stat(resolvedPath)
		if err != nil {
			return false, fmt.Errorf("cannot get fileinfo of the symlink %q target %q, skipping it: %w",
				linkPath, resolvedPath, err)
		}

		return resolvedFile.IsDir(), nil
	}

	return false, nil
}
