package face_detection

import (
	"log"
	"sync"
	"strconv"
	"math/rand"
  "time"
	// "os"
	// "encoding/csv"
	"fmt"
	"encoding/json"

	"github.com/Kagami/go-face"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/PJ-Watson/clusters"
)

type FaceDetector struct {
	mutex           sync.Mutex
	rec             *face.Recognizer
	faceDescriptors []face.Descriptor
	faceGroupIDs    []int32
	imageFaceIDs    []int
}

var GlobalFaceDetector *FaceDetector = nil

func InitializeFaceDetector(db *gorm.DB) error {
	if utils.EnvDisableFaceRecognition.GetBool() {
		log.Printf("Face detection disabled (%s=1)\n", utils.EnvDisableFaceRecognition.GetName())
		return nil
	}

	log.Println("Initializing face detector")

	var rec *face.Recognizer
	var error error

	if utils.EnvAdvancedFaceRecognition.GetBool() {
		minSize, err := strconv.Atoi(utils.EnvFaceMinSize.GetValueWithDefault("150"))
		if err != nil {
			return errors.Wrap(err, "invalid minimum face size")
		}
		padding, err := strconv.ParseFloat(utils.EnvFacePadding.GetValueWithDefault("0.25"), 32)
		if err != nil {
			return errors.Wrap(err, "invalid value for face padding")
		}
		jittering, err := strconv.Atoi(utils.EnvFaceJittering.GetValueWithDefault("0"))
		if err != nil {
			return errors.Wrap(err, "invalid value for jittering")
		}
		rec, error = face.NewRecognizerWithConfig(utils.FaceRecognitionModelsPath(), minSize, float32(padding), jittering)
		log.Println("Using face detector with config")
	} else {
		rec, error = face.NewRecognizer(utils.FaceRecognitionModelsPath())
	}

	if error != nil {
		return errors.Wrap(error, "initialize facedetect recognizer")
	}

	faceDescriptors, faceGroupIDs, imageFaceIDs, err := getSamplesFromDatabase(db)
	if err != nil {
		return errors.Wrap(err, "get face detection samples from database")
	}

	GlobalFaceDetector = &FaceDetector{
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
func (fd *FaceDetector) ReloadFacesFromDatabase(db *gorm.DB) error {
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
func (fd *FaceDetector) DetectFaces(db *gorm.DB, media *models.Media) error {
	if err := db.Model(media).Preload("MediaURL").First(&media).Error; err != nil {
		return err
	}

	var thumbnailURL *models.MediaURL

	if (utils.EnvAdvancedFaceRecognition.GetBool()) && (utils.EnvFaceRecUseLargest.GetBool()) {
		log.Printf("this works")
		for _, url := range media.MediaURL {
			if (url.Purpose == models.MediaOriginal) && (url.ContentType=="image/jpeg") {
				thumbnailURL = &url
				thumbnailURL.Media = media
				break
			}
		}
		if thumbnailURL == nil {
			for _, url := range media.MediaURL {
				if url.Purpose == models.PhotoHighRes {
					thumbnailURL = &url
					thumbnailURL.Media = media
					break
				}
			}
			if thumbnailURL == nil {
				for _, url := range media.MediaURL {
					if url.Purpose == models.PhotoThumbnail {
						thumbnailURL = &url
						thumbnailURL.Media = media
						break
					}
				}
			}
		}
	} else {
		for _, url := range media.MediaURL {
			if url.Purpose == models.PhotoThumbnail {
				thumbnailURL = &url
				thumbnailURL.Media = media
				break
			}
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
		fd.classifyFace(db, &face, media, thumbnailPath)
	}

	// fd.classifyFaces(db, &faces, media, thumbnailPath)

	return nil
}

func (fd *FaceDetector) classifyDescriptor(descriptor face.Descriptor) int32 {
	return int32(fd.rec.ClassifyThreshold(descriptor, 0.3))
}

func (fd *FaceDetector) classifyFace(db *gorm.DB, face *face.Face, media *models.Media, imagePath string) error {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	match := fd.classifyDescriptor(face.Descriptor)

	faceRect, err := models.ToDBFaceRectangle(face.Rectangle, imagePath)
	if err != nil {
		return err
	}

	log.Println(face.Descriptor)

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

		fd.imageFaceIDs = append(fd.imageFaceIDs, faceGroup.ImageFaces[0].ID)

	} else {
		log.Println("Found match")

		if err := db.First(&faceGroup, int(match)).Error; err != nil {
			return err
		}

		if err := db.Model(&faceGroup).Association("ImageFaces").Append(&imageFace); err != nil {
			return err
		}

		fd.imageFaceIDs = append(fd.imageFaceIDs, imageFace.ID)
	}

	fd.faceDescriptors = append(fd.faceDescriptors, face.Descriptor)
	fd.faceGroupIDs = append(fd.faceGroupIDs, int32(faceGroup.ID))

	fd.rec.SetSamples(fd.faceDescriptors, fd.faceGroupIDs)
	return nil
}

func (fd *FaceDetector) MergeCategories(sourceID int32, destID int32) {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	for i := range fd.faceGroupIDs {
		if fd.faceGroupIDs[i] == sourceID {
			fd.faceGroupIDs[i] = destID
		}
	}
}

func (fd *FaceDetector) MergeImageFaces(imageFaceIDs []int, destFaceGroupID int32) {
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

// type f64Descriptor [128]float64

func (fd *FaceDetector) RecognizeUnlabeledFaces(tx *gorm.DB, user *models.User) ([]*models.ImageFace, error) {

	var unrecognizedDescriptorsF64 [][]float64
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

	log.Println("no errors 1")
	log.Println(unlabeledFaceGroups)

	if err != nil {
		return nil, err
	}

	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	log.Println(fd.imageFaceIDs)

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

			// Having to convert types is annoying - will see if clusters or go-face can be changed

			var recastDescriptor []float64

			for i := range descriptor {
				recastDescriptor = append(recastDescriptor, float64(descriptor[i]))
			}

			unrecognizedDescriptorsF64 = append(unrecognizedDescriptorsF64, recastDescriptor)
			unrecognizedFaceGroupIDs = append(unrecognizedFaceGroupIDs, faceGroupID)
			unrecognizedDescriptors = append(unrecognizedDescriptors, descriptor)
			unrecognizedImageFaceIDs = append(unrecognizedImageFaceIDs, imageFaceID)
		} else {
			newFaceGroupIDs = append(newFaceGroupIDs, faceGroupID)
			newDescriptors = append(newDescriptors, descriptor)
			newImageFaceIDs = append(newImageFaceIDs, imageFaceID)
		}
	}

	log.Println(unrecognizedFaceGroupIDs)
	log.Println("no errors 2")

	fd.faceGroupIDs = newFaceGroupIDs
	fd.faceDescriptors = newDescriptors
	fd.imageFaceIDs = newImageFaceIDs

	log.Println(newDescriptors)
	log.Println("")

	updatedImageFaces := make([]*models.ImageFace, 0)

	if len(unrecognizedDescriptorsF64)==0 {
		return updatedImageFaces, nil
	}

	// var f64strSlice [][]string
	// for _, a := range unrecognizedDescriptorsF64{
	// 	var tmp []string
	// 	for _, b := range a {
	// 		tmp = append(tmp, fmt.Sprint(b))
	// 	}
	// 	f64strSlice = append(f64strSlice, tmp)
	// }
	//
	// file, err := os.Create("/home/peter/Downloads/result.csv")
	// if err != nil {
	// 	return updatedImageFaces, err
	// }
	// defer file.Close()
	//
	// writer := csv.NewWriter(file)
	// defer writer.Flush()
	//
	// for _, value := range f64strSlice {
	// 	if err := writer.Write(value); err != nil {
	// 		return updatedImageFaces, err
	// 	}
	// }
	//
	// log.Println("written to csv?")

	var c clusters.HardClusterer

	if c, err = clusters.DBSCAN(3, 0.41, 4, clusters.EuclideanDistance); err != nil {
		return updatedImageFaces, err
	}

	if err = c.Learn(unrecognizedDescriptorsF64); err != nil {
		return updatedImageFaces, err
	}

	log.Println("no errors 3")

	clusterSizes := c.Sizes()

	log.Println(clusterSizes)

	clusterAssignments := c.Guesses()

	log.Println(clusterAssignments)

	// unclusteredIdx := sliceIndicesInt(matches, 2)
	//
	// log.Println(unclusteredIdx)

	if unclusteredIdx := sliceIndicesInt(clusterAssignments, -1); len(unclusteredIdx)>0 {

		log.Println("Reassigning unclustered faces")
		log.Println(unclusteredIdx)
		log.Println(unrecognizedImageFaceIDs)

		for _, idx := range unclusteredIdx {
			// descriptor := unrecognizedDescriptors[i]
			// faceGroupID := unrecognizedFaceGroupIDs[i]
			// imageFaceID := unrecognizedImageFaceIDs[i]

			// fd.faceGroupIDs = append(fd.faceGroupIDs, unrecognizedFaceGroupIDs[idx])
			// fd.faceDescriptors = append(fd.faceDescriptors, unrecognizedDescriptors[idx])
			// fd.imageFaceIDs = append(fd.imageFaceIDs, unrecognizedImageFaceIDs[idx])

			// userOwnedImageFaceIDs := make([]int, 0)
			var newFaceGroup models.FaceGroup
			var retImgFace []models.ImageFace
			if err := tx.Find(&retImgFace, unrecognizedImageFaceIDs[idx]).Error; err != nil {
				return updatedImageFaces, err
			}

			log.Println("This part was successful")
			// err := tx.Model(models.ImageFace).Where("id IN (?)", unrecognizedImageFaceIDs[idx])

			// newLabel := "New"

			newFaceGroup = models.FaceGroup{
				// Label: &newLabel,
				ImageFaces: retImgFace,
			}

			if err := tx.Create(&newFaceGroup).Error; err != nil {
				return updatedImageFaces,err
			}

			fd.faceGroupIDs = append(fd.faceGroupIDs, int32(newFaceGroup.ID))
			fd.faceDescriptors = append(fd.faceDescriptors, unrecognizedDescriptors[idx])
			fd.imageFaceIDs = append(fd.imageFaceIDs, unrecognizedImageFaceIDs[idx])


			log.Println("Made it to here")

			// newFaceGroup := models.FaceGroup{}
			//
			// transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
			//
			// 	if err := tx.Save(&newFaceGroup).Error; err != nil {
			// 		return err
			// 	}
			//
			// 	if err := tx.
			// 		Model(&models.ImageFace{}).
			// 		Where("id IN (?)", userOwnedImageFaceIDs).
			// 		Update("face_group_id", newFaceGroup.ID).Error; err != nil {
			// 		return err
			// 	}
			//
			// 	return nil
			// })
			//
			// if transactionError != nil {
			// 	return nil, transactionError
			// }

		}
	}

	if len(clusterSizes) > 0 {
		for i, size := range clusterSizes {
			log.Printf("%d", i)
			log.Printf("%d", size)

			clusteredIdx := sliceIndicesInt(clusterAssignments, i+1)

			log.Println(clusteredIdx)

			sampleNum := 3

			randSelectIdx := randomUniqueSlice(clusteredIdx, sampleNum)

			log.Println("Random indices")
			log.Println(randSelectIdx)

			var faceMatches []int32
			for _, r := range randSelectIdx {
				faceMatches = append(faceMatches, fd.classifyDescriptor(unrecognizedDescriptors[r]))
			}

			faceMatchID := int32(-1)
			for i := 0; i < sampleNum - 1; i++ {
		    for j := i + 1; j < sampleNum; j++ {
	        if faceMatches[i] == faceMatches[j] {
			    	faceMatchID = faceMatches[i]
	        }
	    	}
			}

			log.Println("All good at this point")
			log.Println(faceMatches)

			if faceMatchID < 0 {

				var newFaceGroup models.FaceGroup
				// newLabel := "New_cluster"

				newFaceGroup = models.FaceGroup{
					// Label: &newLabel,
					// ImageFaces: retImgFace,
				}


				log.Println("new face")

				if err := tx.Create(&newFaceGroup).Error; err != nil {
					return updatedImageFaces,err
				}

				log.Println("new face 2")

				s, _ := json.MarshalIndent(&newFaceGroup, "", "\t")
				fmt.Print(string(s))

				for _, idx := range clusteredIdx {

					log.Println("important info")
					log.Println(clusteredIdx)
					log.Println(idx)
					log.Println(unrecognizedImageFaceIDs)

					var retImgFace models.ImageFace
					if err := tx.Find(&retImgFace, unrecognizedImageFaceIDs[idx]).Error; err != nil {
						return updatedImageFaces, err
					}

					s, _ := json.MarshalIndent(&retImgFace, "", "\t")
					fmt.Print(string(s))

					log.Println("new face 3")

					if err := tx.Model(&newFaceGroup).Association("ImageFaces").Append(&retImgFace); err != nil {
						return updatedImageFaces, err
					}

					updatedImageFaces = append(updatedImageFaces, &retImgFace)

					fd.faceGroupIDs = append(fd.faceGroupIDs, int32(newFaceGroup.ID))
					fd.faceDescriptors = append(fd.faceDescriptors, unrecognizedDescriptors[idx])
					fd.imageFaceIDs = append(fd.imageFaceIDs, unrecognizedImageFaceIDs[idx])


					log.Println("new face 4")
				}
			} else {

				var faceGroup models.FaceGroup

				log.Println("new face 5")

				if err := tx.First(&faceGroup, int(faceMatchID)).Error; err != nil {
					return updatedImageFaces, err
				}

				log.Println("new face 6")
				for _, idx := range clusteredIdx {

					var retImgFace models.ImageFace
					if err := tx.Find(&retImgFace, unrecognizedImageFaceIDs[idx]).Error; err != nil {
						return updatedImageFaces, err
					}
					log.Println("new face 7")

					if err := tx.Model(&faceGroup).Association("ImageFaces").Append(&retImgFace); err != nil {
						return updatedImageFaces, err
					}
					log.Println("new face 8")

					updatedImageFaces = append(updatedImageFaces, &retImgFace)

					fd.faceGroupIDs = append(fd.faceGroupIDs, int32(faceMatchID))
					fd.faceDescriptors = append(fd.faceDescriptors, unrecognizedDescriptors[idx])
					fd.imageFaceIDs = append(fd.imageFaceIDs, unrecognizedImageFaceIDs[idx])
					log.Println("new face 9")
				}
			}



		// for i := 1; i < (len(sizes)+1); i++ {
		// 	size := sizes[i]
		// 	log.Printf("%d", i)
		// 	log.Printf("%d", size)
		}
	}

	// for i, size := range sizes {
	//
	//
	// 	log.Printf("%d", i)
	// 	log.Printf("%d", size)
	// }



	// 	descriptor := unrecognizedDescriptors[i]
	// 	faceGroupID := unrecognizedFaceGroupIDs[i]
	// 	imageFaceID := unrecognizedImageFaceIDs[i]
	//
	// 	// match := fd.classifyDescriptor(descriptor)
	// 	// match := guesses[i]
	//
	// 	if match < 0 {
	// 		// still no match, we can readd it to the list
	// 		fd.faceGroupIDs = append(fd.faceGroupIDs, faceGroupID)
	// 		fd.faceDescriptors = append(fd.faceDescriptors, descriptor)
	// 		fd.imageFaceIDs = append(fd.imageFaceIDs, imageFaceID)
	// 	} else {
	// 		// found new match, update the database
	// 		var imageFace models.ImageFace
	// 		if err := tx.Model(&models.ImageFace{}).First(imageFace, imageFaceID).Error; err != nil {
	// 			return nil, err
	// 		}
	//
	// 		if err := tx.Model(&imageFace).Update("face_group_id", int(faceGroupID)).Error; err != nil {
	// 			return nil, err
	// 		}
	//
	// 		updatedImageFaces = append(updatedImageFaces, &imageFace)
	//
	// 		fd.faceGroupIDs = append(fd.faceGroupIDs, int32(match))
	// 		fd.faceDescriptors = append(fd.faceDescriptors, descriptor)
	// 		fd.imageFaceIDs = append(fd.imageFaceIDs, imageFaceID)
	// 	}
	// }

	// for i, match := range matches {
	// 	descriptor := unrecognizedDescriptors[i]
	// 	faceGroupID := unrecognizedFaceGroupIDs[i]
	// 	imageFaceID := unrecognizedImageFaceIDs[i]
	//
	// 	// match := fd.classifyDescriptor(descriptor)
	// 	// match := guesses[i]
	//
	// 	if match < 0 {
	// 		// still no match, we can readd it to the list
	// 		fd.faceGroupIDs = append(fd.faceGroupIDs, faceGroupID)
	// 		fd.faceDescriptors = append(fd.faceDescriptors, descriptor)
	// 		fd.imageFaceIDs = append(fd.imageFaceIDs, imageFaceID)
	// 	} else {
	// 		// found new match, update the database
	// 		var imageFace models.ImageFace
	// 		if err := tx.Model(&models.ImageFace{}).First(imageFace, imageFaceID).Error; err != nil {
	// 			return nil, err
	// 		}
	//
	// 		if err := tx.Model(&imageFace).Update("face_group_id", int(faceGroupID)).Error; err != nil {
	// 			return nil, err
	// 		}
	//
	// 		updatedImageFaces = append(updatedImageFaces, &imageFace)
	//
	// 		fd.faceGroupIDs = append(fd.faceGroupIDs, int32(match))
	// 		fd.faceDescriptors = append(fd.faceDescriptors, descriptor)
	// 		fd.imageFaceIDs = append(fd.imageFaceIDs, imageFaceID)
	// 	}
	// }

	return updatedImageFaces, nil
}

// God, I miss having np.where
func sliceIndicesInt(inputSlice []int, value int) []int {
	var indices []int
	for i, v := range inputSlice{
		if v == value {
			indices = append(indices, i)
		}
	}
	return indices
}

func randomUniqueSlice(inputSlice []int, outputLength int) []int {
	rand.Seed(time.Now().Unix())
	var result []int
	inputLength := len(inputSlice)
  p := rand.Perm(inputLength)
  for _, r := range p[:outputLength] {
    result = append(result, inputSlice[r])
  }
	return result
}

// func (fd *FaceDetector) clusterFaces(db *gorm.DB, newFaces [], newMedia *models.Media, imagePaths string) error {
// 	fd.mutex.Lock()
// 	defer fd.mutex.Unlock()
//
// 	var c clusters.HardClusterer
// 	var err error
//
// 	if c, err = clusters.DBSCAN(3, 0.4, 4, clusters.EuclideanDistance); err != nil {
// 		return err
// 	} else if err = c.Learn(newFaces); err != nil {
// 		return err
// 	}
//
// }
