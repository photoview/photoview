package exif

import (
	"log"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/photoview/photoview/api/graphql/models"
)

type externalExifParser struct{}

func (p *externalExifParser) ParseExif(media_path string) (returnExif *models.MediaEXIF, returnErr error) {
	// ExifTool - No print conversion mode
	et, err := exiftool.NewExiftool(exiftool.NoPrintConversion())
	if err != nil {
		log.Printf("Error initializing ExifTool: %s\n", err)
		return nil, err
	}
	defer et.Close()

	fileInfos := et.ExtractMetadata(media_path)
	newExif := models.MediaEXIF{}
	found_exif := false

	for _, fileInfo := range fileInfos {
		if fileInfo.Err != nil {
			log.Printf("Fileinfo error: %v\n", fileInfo.Err)
			continue
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
		date, err := fileInfo.GetString("DateTimeOriginal")
		if err == nil {
			layout := "2006:01:02 15:04:05"
			dateTime, err := time.Parse(layout, date)
			if err == nil {
				found_exif = true
				newExif.DateShot = &dateTime
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
	}

	if !found_exif {
		return nil, nil
	}

	returnExif = &newExif
	return
}
