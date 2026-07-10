package models

import (
	"fmt"
	"time"
)

type MediaEXIF struct {
	Model
	Description     *string
	Camera          *string
	Maker           *string
	Lens            *string
	DateShot        *time.Time
	OffsetSecShot   *int
	Exposure        *float64
	Aperture        *float64
	Iso             *int64
	FocalLength     *float64
	Flash           *int64
	Orientation     *int64
	ExposureProgram *int64
	GPSLatitude     *float64
	GPSLongitude    *float64
}

func (MediaEXIF) TableName() string {
	return "media_exif"
}

func (exif *MediaEXIF) Media() *Media {
	panic("not implemented")
}

func (exif *MediaEXIF) Coordinates() *Coordinates {
	if exif.GPSLatitude == nil || exif.GPSLongitude == nil {
		return nil
	}

	return &Coordinates{
		Latitude:  *exif.GPSLatitude,
		Longitude: *exif.GPSLongitude,
	}
}

const rfc3339WithoutTimezone = "2006-01-02T15:04:05.999"

func (exif *MediaEXIF) DateShotWithOffset() *string {
	if exif.DateShot == nil {
		return nil
	}
	dateShot := exif.DateShot.UTC()
	dateNoTimezone := dateShot.Format(rfc3339WithoutTimezone)

	if exif.OffsetSecShot == nil {
		return &dateNoTimezone
	}

	offsetAbs := *exif.OffsetSecShot
	sign := "+"
	if offsetAbs < 0 {
		offsetAbs = -offsetAbs
		sign = "-"
	}
	hour := offsetAbs / 60 / 60
	minute := offsetAbs / 60 % 60
	date := fmt.Sprintf("%s%s%02d:%02d", dateNoTimezone, sign, hour, minute)

	return &date
}
