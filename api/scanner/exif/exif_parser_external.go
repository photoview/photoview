package exif

import (
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/photoview/photoview/api/graphql/models"
)

type externalExifParser struct{}

func (p *externalExifParser) ParseExif(media *models.Media) (returnExif *models.MediaEXIF, returnErr error) {
	// ExifTool - No print conversion mode
	et, err := exiftool.NewExiftool(exiftool.NoPrintConversion())
	if err != nil {
		log.Printf("Error initializing ExifTool: %s\n", err)
		return nil, err
	}
	defer et.Close()

	fileInfos := et.ExtractMetadata(media.Path)
	newExif := models.MediaEXIF{}

	for _, fileInfo := range fileInfos {
		if fileInfo.Err != nil {
			log.Printf("Fileinfo error\n")
			continue
		}

		// Get camera model
		model, err := fileInfo.GetString("Model")
		if err == nil {
			log.Printf("Camera model: %v", model)
			newExif.Camera = &model
		}

		// Get Camera make
		make, err := fileInfo.GetString("Make")
		if err == nil {
			log.Printf("Camera make: %v", make)
			newExif.Maker = &make
		}

		// Get lens
		lens, err := fileInfo.GetString("LensModel")
		if err == nil {
			log.Printf("Lens: %v", lens)
			newExif.Lens = &lens
		}

		//Get time of photo
		date, err := fileInfo.GetString("DateTimeOriginal")
		if err == nil {
			log.Printf("Date shot: %s", date)
			layout := "2006:01:02 15:04:05"
			dateTime, err := time.Parse(layout, date)
			if err == nil {
				newExif.DateShot = &dateTime
			}
		}

		// Get exposure time
		exposureTime, err := fileInfo.GetFloat("ExposureTime")
		if err == nil {
			log.Printf("Exposure time: %f", exposureTime)
			newExif.Exposure = &exposureTime
		}

		// Get aperture
		aperture, err := fileInfo.GetFloat("Aperture")
		if err == nil {
			log.Printf("Aperture: %f", aperture)
			newExif.Aperture = &aperture
		}

		// Get ISO
		iso, err := fileInfo.GetInt("ISO")
		if err == nil {
			log.Printf("ISO: %d", iso)
			newExif.Iso = &iso
		}

		// Get focal length
		focalLen, err := fileInfo.GetString("FocalLength")
		if err == nil {
			log.Printf("Focal length: %s", focalLen)
			reg, _ := regexp.Compile("[0-9.]+")
			focalLenStr := reg.FindString(focalLen)
			focalLenFloat, err := strconv.ParseFloat(focalLenStr, 64)
			if err == nil {
				newExif.FocalLength = &focalLenFloat
			}
		}

		// Get flash info
		flash, err := fileInfo.GetInt("Flash")
		if err == nil {
			log.Printf("Flash: %d", flash)
			newExif.Flash = &flash
		}

		// Get orientation
		orientation, err := fileInfo.GetInt("Orientation")
		if err == nil {
			log.Printf("Orientation: %d", orientation)
			newExif.Orientation = &orientation
		}

		// Get exposure program
		expProgram, err := fileInfo.GetInt("ExposureProgram")
		if err == nil {
			log.Printf("Exposure Program: %d", expProgram)
			newExif.ExposureProgram = &expProgram
		}

		// GPS coordinates - longitude
		longitudeRaw, err := fileInfo.GetFloat("GPSLongitude")
		if err == nil {
			log.Printf("GPS longitude: %f", longitudeRaw)
			newExif.GPSLongitude = &longitudeRaw
		}

		// GPS coordinates - latitude
		latitudeRaw, err := fileInfo.GetFloat("GPSLatitude")
		if err == nil {
			log.Printf("GPS latitude: %f", latitudeRaw)
			newExif.GPSLatitude = &latitudeRaw
		}
	}

	returnExif = &newExif
	return
}
