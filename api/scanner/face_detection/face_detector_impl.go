//go:build !no_face_detection

package face_detection

import (
	"log"
	"sync"

	"github.com/Kagami/go-face"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type faceDetector struct {
	mutex           sync.Mutex
	rec             *face.Recognizer
	faceDescriptors []face.Descriptor
	faceGroupIDs    []int32
	imageFaceIDs    []int
}

func InitializeFaceDetector(db *gorm.DB) error {
	if utils.EnvDisableFaceRecognition.GetBool() {
		log.Printf("Face detection disabled (%s=1)\n", utils.EnvDisableFaceRecognition.GetName())
		return nil
	}

	log.Println("Initializing face detector")

	rec, err := face.NewRecognizer(utils.FaceRecognitionModelsPath())
	if err != nil {
		return errors.Wrap(err, "initialize facedetect recognizer")
	}

	faceDescriptors, faceGroupIDs, imageFaceIDs, err := getSamplesFromDatabase(db)
	if err != nil {
		return errors.Wrap(err, "get face detection samples from database")
	}

	GlobalFaceDetector = &faceDetector{
		rec:             rec,
		faceDescriptors: faceDescriptors,
		faceGroupIDs:    faceGroupIDs,
		imageFaceIDs:    imageFaceIDs,
	}

	return nil
}

func getSamplesFromDatabase(db *gorm.DB) (samples []face.Descriptor, faceGroupIDs []int32, imageFaceIDs []int, err error) {

	var imageFaces []*models.ImageFace

	if err = db.Find(&imageFaces).Error; err != nil {
		return
	}

	samples = make([]face.Descriptor, len(imageFaces))
	faceGroupIDs = make([]int32, len(imageFaces))
	imageFaceIDs = make([]int, len(imageFaces))

	for i, imgFace := range imageFaces {
		samples[i] = face.Descriptor(imgFace.Descriptor)
		faceGroupIDs[i] = int32(imgFace.FaceGroupID)
		imageFaceIDs[i] = imgFace.ID
	}

	return
}

// ReloadFacesFromDatabase replaces the in-memory face descriptors with the ones in the database
func (fd *faceDetector) ReloadFacesFromDatabase(db *gorm.DB) error {
	faceDescriptors, faceGroupIDs, imageFaceIDs, err := getSamplesFromDatabase(db)
	if err != nil {
		return err
	}

	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	fd.faceDescriptors = faceDescriptors
	fd.faceGroupIDs = faceGroupIDs
	fd.imageFaceIDs = imageFaceIDs

	return nil
}

// DetectFaces finds the faces in the given image and saves them to the database
func (fd *faceDetector) DetectFaces(db *gorm.DB, media *models.Media, isReDetection bool) error {
	if err := db.Model(media).Preload("MediaURL").First(&media).Error; err != nil {
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
		fd.classifyFace(db, &face, media, thumbnailPath, isReDetection)
	}

	return nil
}

func (fd *faceDetector) classifyDescriptor(descriptor face.Descriptor) int32 {
	return int32(fd.rec.ClassifyThreshold(descriptor, 0.2))
}

func (fd *faceDetector) classifyFace(db *gorm.DB, face *face.Face, media *models.Media, imagePath string, isReDetection bool) error {
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

		if err := db.Create(&faceGroup).Error; err != nil {
			return err
		}
		log.Printf("Created FaceGroup with ID: %d", faceGroup.ID)

	} else {
		log.Println("Found match")

		if err := db.First(&faceGroup, int(match)).Error; err != nil {
			return err
		}

		if isReDetection {
			// Check if an image face with the same media_id and face_group_id already exists in case of re-detection
			var existingImageFace models.ImageFace
			if err := db.Where("media_id = ? AND face_group_id = ?", media.ID, faceGroup.ID).First(&existingImageFace).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
			}
			if existingImageFace.ID > 0 {
				return nil
			}
		}

		if err := db.Model(&faceGroup).Association("ImageFaces").Append(&imageFace); err != nil {
			return err
		}
	}

	fd.faceDescriptors = append(fd.faceDescriptors, face.Descriptor)
	fd.faceGroupIDs = append(fd.faceGroupIDs, int32(faceGroup.ID))
	fd.imageFaceIDs = append(fd.imageFaceIDs, imageFace.ID)

	fd.rec.SetSamples(fd.faceDescriptors, fd.faceGroupIDs)
	return nil
}

func (fd *faceDetector) MergeCategories(sourceID int32, destID int32) {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	for i := range fd.faceGroupIDs {
		if fd.faceGroupIDs[i] == sourceID {
			fd.faceGroupIDs[i] = destID
		}
	}
}

func (fd *faceDetector) MergeImageFaces(imageFaceIDs []int, destFaceGroupID int32) {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	for i := range fd.faceGroupIDs {
		imageFaceID := fd.imageFaceIDs[i]

		for _, id := range imageFaceIDs {
			if imageFaceID == id {
				fd.faceGroupIDs[i] = destFaceGroupID
				break
			}
		}
	}
}

