//go:build no_face_detection

package face_detection

import (
	"log"

	"gorm.io/gorm"
)

func InitializeFaceDetector(db *gorm.DB) error {
	log.Printf("Face detection disabled (at build-time)")
	return nil
}
