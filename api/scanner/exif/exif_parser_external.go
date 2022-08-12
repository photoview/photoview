package exif

import (
	"log"
	"math"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/photoview/photoview/api/dataloader"
	"github.com/photoview/photoview/api/graphql/models"
)

type externalExifParser struct {
	et         *exiftool.Exiftool
	dataLoader *dataloader.ExiftoolLoader
}

func NewExiftoolParser() (ExifParser, error) {
	et, err := exiftool.NewExiftool(exiftool.NoPrintConversion())

	if err != nil {
		log.Printf("Error initializing ExifTool: %s\n", err)
		return nil, err
	}

	return &externalExifParser{
		et:         et,
		dataLoader: dataloader.NewExiftoolLoader(et),
	}, nil
}

// isFloatReal returns true when the float value represents a real number
// (different than +Inf, -Inf or NaN)
func isFloatReal(v float64) bool {
	if math.IsInf(v, 1) {
		return false
	} else if math.IsInf(v, -1) {
		return false
	} else if math.IsNaN(v) {
		return false
	}
	return true
}

// sanitizeEXIF removes any EXIF float64 field that is not a real number (+Inf,
// -Inf or Nan)
func sanitizeEXIF(exif *models.MediaEXIF) {
	if exif.Exposure != nil && !isFloatReal(*exif.Exposure) {
		exif.Exposure = nil
	}
	if exif.Aperture != nil && !isFloatReal(*exif.Aperture) {
		exif.Aperture = nil
	}
	if exif.FocalLength != nil && !isFloatReal(*exif.FocalLength) {
		exif.FocalLength = nil
	}
	if (exif.GPSLatitude != nil && !isFloatReal(*exif.GPSLatitude)) ||
		(exif.GPSLongitude != nil && !isFloatReal(*exif.GPSLongitude)) {
		exif.GPSLatitude = nil
		exif.GPSLongitude = nil
	}
}

func (p *externalExifParser) ParseExif(media_path string) (returnExif *models.MediaEXIF, returnErr error) {
	// ExifTool - No print conversion mode
	if p.et == nil {
		et, err := exiftool.NewExiftool(exiftool.NoPrintConversion())
		p.et = et

		if err != nil {
			log.Printf("Error initializing ExifTool: %s\n", err)
			return nil, err
		}
	}

	fileInfo, err := p.dataLoader.Load(media_path)
	if err != nil {
		return nil, err
	}

	newExif := models.MediaEXIF{}
	found_exif := false

	// Get description
	description, err := fileInfo.GetString("ImageDescription")
	if err == nil {
		found_exif = true
		newExif.Description = &description
	}

	// Get camera model
	model, err := fileInfo.GetString("Model")
	if err == nil {
		found_exif = true
		newExif.Camera = &model
	}

	// Get Camera make
	make, err := fileInfo.GetString("Make")
	if err == nil {
		found_exif = true
		newExif.Maker = &make
	}

	// Get lens
	lens, err := fileInfo.GetString("LensModel")
	if err == nil {
		found_exif = true
		newExif.Lens = &lens
	}

	//Get time of photo
	createDateKeys := []string{"DateTimeOriginal", "CreateDate", "TrackCreateDate", "MediaCreateDate", "FileCreateDate", "ModifyDate", "TrackModifyDate", "MediaModifyDate", "FileModifyDate"}
	for _, createDateKey := range createDateKeys {
		date, err := fileInfo.GetString(createDateKey)
		if err == nil {
			layout := "2006:01:02 15:04:05"
			dateTime, err := time.Parse(layout, date)
			if err == nil {
				found_exif = true
				newExif.DateShot = &dateTime
			}
			break
		}
	}

	// Get exposure time
	exposureTime, err := fileInfo.GetFloat("ExposureTime")
	if err == nil {
		found_exif = true
		newExif.Exposure = &exposureTime
	}

	// Get aperture
	aperture, err := fileInfo.GetFloat("Aperture")
	if err == nil {
		found_exif = true
		newExif.Aperture = &aperture
	}

	// Get ISO
	iso, err := fileInfo.GetInt("ISO")
	if err == nil {
		found_exif = true
		newExif.Iso = &iso
	}

	// Get focal length
	focalLen, err := fileInfo.GetFloat("FocalLength")
	if err == nil {
		found_exif = true
		newExif.FocalLength = &focalLen
	}

	// Get flash info
	flash, err := fileInfo.GetInt("Flash")
	if err == nil {
		found_exif = true
		newExif.Flash = &flash
	}

	// Get orientation
	orientation, err := fileInfo.GetInt("Orientation")
	if err == nil {
		found_exif = true
		newExif.Orientation = &orientation
	}

	// Get exposure program
	expProgram, err := fileInfo.GetInt("ExposureProgram")
	if err == nil {
		found_exif = true
		newExif.ExposureProgram = &expProgram
	}

	// GPS coordinates - longitude
	longitudeRaw, err := fileInfo.GetFloat("GPSLongitude")
	if err == nil {
		found_exif = true
		newExif.GPSLongitude = &longitudeRaw
	}

	// GPS coordinates - latitude
	latitudeRaw, err := fileInfo.GetFloat("GPSLatitude")
	if err == nil {
		found_exif = true
		newExif.GPSLatitude = &latitudeRaw
	}

	if !found_exif {
		return nil, nil
	}

	returnExif = &newExif
	sanitizeEXIF(returnExif)
	return
}
