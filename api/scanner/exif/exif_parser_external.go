package exif

import (
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/kjeldgaard/go-exiftool"
	"github.com/photoview/photoview/api/graphql/models"
)

type externalExifParser struct{}

func (p *externalExifParser) ParseExif(media *models.Media) (returnExif *models.MediaEXIF, returnErr error) {
	// Init ExifTool

	et2 := exiftool.Charset("-n")

	et, err := 																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																			.NewExiftool(et2)
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
		exposureTime, err := fileInfo.GetString("ExposureTime")
		if err == nil {
			log.Printf("Exposure time: %s", exposureTime)
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
		flash, err := fileInfo.GetString("Flash")
		if err == nil {
			log.Printf("Flash: %s", flash)
			newExif.Flash = &flash
		}

		// Get orientation
		orientation, err := fileInfo.GetInt("Orientation")
		if err == nil {
			log.Printf("Orientation: %d", orientation)
			newExif.Orientation = &orientation
		}

		// Get exposure program
		expProgram, err := fileInfo.GetStrings("ExposureProgram")
		if err == nil {
			for _, value := range expProgram {
				log.Printf("%s", value)
			}
			//log.Printf("Exposure Program: %d", expProgram)
		}

		// GPS coordinates - longitude
		longitudeRaw, err := fileInfo.GetString("GPSLongitude")
		if err == nil {
			log.Printf("GPS longitude: %s", longitudeRaw)
			value, err := ConvertCoodinateToFloat(longitudeRaw)
			if err == nil {
				newExif.GPSLongitude = &value
			}
		}

		// GPS coordinates - latitude
		latitudeRaw, err := fileInfo.GetString("GPSLatitude")
		if err == nil {
			log.Printf("GPS latitude: %s", latitudeRaw)
			value, err := ConvertCoodinateToFloat(latitudeRaw)
			if err == nil {
				newExif.GPSLatitude = &value
			}
		}
	}

	returnExif = &newExif
	return
}

func ConvertCoodinateToFloat(coordinate string) (value float64, err error) {
	reg, err := regexp.Compile("[0-9.]+")
	if err != nil {
		return 0, err
	}

	coordinateStr := reg.FindAllString(coordinate, -1)
	log.Printf("GPS: %s length: %d\n", coordinateStr, len(coordinateStr))
	if len(coordinateStr) != 3 {
		return 0, err
	}

	deg, err := strconv.ParseFloat(coordinateStr[0], 64)
	if err != nil {
		return 0, err
	}

	minute, err := strconv.ParseFloat(coordinateStr[1], 64)
	if err != nil {
		return 0, err
	}

	second, err := strconv.ParseFloat(coordinateStr[2], 64)
	if err != nil {
		return 0, err
	}

	var multiplier float64 = 1
	if deg < 0 {
		multiplier = -1
	}

	value = (deg + minute / 60 + second / 3600) * multiplier
	return value, nil
}