func (fd *faceDetector) RecognizeUnlabeledFaces(tx *gorm.DB, user *models.User) ([]*models.ImageFace, error) {
	unrecognizedDescriptors := make([]face.Descriptor, 0)
	unrecognizedFaceGroupIDs := make([]int32, 0)
	unrecognizedImageFaceIDs := make([]int, 0)

	newFaceGroupIDs := make([]int32, 0)
	newDescriptors := make([]face.Descriptor, 0)
	newImageFaceIDs := make([]int, 0)

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

	for i := range fd.faceDescriptors {
		descriptor := fd.faceDescriptors[i]
		faceGroupID := fd.faceGroupIDs[i]
		imageFaceID := fd.imageFaceIDs[i]

		isUnlabeled := false
		for _, unlabeledFaceGroup := range unlabeledFaceGroups {
			if faceGroupID == int32(unlabeledFaceGroup.ID) {
				isUnlabeled = true
				continue
			}
		}

		if isUnlabeled {
			unrecognizedFaceGroupIDs = append(unrecognizedFaceGroupIDs, faceGroupID)
			unrecognizedDescriptors = append(unrecognizedDescriptors, descriptor)
			unrecognizedImageFaceIDs = append(unrecognizedImageFaceIDs, imageFaceID)
		} else {
			newFaceGroupIDs = append(newFaceGroupIDs, faceGroupID)
			newDescriptors = append(newDescriptors, descriptor)
			newImageFaceIDs = append(newImageFaceIDs, imageFaceID)
		}
	}

	fd.faceGroupIDs = newFaceGroupIDs
	fd.faceDescriptors = newDescriptors
	fd.imageFaceIDs = newImageFaceIDs

	updatedImageFaces := make([]*models.ImageFace, 0)

	for i := range unrecognizedDescriptors {
		descriptor := unrecognizedDescriptors[i]
		faceGroupID := unrecognizedFaceGroupIDs[i]
		imageFaceID := unrecognizedImageFaceIDs[i]

		match := fd.classifyDescriptor(descriptor)

		if match < 0 {
			// still no match, we can readd it to the list
			fd.faceGroupIDs = append(fd.faceGroupIDs, faceGroupID)
			fd.faceDescriptors = append(fd.faceDescriptors, descriptor)
			fd.imageFaceIDs = append(fd.imageFaceIDs, imageFaceID)
		} else {
			// found new match, update the database
			var imageFace models.ImageFace
			if err := tx.Model(&models.ImageFace{}).First(imageFace, imageFaceID).Error; err != nil {
				return nil, err
			}

			if err := tx.Model(&imageFace).Update("face_group_id", int(faceGroupID)).Error; err != nil {
				return nil, err
			}

			updatedImageFaces = append(updatedImageFaces, &imageFace)

			fd.faceGroupIDs = append(fd.faceGroupIDs, match)
			fd.faceDescriptors = append(fd.faceDescriptors, descriptor)
			fd.imageFaceIDs = append(fd.imageFaceIDs, imageFaceID)
		}
	}

	return updatedImageFaces, nil
}
func (fd *faceDetector) ReDetectFaces(db *gorm.DB, media *models.Media) error {
	err := db.Transaction(func(tx *gorm.DB) error {

		//Delete all image_faces with the given media_id where the face_group has no label
		var faceGroupIDs []int
		if err := tx.Model(&models.ImageFace{}).
			Joins("JOIN face_groups ON face_groups.id = image_faces.face_group_id").
			Where("image_faces.media_id = ? AND face_groups.label IS NULL", media.ID).
			Pluck("image_faces.face_group_id", &faceGroupIDs).Error; err != nil {
			return err
		}

		if err := tx.Where("media_id = ? AND face_group_id IN (?)", media.ID, faceGroupIDs).
			Delete(&models.ImageFace{}).Error; err != nil {
			return err
		}

		//Check and delete face_groups with no associated image_faces
		if len(faceGroupIDs) > 0 {
			if err := tx.Where("id IN (?) AND NOT EXISTS (SELECT 1 FROM image_faces WHERE face_group_id = face_groups.id)", faceGroupIDs).
				Delete(&models.FaceGroup{}).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	if err := fd.ReloadFacesFromDatabase(db); err != nil {
		return err
	}

	fd.setSamplesWithFallback(fd.faceDescriptors, fd.faceGroupIDs)

	if err := fd.DetectFaces(db, media, true); err != nil {
		return err
	}

	return nil
}

func (fd *faceDetector) setSamplesWithFallback(samples []face.Descriptor, cats []int32) {
	if len(samples) > 0 {
		fd.rec.SetSamples(samples, cats)
	} else {
		dummyDescriptor := face.Descriptor{}
		dummyCategory := int32(0)
		fd.rec.SetSamples([]face.Descriptor{dummyDescriptor}, []int32{dummyCategory})
	}
}
