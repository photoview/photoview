package face_detection

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/Kagami/go-face"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type FaceDetector struct {
	mutex   sync.Mutex
	db      *gorm.DB
	rec     *face.Recognizer
	samples []face.Descriptor
	cats    []int32
}

var GlobalFaceDetector FaceDetector

func InitializeFaceDetector(db *gorm.DB) error {

	log.Println("Initializing face detector")

	rec, err := face.NewRecognizer(filepath.Join("data", "models"))
	if err != nil {
		return errors.Wrap(err, "initialize facedetect recognizer")
	}

	samples, cats, err := getSamplesFromDatabase(db)
	if err != nil {
		return errors.Wrap(err, "get face detection samples from database")
	}

	GlobalFaceDetector = FaceDetector{
		db:      db,
		rec:     rec,
		samples: samples,
		cats:    cats,
	}

	return nil
}

func getSamplesFromDatabase(db *gorm.DB) (samples []face.Descriptor, cats []int32, err error) {

	var imageFaces []*models.ImageFace

	if err = db.Find(&imageFaces).Error; err != nil {
		return
	}

	samples = make([]face.Descriptor, len(imageFaces))
	cats = make([]int32, len(imageFaces))

	for i, imgFace := range imageFaces {
		samples[i] = face.Descriptor(imgFace.Descriptor)
		cats[i] = int32(imgFace.FaceGroupID)
	}

	return
}

// DetectFaces finds the faces in the given image and saves them to the database
func (fd *FaceDetector) DetectFaces(media *models.Media) error {
	if err := fd.db.Model(media).Preload("MediaURL").First(&media).Error; err != nil {
		return err
	}

	var thumbnailURL *models.MediaURL
	for _, url := range media.MediaURL {
		if url.Purpose == models.PhotoThumbnail {
			thumbnailURL = &url
			thumbnailURL.Media = media
			break
		}
	}

	if thumbnailURL == nil {
		return errors.New("thumbnail url is missing")
	}

	thumbnailPath, err := thumbnailURL.CachedPath()
	if err != nil {
		return err
	}

	fd.mutex.Lock()
	faces, err := fd.rec.RecognizeFile(thumbnailPath)
	fd.mutex.Unlock()

	if err != nil {
		return errors.Wrap(err, "error read faces")
	}

	for _, face := range faces {
		fd.classifyFace(&face, media)
	}

	return nil
}

func (fd *FaceDetector) classifyFace(face *face.Face, media *models.Media) error {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	match := fd.rec.ClassifyThreshold(face.Descriptor, 0.2)

	imageFace := models.ImageFace{
		MediaID:    media.ID,
		Descriptor: models.FaceDescriptor(face.Descriptor),
	}

	var faceGroup models.FaceGroup

	// If no match add it new to samples
	if match < 0 {
		log.Println("No match, assigning new face")

		faceGroup = models.FaceGroup{
			ImageFaces: []models.ImageFace{imageFace},
		}

		if err := fd.db.Create(&faceGroup).Error; err != nil {
			return err
		}

	} else {
		log.Println("Found match")

		if err := fd.db.First(&faceGroup, int(match)).Error; err != nil {
			return err
		}

		if err := fd.db.Model(&faceGroup).Association("ImageFaces").Append(&imageFace); err != nil {
			return err
		}
	}

	fd.samples = append(fd.samples, face.Descriptor)
	fd.cats = append(fd.cats, int32(faceGroup.ID))

	fd.rec.SetSamples(fd.samples, fd.cats)
	return nil
}
