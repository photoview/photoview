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
		fd.classifyFace(&face, media, thumbnailPath)
	}

	return nil
}

func (fd *FaceDetector) classifyDescriptor(descriptor face.Descriptor) int32 {
	return int32(fd.rec.ClassifyThreshold(descriptor, 0.3))
}

func (fd *FaceDetector) classifyFace(face *face.Face, media *models.Media, imagePath string) error {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	match := fd.classifyDescriptor(face.Descriptor)

	faceRect, err := models.ToDBFaceRectangle(face.Rectangle, imagePath)
	if err != nil {
		return err
	}

	imageFace := models.ImageFace{
		MediaID:    media.ID,
		Descriptor: models.FaceDescriptor(face.Descriptor),
		Rectangle:  *faceRect,
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

func (fd *FaceDetector) MergeCategories(sourceID int32, destID int32) {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	for i := range fd.cats {
		if fd.cats[i] == sourceID {
			fd.cats[i] = destID
		}
	}
}

func (fd *FaceDetector) RecognizeUnlabeledFaces(tx *gorm.DB, user *models.User) ([]*models.ImageFace, error) {
	unrecognizedSamples := make([]face.Descriptor, 0)
	unrecognizedCats := make([]int32, 0)

	newCats := make([]int32, 0)
	newSamples := make([]face.Descriptor, 0)

	var unlabeledFaceGroups []*models.FaceGroup

	err := tx.
		Joins("JOIN image_faces ON image_faces.face_group_id = face_groups.id").
		Joins("JOIN media ON image_faces.media_id = media.id").
		Where("face_groups.label IS NULL").
		Where("media.album_id IN (?)",
			tx.Select("album_id").Table("user_albums").Where("user_id = ?", user.ID),
		).
		Find(&unlabeledFaceGroups).Error

	if err != nil {
		return nil, err
	}

	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	for i := range fd.samples {
		cat := fd.cats[i]
		sample := fd.samples[i]

		catIsUnlabeled := false
		for _, unlabeledFaceGroup := range unlabeledFaceGroups {
			if cat == int32(unlabeledFaceGroup.ID) {
				catIsUnlabeled = true
				continue
			}
		}

		if catIsUnlabeled {
			unrecognizedCats = append(unrecognizedCats, cat)
			unrecognizedSamples = append(unrecognizedSamples, sample)
		} else {
			newCats = append(newCats, cat)
			newSamples = append(newSamples, sample)
		}
	}

	fd.cats = newCats
	fd.samples = newSamples

	updatedImageFaces := make([]*models.ImageFace, 0)

	for i := range unrecognizedSamples {
		cat := unrecognizedCats[i]
		sample := unrecognizedSamples[i]

		match := fd.classifyDescriptor(sample)

		if match < 0 {
			// still no match, we can readd it to the list
			fd.cats = append(fd.cats, cat)
			fd.samples = append(fd.samples, sample)
		} else {
			// found new match, update the database
			var imageFace models.ImageFace
			if err := tx.Model(&models.ImageFace{
				Descriptor: models.FaceDescriptor(sample),
			}).First(imageFace).Error; err != nil {
				return nil, err
			}

			if err := tx.Model(&imageFace).Update("face_group_id", int(cat)).Error; err != nil {
				return nil, err
			}

			updatedImageFaces = append(updatedImageFaces, &imageFace)

			fd.cats = append(fd.cats, match)
			fd.samples = append(fd.samples, sample)
		}
	}

	return updatedImageFaces, nil
}
